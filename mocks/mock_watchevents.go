// Code generated by MockGen. DO NOT EDIT.
// Source: watch-informer/api (interfaces: Watcher_WatchEventsServer)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	api "watch-informer/api"

	gomock "github.com/golang/mock/gomock"
	metadata "google.golang.org/grpc/metadata"
)

// MockWatcher_WatchEventsServer is a mock of Watcher_WatchEventsServer interface.
type MockWatcher_WatchEventsServer struct {
	ctrl     *gomock.Controller
	recorder *MockWatcher_WatchEventsServerMockRecorder
}

// MockWatcher_WatchEventsServerMockRecorder is the mock recorder for MockWatcher_WatchEventsServer.
type MockWatcher_WatchEventsServerMockRecorder struct {
	mock *MockWatcher_WatchEventsServer
}

// NewMockWatcher_WatchEventsServer creates a new mock instance.
func NewMockWatcher_WatchEventsServer(ctrl *gomock.Controller) *MockWatcher_WatchEventsServer {
	mock := &MockWatcher_WatchEventsServer{ctrl: ctrl}
	mock.recorder = &MockWatcher_WatchEventsServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWatcher_WatchEventsServer) EXPECT() *MockWatcher_WatchEventsServerMockRecorder {
	return m.recorder
}

// Context mocks base method.
func (m *MockWatcher_WatchEventsServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockWatcher_WatchEventsServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).Context))
}

// Recv mocks base method.
func (m *MockWatcher_WatchEventsServer) Recv() (*api.EventRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*api.EventRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recv indicates an expected call of Recv.
func (mr *MockWatcher_WatchEventsServerMockRecorder) Recv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).Recv))
}

// RecvMsg mocks base method.
func (m *MockWatcher_WatchEventsServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockWatcher_WatchEventsServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockWatcher_WatchEventsServer) Send(arg0 *api.ResourceEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockWatcher_WatchEventsServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockWatcher_WatchEventsServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockWatcher_WatchEventsServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m *MockWatcher_WatchEventsServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockWatcher_WatchEventsServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method.
func (m *MockWatcher_WatchEventsServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockWatcher_WatchEventsServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockWatcher_WatchEventsServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockWatcher_WatchEventsServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockWatcher_WatchEventsServer)(nil).SetTrailer), arg0)
}