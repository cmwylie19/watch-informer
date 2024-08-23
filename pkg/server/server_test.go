package server

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/client-go/dynamic"

	"github.com/cmwylie19/watch-informer/api"
	"github.com/cmwylie19/watch-informer/mocks"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	var mockClient dynamic.Interface
	s := NewServer(mockClient, mockLogger)

	if s == nil {
		t.Errorf("Expected server object, got nil")
	}
}

func TestFormatRequest(t *testing.T) {
	tests := []struct {
		name     string
		inputReq *api.WatchRequest
		expected *api.WatchRequest
	}{
		{
			name:     "Lowercase resource with missing plural",
			inputReq: &api.WatchRequest{Resource: "Pod", Group: "v1"},
			expected: &api.WatchRequest{Resource: "pods", Group: "v1"},
		},
		{
			name:     "Uppercase resource with missing plural",
			inputReq: &api.WatchRequest{Resource: "POD", Group: "v1"},
			expected: &api.WatchRequest{Resource: "pods", Group: "v1"},
		},
		{
			name:     "Resource already plural",
			inputReq: &api.WatchRequest{Resource: "Deployments"},
			expected: &api.WatchRequest{Resource: "deployments", Group: ""},
		},
		{
			name:     "Group with leading slash",
			inputReq: &api.WatchRequest{Resource: "Services", Group: "v1"},
			expected: &api.WatchRequest{Resource: "services", Group: "v1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := formatRequest(tc.inputReq)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected: %#v, got: %#v", tc.expected, actual)
			}
		})
	}
}
