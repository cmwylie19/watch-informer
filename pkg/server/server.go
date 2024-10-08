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
	req = formatRequest(req)
	gvr := schema.GroupVersionResource{
		Group:    req.Group,
		Version:  req.Version,
		Resource: req.Resource,
	}

	sessionId := fmt.Sprintf("%s-%s-%s-%s", req.Group, req.Version, req.Resource, req.Namespace)
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

func StartGRPCServer(address string, dynamicClient dynamic.Interface, logger logging.LoggerInterface) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := NewServer(dynamicClient, logger)
	grpcServer := grpc.NewServer()
	api.RegisterWatchServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	logger.Info(fmt.Sprintf("Server listening at %s", address))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func formatRequest(req *api.WatchRequest) *api.WatchRequest {
	req.Resource = strings.ToLower(req.Resource)
	if !strings.HasSuffix(req.Resource, "s") {
		req.Resource = fmt.Sprint(req.Resource, "s")
	}

	return req
}
