// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/flaki-service/pkg/health (interfaces: Sentry,SentryModule,Redis,RedisModule,Influx,InfluxModule,Jaeger,JaegerModule,Component)

// Package health is a generated GoMock package.
package health

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockSentry is a mock of Sentry interface
type MockSentry struct {
	ctrl     *gomock.Controller
	recorder *MockSentryMockRecorder
}

// MockSentryMockRecorder is the mock recorder for MockSentry
type MockSentryMockRecorder struct {
	mock *MockSentry
}

// NewMockSentry creates a new mock instance
func NewMockSentry(ctrl *gomock.Controller) *MockSentry {
	mock := &MockSentry{ctrl: ctrl}
	mock.recorder = &MockSentryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSentry) EXPECT() *MockSentryMockRecorder {
	return m.recorder
}

// URL mocks base method
func (m *MockSentry) URL() string {
	ret := m.ctrl.Call(m, "URL")
	ret0, _ := ret[0].(string)
	return ret0
}

// URL indicates an expected call of URL
func (mr *MockSentryMockRecorder) URL() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "URL", reflect.TypeOf((*MockSentry)(nil).URL))
}

// MockSentryModule is a mock of SentryModule interface
type MockSentryModule struct {
	ctrl     *gomock.Controller
	recorder *MockSentryModuleMockRecorder
}

// MockSentryModuleMockRecorder is the mock recorder for MockSentryModule
type MockSentryModuleMockRecorder struct {
	mock *MockSentryModule
}

// NewMockSentryModule creates a new mock instance
func NewMockSentryModule(ctrl *gomock.Controller) *MockSentryModule {
	mock := &MockSentryModule{ctrl: ctrl}
	mock.recorder = &MockSentryModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSentryModule) EXPECT() *MockSentryModuleMockRecorder {
	return m.recorder
}

// HealthChecks mocks base method
func (m *MockSentryModule) HealthChecks(arg0 context.Context) []sentryHealthReport {
	ret := m.ctrl.Call(m, "HealthChecks", arg0)
	ret0, _ := ret[0].([]sentryHealthReport)
	return ret0
}

// HealthChecks indicates an expected call of HealthChecks
func (mr *MockSentryModuleMockRecorder) HealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthChecks", reflect.TypeOf((*MockSentryModule)(nil).HealthChecks), arg0)
}

// MockRedis is a mock of Redis interface
type MockRedis struct {
	ctrl     *gomock.Controller
	recorder *MockRedisMockRecorder
}

// MockRedisMockRecorder is the mock recorder for MockRedis
type MockRedisMockRecorder struct {
	mock *MockRedis
}

// NewMockRedis creates a new mock instance
func NewMockRedis(ctrl *gomock.Controller) *MockRedis {
	mock := &MockRedis{ctrl: ctrl}
	mock.recorder = &MockRedisMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRedis) EXPECT() *MockRedisMockRecorder {
	return m.recorder
}

// Do mocks base method
func (m *MockRedis) Do(arg0 string, arg1 ...interface{}) (interface{}, error) {
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Do", varargs...)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do
func (mr *MockRedisMockRecorder) Do(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockRedis)(nil).Do), varargs...)
}

// MockRedisModule is a mock of RedisModule interface
type MockRedisModule struct {
	ctrl     *gomock.Controller
	recorder *MockRedisModuleMockRecorder
}

// MockRedisModuleMockRecorder is the mock recorder for MockRedisModule
type MockRedisModuleMockRecorder struct {
	mock *MockRedisModule
}

// NewMockRedisModule creates a new mock instance
func NewMockRedisModule(ctrl *gomock.Controller) *MockRedisModule {
	mock := &MockRedisModule{ctrl: ctrl}
	mock.recorder = &MockRedisModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRedisModule) EXPECT() *MockRedisModuleMockRecorder {
	return m.recorder
}

// HealthChecks mocks base method
func (m *MockRedisModule) HealthChecks(arg0 context.Context) []redisHealthReport {
	ret := m.ctrl.Call(m, "HealthChecks", arg0)
	ret0, _ := ret[0].([]redisHealthReport)
	return ret0
}

// HealthChecks indicates an expected call of HealthChecks
func (mr *MockRedisModuleMockRecorder) HealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthChecks", reflect.TypeOf((*MockRedisModule)(nil).HealthChecks), arg0)
}

// MockInflux is a mock of Influx interface
type MockInflux struct {
	ctrl     *gomock.Controller
	recorder *MockInfluxMockRecorder
}

// MockInfluxMockRecorder is the mock recorder for MockInflux
type MockInfluxMockRecorder struct {
	mock *MockInflux
}

// NewMockInflux creates a new mock instance
func NewMockInflux(ctrl *gomock.Controller) *MockInflux {
	mock := &MockInflux{ctrl: ctrl}
	mock.recorder = &MockInfluxMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInflux) EXPECT() *MockInfluxMockRecorder {
	return m.recorder
}

// Ping mocks base method
func (m *MockInflux) Ping(arg0 time.Duration) (time.Duration, string, error) {
	ret := m.ctrl.Call(m, "Ping", arg0)
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Ping indicates an expected call of Ping
func (mr *MockInfluxMockRecorder) Ping(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockInflux)(nil).Ping), arg0)
}

// MockInfluxModule is a mock of InfluxModule interface
type MockInfluxModule struct {
	ctrl     *gomock.Controller
	recorder *MockInfluxModuleMockRecorder
}

// MockInfluxModuleMockRecorder is the mock recorder for MockInfluxModule
type MockInfluxModuleMockRecorder struct {
	mock *MockInfluxModule
}

