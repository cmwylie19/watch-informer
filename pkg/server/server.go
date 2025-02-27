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
	eventChans      map[string]chan *api.WatchResponse
	mu              sync.Mutex
	getResourceName func(*rest.Config, string, string, string) (string, error)
}

// NewServer creates a new server instance
func NewServer(dynamicClient dynamic.Interface, restConfig *rest.Config, logger logging.LoggerInterface) *server {
	return &server{
		dynamicClient:   dynamicClient,
		eventChans:      make(map[string]chan *api.WatchResponse),
		Logger:          logger,
		config:          restConfig,
		getResourceName: getResourceName,
	}
}

// toJson converts an object to a JSON string
func toJson(obj interface{}) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("Error converting to JSON: %v", err)
	}
	return fmt.Sprintf("%v", string(jsonData))
}

// Watch implements the WatchServiceServer interface
func (s *server) Watch(req *api.WatchRequest, srv api.WatchService_WatchServer) error {
	req, err := s.formatRequest(req)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to format request: %v", err))
	}
	gvr := schema.GroupVersionResource{
		Group:    req.Group,
		Version:  req.Version,
		Resource: req.Resource,
	}
	sessionId := formatSessionID(req)

	s.Logger.Info(fmt.Sprintf("Starting watch for %s", sessionId))
	s.Logger.Debug(fmt.Sprintf("GVR: %v", gvr))

	if s.dynamicClient == nil {
		return fmt.Errorf("dynamic client is not initialized")
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.dynamicClient, 5*time.Minute, req.Namespace, nil)
	informer := factory.ForResource(gvr).Informer()

	eventChan := make(chan *api.WatchResponse, 100)
	s.mu.Lock()
	s.eventChans[sessionId] = eventChan
	s.mu.Unlock()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			s.Logger.Debug(fmt.Sprintf("EventType: ADD, Details: %v", toJson(obj)))
			select {
			case eventChan <- &api.WatchResponse{EventType: "ADD", Details: toJson(obj)}:
			default:
				s.Logger.Error("Event channel is full, dropping event")
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			s.Logger.Debug(fmt.Sprintf("EventType: UPDATE, Details: %v", toJson(newObj)))
			select {
			case eventChan <- &api.WatchResponse{EventType: "UPDATE", Details: toJson(newObj)}:
			default:
				s.Logger.Error("Event channel is full, dropping event")
			}
		},
		DeleteFunc: func(obj interface{}) {
			s.Logger.Debug(fmt.Sprintf("EventType: DELETE, Details: %v", toJson(obj)))
			select {
			case eventChan <- &api.WatchResponse{EventType: "DELETE", Details: toJson(obj)}:
			default:
				s.Logger.Error("Event channel is full, dropping event")
			}
		},
	})

	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error(fmt.Sprint("Recovered in StartWatch", r))
		}
	}()
	go informer.Run(make(chan struct{}))

	go func() {
		for event := range eventChan {
			if err := srv.Send(event); err != nil {
				s.Logger.Error(fmt.Sprint("Failed to send event: ", err))
				return
			}
		}
	}()
	<-srv.Context().Done()
	return srv.Context().Err()
}

// StartGRPCServer starts the gRPC server
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

	// Fetch the correct plural name dynamically
	resourceName, err := s.getResourceName(s.config, req.Group, req.Version, req.Resource)
	if err != nil {
		return nil, err
	}

	req.Resource = resourceName
	return req, nil
}

// formatSessionID formats the session ID
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

// getResourceName fetches the correct resource name
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

// getFormattedGV formats the group and version
func getFormattedGV(group, version string) string {
	if group == "" {
		return version
	}
	return fmt.Sprintf("%s/%s", group, version)
}
