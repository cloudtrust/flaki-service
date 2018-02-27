// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/flaki-service/pkg/flaki (interfaces: Component)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// FlakiComponent is a mock of Component interface
type FlakiComponent struct {
	ctrl     *gomock.Controller
	recorder *FlakiComponentMockRecorder
}

// FlakiComponentMockRecorder is the mock recorder for FlakiComponent
type FlakiComponentMockRecorder struct {
	mock *FlakiComponent
}

// NewFlakiComponent creates a new mock instance
func NewFlakiComponent(ctrl *gomock.Controller) *FlakiComponent {
	mock := &FlakiComponent{ctrl: ctrl}
	mock.recorder = &FlakiComponentMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *FlakiComponent) EXPECT() *FlakiComponentMockRecorder {
	return m.recorder
}

// NextID mocks base method
func (m *FlakiComponent) NextID(arg0 context.Context) (string, error) {
	ret := m.ctrl.Call(m, "NextID", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NextID indicates an expected call of NextID
func (mr *FlakiComponentMockRecorder) NextID(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextID", reflect.TypeOf((*FlakiComponent)(nil).NextID), arg0)
}

// NextValidID mocks base method
func (m *FlakiComponent) NextValidID(arg0 context.Context) string {
	ret := m.ctrl.Call(m, "NextValidID", arg0)
	ret0, _ := ret[0].(string)
	return ret0
}

// NextValidID indicates an expected call of NextValidID
func (mr *FlakiComponentMockRecorder) NextValidID(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextValidID", reflect.TypeOf((*FlakiComponent)(nil).NextValidID), arg0)
}