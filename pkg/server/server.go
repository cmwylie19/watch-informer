package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cmwylie19/watch-informer/api"
	"github.com/cmwylie19/watch-informer/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type server struct {
	api.UnimplementedWatchServiceServer
	dynamicClient   dynamic.Interface
	config          *rest.Config
	Logger          logging.LoggerInterface
	getResourceName func(*rest.Config, string, string, string) (string, error)

	mu   sync.Mutex
	hubs map[watchKey]*watchHub
}

func NewServer(dynamicClient dynamic.Interface, restConfig *rest.Config, logger logging.LoggerInterface) *server {
	return &server{
		dynamicClient:   dynamicClient,
		Logger:          logger,
		config:          restConfig,
		getResourceName: getResourceName,
		hubs:            make(map[watchKey]*watchHub),
	}
}

type watchKey struct {
	Group     string
	Version   string
	Resource  string
	Namespace string

	FieldSelector string
	LabelSelector string

	ResyncSeconds int32

	// NOTE: startResourceVersion is NOT part of the key on purpose.
	// This service uses one shared informer per (GVR+selectors+namespace+resync).
	// startResourceVersion can only be applied when creating a NEW informer.
}

type subscriber struct {
	ch chan *api.WatchResponse
}

type watchHub struct {
	key      watchKey
	informer cache.SharedIndexInformer
	stopCh   chan struct{}

	mu   sync.Mutex
	subs map[*subscriber]struct{}

	logger logging.LoggerInterface
}

func (h *watchHub) addSubscriber(sub *subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subs[sub] = struct{}{}
}

func (h *watchHub) removeSubscriber(sub *subscriber) (remaining int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs, sub)
	return len(h.subs)
}

func (h *watchHub) broadcast(evt *api.WatchResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for sub := range h.subs {
		select {
		case sub.ch <- evt:
		default:
			// Backpressure policy: drop for that subscriber.
			// Optional alternative: disconnect slow subscribers with RESOURCE_EXHAUSTED.
			h.logger.Error("Subscriber channel full, dropping event")
		}
	}
}

func (h *watchHub) close() {
	select {
	case <-h.stopCh:
	default:
		close(h.stopCh)
	}
}

// ---------- gRPC Watch ----------

func (s *server) Watch(req *api.WatchRequest, srv api.WatchService_WatchServer) error {
	if s.dynamicClient == nil {
		return fmt.Errorf("dynamic client is not initialized")
	}

	// Normalize + pluralize resource
	req, err := s.formatRequest(req)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to format request: %v", err))
		return err
	}

	key := watchKey{
		Group:         req.Group,
		Version:       req.Version,
		Resource:      req.Resource,
		Namespace:     req.Namespace,
		FieldSelector: req.FieldSelector,
		LabelSelector: req.LabelSelector,
		ResyncSeconds: req.ResyncPeriodSeconds,
	}

	sessionID := formatSessionID(req)
	s.Logger.Info(fmt.Sprintf("Watch subscribe: %s", sessionID))

	// NOTE: startResourceVersion is only applied when creating a new hub/informer.
	// If a hub already exists, the subscriber is attached to the current shared stream.
	hub, created, err := s.getOrCreateHub(key, req.StartResourceVersion)
	if err != nil {
		return err
	}
	if !created && req.StartResourceVersion != "" {
		s.Logger.Debug("Ignoring startResourceVersion because hub already exists; use sendInitialList=true to resync")
	}

	// Each stream gets its own buffered channel
	sub := &subscriber{ch: make(chan *api.WatchResponse, 256)}
	hub.addSubscriber(sub)

	// Optional initial list to THIS subscriber only, then SNAPSHOT_END.
	if req.SendInitialList {
		if err := s.sendInitialSnapshot(hub, sub); err != nil {
			s.Logger.Error(fmt.Sprintf("Failed to send initial snapshot: %v", err))
			remaining := hub.removeSubscriber(sub)
			close(sub.ch)
			s.maybeStopHub(key, hub, remaining)
			return err
		}
	}

	// Stream sender loop
	sendErrCh := make(chan error, 1)
	go func() {
		for {
			select {
			case evt, ok := <-sub.ch:
				if !ok {
					sendErrCh <- nil
					return
				}
				if err := srv.Send(evt); err != nil {
					sendErrCh <- err
					return
				}
			case <-srv.Context().Done():
				sendErrCh <- srv.Context().Err()
				return
			}
		}
	}()

	// Wait for completion
	err = <-sendErrCh

	remaining := hub.removeSubscriber(sub)
	close(sub.ch)
	s.maybeStopHub(key, hub, remaining)

	return err
}

