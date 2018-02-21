// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-kit/kit/transport/grpc (interfaces: Handler)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// Handler is a mock of Handler interface
type Handler struct {
	ctrl     *gomock.Controller
	recorder *HandlerMockRecorder
}

// HandlerMockRecorder is the mock recorder for Handler
type HandlerMockRecorder struct {
	mock *Handler
}

// NewHandler creates a new mock instance
func NewHandler(ctrl *gomock.Controller) *Handler {
	mock := &Handler{ctrl: ctrl}
	mock.recorder = &HandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Handler) EXPECT() *HandlerMockRecorder {
	return m.recorder
}

// ServeGRPC mocks base method
func (m *Handler) ServeGRPC(arg0 context.Context, arg1 interface{}) (context.Context, interface{}, error) {
	ret := m.ctrl.Call(m, "ServeGRPC", arg0, arg1)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(interface{})
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ServeGRPC indicates an expected call of ServeGRPC
func (mr *HandlerMockRecorder) ServeGRPC(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServeGRPC", reflect.TypeOf((*Handler)(nil).ServeGRPC), arg0, arg1)
}