// NewMockInfluxModule creates a new mock instance
func NewMockInfluxModule(ctrl *gomock.Controller) *MockInfluxModule {
	mock := &MockInfluxModule{ctrl: ctrl}
	mock.recorder = &MockInfluxModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInfluxModule) EXPECT() *MockInfluxModuleMockRecorder {
	return m.recorder
}

// HealthChecks mocks base method
func (m *MockInfluxModule) HealthChecks(arg0 context.Context) []influxHealthReport {
	ret := m.ctrl.Call(m, "HealthChecks", arg0)
	ret0, _ := ret[0].([]influxHealthReport)
	return ret0
}

// HealthChecks indicates an expected call of HealthChecks
func (mr *MockInfluxModuleMockRecorder) HealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthChecks", reflect.TypeOf((*MockInfluxModule)(nil).HealthChecks), arg0)
}

// MockJaeger is a mock of Jaeger interface
type MockJaeger struct {
	ctrl     *gomock.Controller
	recorder *MockJaegerMockRecorder
}

// MockJaegerMockRecorder is the mock recorder for MockJaeger
type MockJaegerMockRecorder struct {
	mock *MockJaeger
}

// NewMockJaeger creates a new mock instance
func NewMockJaeger(ctrl *gomock.Controller) *MockJaeger {
	mock := &MockJaeger{ctrl: ctrl}
	mock.recorder = &MockJaegerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJaeger) EXPECT() *MockJaegerMockRecorder {
	return m.recorder
}

// MockJaegerModule is a mock of JaegerModule interface
type MockJaegerModule struct {
	ctrl     *gomock.Controller
	recorder *MockJaegerModuleMockRecorder
}

// MockJaegerModuleMockRecorder is the mock recorder for MockJaegerModule
type MockJaegerModuleMockRecorder struct {
	mock *MockJaegerModule
}

// NewMockJaegerModule creates a new mock instance
func NewMockJaegerModule(ctrl *gomock.Controller) *MockJaegerModule {
	mock := &MockJaegerModule{ctrl: ctrl}
	mock.recorder = &MockJaegerModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJaegerModule) EXPECT() *MockJaegerModuleMockRecorder {
	return m.recorder
}

// HealthChecks mocks base method
func (m *MockJaegerModule) HealthChecks(arg0 context.Context) []jaegerHealthReport {
	ret := m.ctrl.Call(m, "HealthChecks", arg0)
	ret0, _ := ret[0].([]jaegerHealthReport)
	return ret0
}

// HealthChecks indicates an expected call of HealthChecks
func (mr *MockJaegerModuleMockRecorder) HealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthChecks", reflect.TypeOf((*MockJaegerModule)(nil).HealthChecks), arg0)
}

// MockComponent is a mock of Component interface
type MockComponent struct {
	ctrl     *gomock.Controller
	recorder *MockComponentMockRecorder
}

// MockComponentMockRecorder is the mock recorder for MockComponent
type MockComponentMockRecorder struct {
	mock *MockComponent
}

// NewMockComponent creates a new mock instance
func NewMockComponent(ctrl *gomock.Controller) *MockComponent {
	mock := &MockComponent{ctrl: ctrl}
	mock.recorder = &MockComponentMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockComponent) EXPECT() *MockComponentMockRecorder {
	return m.recorder
}

// InfluxHealthChecks mocks base method
func (m *MockComponent) InfluxHealthChecks(arg0 context.Context) HealthReports {
	ret := m.ctrl.Call(m, "InfluxHealthChecks", arg0)
	ret0, _ := ret[0].(HealthReports)
	return ret0
}

// InfluxHealthChecks indicates an expected call of InfluxHealthChecks
func (mr *MockComponentMockRecorder) InfluxHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfluxHealthChecks", reflect.TypeOf((*MockComponent)(nil).InfluxHealthChecks), arg0)
}

// JaegerHealthChecks mocks base method
func (m *MockComponent) JaegerHealthChecks(arg0 context.Context) HealthReports {
	ret := m.ctrl.Call(m, "JaegerHealthChecks", arg0)
	ret0, _ := ret[0].(HealthReports)
	return ret0
}

// JaegerHealthChecks indicates an expected call of JaegerHealthChecks
func (mr *MockComponentMockRecorder) JaegerHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JaegerHealthChecks", reflect.TypeOf((*MockComponent)(nil).JaegerHealthChecks), arg0)
}

// RedisHealthChecks mocks base method
func (m *MockComponent) RedisHealthChecks(arg0 context.Context) HealthReports {
	ret := m.ctrl.Call(m, "RedisHealthChecks", arg0)
	ret0, _ := ret[0].(HealthReports)
	return ret0
}

// RedisHealthChecks indicates an expected call of RedisHealthChecks
func (mr *MockComponentMockRecorder) RedisHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RedisHealthChecks", reflect.TypeOf((*MockComponent)(nil).RedisHealthChecks), arg0)
}

// SentryHealthChecks mocks base method
func (m *MockComponent) SentryHealthChecks(arg0 context.Context) HealthReports {
	ret := m.ctrl.Call(m, "SentryHealthChecks", arg0)
	ret0, _ := ret[0].(HealthReports)
	return ret0
}

// SentryHealthChecks indicates an expected call of SentryHealthChecks
func (mr *MockComponentMockRecorder) SentryHealthChecks(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SentryHealthChecks", reflect.TypeOf((*MockComponent)(nil).SentryHealthChecks), arg0)
}
