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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type server struct {
	api.UnimplementedWatchServiceServer
	dynamicClient dynamic.Interface
	Logger        logging.LoggerInterface
	eventChans    map[string]chan *api.WatchResponse
	mu            sync.Mutex
}

func NewServer(dynamicClient dynamic.Interface, logger logging.LoggerInterface) *server {
	return &server{
		dynamicClient: dynamicClient,
		eventChans:    make(map[string]chan *api.WatchResponse),
		Logger:        logger,
	}
}
func toJson(obj interface{}) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("Error converting to JSON: %v", err)
	}
	return fmt.Sprintf("%v", string(jsonData))
}
func (s *server) Watch(req *api.WatchRequest, srv api.WatchService_WatchServer) error {
	req.Resource = fmt.Sprint(strings.ToLower(req.Resource), "s")
	req.Group = strings.TrimPrefix(req.Group, "/")
	gvr := schema.GroupVersionResource{
		Group:    req.Group,
		Version:  req.Version,
		Resource: req.Resource,
	}

	// Check if namespace is provided, else watch all namespaces
	// if req.Namespace == "" {
	// 	req.Namespace = "all"
	// }
	sessionId := fmt.Sprintf("%s-%s-%s-%s", req.Group, req.Version, req.Resource, req.Namespace)
	s.Logger.Info(fmt.Sprintf("Starting watch for %s", sessionId))
	s.Logger.Debug(fmt.Sprintf("GVR: %v", gvr))
	// Initialize informer factory with a check for the dynamic client
	if s.dynamicClient == nil {
		return fmt.Errorf("dynamic client is not initialized")
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.dynamicClient, 5*time.Minute, req.Namespace, nil)
	informer := factory.ForResource(gvr).Informer()

	// Create and register an event channel
	eventChan := make(chan *api.WatchResponse, 100)
	s.mu.Lock()
	s.eventChans[sessionId] = eventChan
	s.mu.Unlock()

	// Set up event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			s.Logger.Debug(fmt.Sprintf("EventType: ADD, Details: %v", obj))
			select {
			case eventChan <- &api.WatchResponse{EventType: "ADD", Details: fmt.Sprintf("%v", toJson(obj))}:
			default:
				s.Logger.Error("Event channel is full, dropping event")
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			s.Logger.Debug(fmt.Sprintf("EventType: UPDATE, Details: %v", newObj))
			select {
			case eventChan <- &api.WatchResponse{EventType: "UPDATE", Details: fmt.Sprintf("%v", toJson(newObj))}:
			default:
				s.Logger.Error("Event channel is full, dropping event")
			}
		},
		DeleteFunc: func(obj interface{}) {
			s.Logger.Debug(fmt.Sprintf("EventType: DELETE, Details: %v", obj))
			select {
			case eventChan <- &api.WatchResponse{EventType: "DELETE", Details: fmt.Sprintf("%v", toJson(obj))}:
			default:
				s.Logger.Error("Event channel is full, dropping event")
			}
		},
	})

	// Start the informer and handle any potential panics
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
	// Wait for the channel to close or the context to be canceled
	<-srv.Context().Done()
	return srv.Context().Err()
}

func StartGRPCServer(address string, dynamicClient dynamic.Interface, logger logging.LoggerInterface) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := NewServer(dynamicClient, logger)
	grpcServer := grpc.NewServer()
	api.RegisterWatchServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.Logger.Info(fmt.Sprintf("Server listening at %s", address))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
