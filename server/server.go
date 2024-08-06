package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"watch-informer/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type server struct {
	api.UnimplementedWatcherServer
	dynamicClient dynamic.Interface
	eventChans    map[string]chan *api.ResourceEvent
	mu            sync.Mutex
}

func NewServer(dynamicClient dynamic.Interface) *server {
	return &server{
		dynamicClient: dynamicClient,
		eventChans:    make(map[string]chan *api.ResourceEvent),
	}
}

func (s *server) StartWatch(ctx context.Context, req *api.WatchRequest) (*api.WatchResponse, error) {
	gvr := schema.GroupVersionResource{
		Group:    req.Group,
		Version:  req.Version,
		Resource: req.Resource,
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.dynamicClient, 0, req.Namespace, nil)
	informer := factory.ForResource(gvr).Informer()

	eventChan := make(chan *api.ResourceEvent, 100)
	sessionId := fmt.Sprintf("%s-%s-%s", req.Group, req.Version, req.Resource)
	s.mu.Lock()
	s.eventChans[sessionId] = eventChan
	s.mu.Unlock()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			eventChan <- &api.ResourceEvent{EventType: "ADD", Details: fmt.Sprintf("Added: %v", obj)}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			eventChan <- &api.ResourceEvent{EventType: "UPDATE", Details: fmt.Sprintf("Updated from %v to %v", oldObj, newObj)}
		},
		DeleteFunc: func(obj interface{}) {
			eventChan <- &api.ResourceEvent{EventType: "DELETE", Details: fmt.Sprintf("Deleted: %v", obj)}
		},
	})

	go informer.Run(make(chan struct{}))

	return &api.WatchResponse{Message: "Informer started successfully"}, nil
}

func (s *server) WatchEvents(stream api.Watcher_WatchEventsServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		s.mu.Lock()
		eventChan, exists := s.eventChans[req.SessionId]
		s.mu.Unlock()

		if !exists {
			return fmt.Errorf("invalid session ID")
		}

		for event := range eventChan {
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

func StartGRPCServer(address string, dynamicClient dynamic.Interface, group, version, resource, namespace string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := NewServer(dynamicClient)
	grpcServer := grpc.NewServer()
	api.RegisterWatcherServer(grpcServer, s)
	reflection.Register(grpcServer)

	// Optionally start watching a resource if configured via CLI
	if group != "" && version != "" && resource != "" {
		_, err := s.StartWatch(context.Background(), &api.WatchRequest{
			Group:     group,
			Version:   version,
			Resource:  resource,
			Namespace: namespace,
		})
		if err != nil {
			log.Fatalf("Failed to start initial watch: %v", err)
		}
	}

	log.Println("Server listening at", address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
