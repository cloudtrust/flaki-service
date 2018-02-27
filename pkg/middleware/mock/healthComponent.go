// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/flaki-service/pkg/health (interfaces: Component)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	health "github.com/cloudtrust/flaki-service/pkg/health"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// HealthComponent is a mock of Component interface
type HealthComponent struct {
	ctrl     *gomock.Controller
	recorder *HealthComponentMockRecorder
}

// HealthComponentMockRecorder is the mock recorder for HealthComponent
type HealthComponentMockRecorder struct {
	mock *HealthComponent
}

// NewHealthComponent creates a new mock instance
func NewHealthComponent(ctrl *gomock.Controller) *HealthComponent {
	mock := &HealthComponent{ctrl: ctrl}
	mock.recorder = &HealthComponentMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *HealthComponent) EXPECT() *HealthComponentMockRecorder {
	return m.recorder
}

// InfluxHealthChecks mocks base method
func (m *HealthComponent) InfluxHealthChecks(arg0 context.Context) health.HealthReports {
	ret := m.ctrl.Call(m, "InfluxHealthChecks", arg0)
	ret0, _ := ret[0].(health.HealthReports)
	return ret0
}

// InfluxHealthChecks indicates an expected call of InfluxHealthChecks
func (mr *HealthComponentMockRecorder) InfluxHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfluxHealthChecks", reflect.TypeOf((*HealthComponent)(nil).InfluxHealthChecks), arg0)
}

// JaegerHealthChecks mocks base method
func (m *HealthComponent) JaegerHealthChecks(arg0 context.Context) health.HealthReports {
	ret := m.ctrl.Call(m, "JaegerHealthChecks", arg0)
	ret0, _ := ret[0].(health.HealthReports)
	return ret0
}

// JaegerHealthChecks indicates an expected call of JaegerHealthChecks
func (mr *HealthComponentMockRecorder) JaegerHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JaegerHealthChecks", reflect.TypeOf((*HealthComponent)(nil).JaegerHealthChecks), arg0)
}

// RedisHealthChecks mocks base method
func (m *HealthComponent) RedisHealthChecks(arg0 context.Context) health.HealthReports {
	ret := m.ctrl.Call(m, "RedisHealthChecks", arg0)
	ret0, _ := ret[0].(health.HealthReports)
	return ret0
}

// RedisHealthChecks indicates an expected call of RedisHealthChecks
func (mr *HealthComponentMockRecorder) RedisHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RedisHealthChecks", reflect.TypeOf((*HealthComponent)(nil).RedisHealthChecks), arg0)
}

// SentryHealthChecks mocks base method
func (m *HealthComponent) SentryHealthChecks(arg0 context.Context) health.HealthReports {
	ret := m.ctrl.Call(m, "SentryHealthChecks", arg0)
	ret0, _ := ret[0].(health.HealthReports)
	return ret0
}

// SentryHealthChecks indicates an expected call of SentryHealthChecks
func (mr *HealthComponentMockRecorder) SentryHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SentryHealthChecks", reflect.TypeOf((*HealthComponent)(nil).SentryHealthChecks), arg0)
}
