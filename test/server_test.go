package server_test

import (
	"context"
	"testing"

	"watch-informer/api"
	"watch-informer/server"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

func TestStartWatch(t *testing.T) {
	fakeClient := fake.NewSimpleDynamicClient(runtime.NewScheme())
	s := server.NewServer(fakeClient)
	req := &api.WatchRequest{
		Group:     "core",
		Version:   "v1",
		Resource:  "pods",
		Namespace: "default",
	}
	_, err := s.StartWatch(context.Background(), req)
	if err != nil {
		t.Errorf("Failed to start watch: %v", err)
	}
}
