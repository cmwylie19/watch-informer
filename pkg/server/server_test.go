package server

import (
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/client-go/dynamic"

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