func (s *server) getOrCreateHub(key watchKey, startResourceVersion string) (*watchHub, bool, error) {
	fmt.Println("getOrCreateHub called with key:", key)
	s.mu.Lock()
	defer s.mu.Unlock()

	if hub, ok := s.hubs[key]; ok {
		return hub, false, nil
	}

	resync := time.Duration(0)
	if key.ResyncSeconds > 0 {
		resync = time.Duration(key.ResyncSeconds) * time.Second
	}

	gvr := schema.GroupVersionResource{
		Group:    key.Group,
		Version:  key.Version,
		Resource: key.Resource,
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		s.dynamicClient,
		resync,
		key.Namespace,
		func(listOptions *metav1.ListOptions) {
			if key.FieldSelector != "" {
				listOptions.FieldSelector = key.FieldSelector
			}
			if key.LabelSelector != "" {
				listOptions.LabelSelector = key.LabelSelector
			}
			if startResourceVersion != "" {
				listOptions.ResourceVersion = startResourceVersion
			}
		},
	)

	informer := factory.ForResource(gvr).Informer()

	hub := &watchHub{
		key:      key,
		informer: informer,
		stopCh:   make(chan struct{}),
		subs:     make(map[*subscriber]struct{}),
		logger:   s.Logger,
	}

	// Event handlers -> broadcast
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			obj, rv := unwrapAndRV(obj)
			b, err := toJSONBytes(obj)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("Failed to marshal ADDED object: %v", err))
				return
			}
			hub.broadcast(&api.WatchResponse{
				EventType:       api.EventType_ADDED,
				Details:         b,
				ResourceVersion: rv,
			})
		},
		UpdateFunc: func(_, newObj interface{}) {
			newObj, rv := unwrapAndRV(newObj)
			b, err := toJSONBytes(newObj)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("Failed to marshal MODIFIED object: %v", err))
				return
			}
			hub.broadcast(&api.WatchResponse{
				EventType:       api.EventType_MODIFIED,
				Details:         b,
				ResourceVersion: rv,
			})
		},
		DeleteFunc: func(obj interface{}) {
			obj, rv := unwrapAndRV(obj)
			b, err := toJSONBytes(obj)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("Failed to marshal DELETED object: %v", err))
				return
			}
			hub.broadcast(&api.WatchResponse{
				EventType:       api.EventType_DELETED,
				Details:         b,
				ResourceVersion: rv,
			})
		},
	})

	// Start informer for this hub
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.Logger.Error(fmt.Sprint("Recovered in informer.Run: ", r))
			}
		}()
		informer.Run(hub.stopCh)
	}()

	// Cache sync health log
	go func() {
		if !cache.WaitForCacheSync(hub.stopCh, informer.HasSynced) {
			s.Logger.Error("Informer cache failed to sync")
		}
	}()

	s.hubs[key] = hub
	s.Logger.Info(fmt.Sprintf("Created hub: %s", formatWatchKey(key)))

	return hub, true, nil
}

func (s *server) maybeStopHub(key watchKey, hub *watchHub, remaining int) {
	if remaining > 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cur, ok := s.hubs[key]
	if !ok || cur != hub {
		return
	}

	delete(s.hubs, key)
	hub.close()
	s.Logger.Info(fmt.Sprintf("Stopped hub: %s", formatWatchKey(key)))
}

// ---------- Initial snapshot per subscriber ----------

func (s *server) sendInitialSnapshot(hub *watchHub, sub *subscriber) error {
	if !cache.WaitForCacheSync(hub.stopCh, hub.informer.HasSynced) {
		return fmt.Errorf("cache not synced")
	}

	items := hub.informer.GetStore().List()
	for _, item := range items {
		item, rv := unwrapAndRV(item)
		b, err := toJSONBytes(item)
		if err != nil {
			return fmt.Errorf("failed to marshal snapshot item: %w", err)
		}
		select {
		case sub.ch <- &api.WatchResponse{
			EventType:       api.EventType_ADDED,
			Details:         b,
			ResourceVersion: rv,
		}:
		default:
			return fmt.Errorf("subscriber buffer full while sending snapshot")
		}
	}

	// Marker event with the informer’s last synced RV (useful for client resume point)
	endRV := hub.informer.LastSyncResourceVersion()
	select {
	case sub.ch <- &api.WatchResponse{
		EventType:       api.EventType_SNAPSHOT_END,
		Details:         nil,
		ResourceVersion: endRV,
	}:
	default:
		return fmt.Errorf("subscriber buffer full while sending SNAPSHOT_END")
	}

	return nil
}

// ---------- Utilities ----------

func unwrapAndRV(obj interface{}) (interface{}, string) {
	// Handle delete tombstone
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = tombstone.Obj
	}

	if u, ok := obj.(*unstructured.Unstructured); ok && u != nil {
		return obj, u.GetResourceVersion()
	}

	// Dynamic informer should normally yield *unstructured.Unstructured,
	// but return empty RV if not.
	return obj, ""
}

func toJSONBytes(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func formatWatchKey(k watchKey) string {
	ns := k.Namespace
	if ns == "" {
		ns = "*"
	}
	g := k.Group
	if g == "" {
		g = "''"
	}
	return fmt.Sprintf("Group=%s Version=%s Resource=%s Namespace=%s FieldSelector=%q LabelSelector=%q Resync=%ds",
		g, k.Version, k.Resource, ns, k.FieldSelector, k.LabelSelector, k.ResyncSeconds,
	)
}

func StartGRPCServer(address string, dynamicClient dynamic.Interface, restConfig *rest.Config, logger logging.LoggerInterface) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := NewServer(dynamicClient, restConfig, logger)
	grpcServer := grpc.NewServer()
	api.RegisterWatchServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	logger.Info(fmt.Sprintf("Server listening at %s", address))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// formatRequest formats the request to lowercase and fetches the correct plural name
func (s *server) formatRequest(req *api.WatchRequest) (*api.WatchRequest, error) {
	req.Resource = strings.ToLower(req.Resource)

	resourceName, err := s.getResourceName(s.config, req.Group, req.Version, req.Resource)
	if err != nil {
		return nil, err
	}

	req.Resource = resourceName
	return req, nil
}

func formatSessionID(req *api.WatchRequest) string {
	var namespace, group string
	if req.Namespace == "" {
		namespace = "*"
	} else {
		namespace = req.Namespace
	}
	if req.Group == "" {
		group = "''"
	} else {
		group = req.Group
	}

	return fmt.Sprintf("Group: %s, Version: %s, Resource: %s, Namespace: %s", group, req.Version, req.Resource, namespace)
}

func getResourceName(restConfig *rest.Config, group, version, resource string) (string, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create discovery client: %w", err)
	}
	formattedGV := getFormattedGV(group, version)
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion(formattedGV)
	if err != nil {
		return "", fmt.Errorf("failed to fetch resource list: %w", err)
	}

	for _, apiResource := range resourceList.APIResources {
		if apiResource.SingularName == resource || apiResource.Name == resource {
			return apiResource.Name, nil
		}
	}

	return "", fmt.Errorf("resource not found: %s", resource)
}

func getFormattedGV(group, version string) string {
	if group == "" {
		return version
	}
	return fmt.Sprintf("%s/%s", group, version)
}
