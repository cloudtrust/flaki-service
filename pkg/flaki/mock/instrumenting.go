// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-kit/kit/metrics (interfaces: Counter,Histogram)

// Package mock is a generated GoMock package.
package mock

import (
	metrics "github.com/go-kit/kit/metrics"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// Counter is a mock of Counter interface
type Counter struct {
	ctrl     *gomock.Controller
	recorder *CounterMockRecorder
}

// CounterMockRecorder is the mock recorder for Counter
type CounterMockRecorder struct {
	mock *Counter
}

// NewCounter creates a new mock instance
func NewCounter(ctrl *gomock.Controller) *Counter {
	mock := &Counter{ctrl: ctrl}
	mock.recorder = &CounterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Counter) EXPECT() *CounterMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *Counter) Add(arg0 float64) {
	m.ctrl.Call(m, "Add", arg0)
}

// Add indicates an expected call of Add
func (mr *CounterMockRecorder) Add(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*Counter)(nil).Add), arg0)
}

// With mocks base method
func (m *Counter) With(arg0 ...string) metrics.Counter {
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(metrics.Counter)
	return ret0
}

// With indicates an expected call of With
func (mr *CounterMockRecorder) With(arg0 ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*Counter)(nil).With), arg0...)
}

// Histogram is a mock of Histogram interface
type Histogram struct {
	ctrl     *gomock.Controller
	recorder *HistogramMockRecorder
}

// HistogramMockRecorder is the mock recorder for Histogram
type HistogramMockRecorder struct {
	mock *Histogram
}

// NewHistogram creates a new mock instance
func NewHistogram(ctrl *gomock.Controller) *Histogram {
	mock := &Histogram{ctrl: ctrl}
	mock.recorder = &HistogramMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Histogram) EXPECT() *HistogramMockRecorder {
	return m.recorder
}

// Observe mocks base method
func (m *Histogram) Observe(arg0 float64) {
	m.ctrl.Call(m, "Observe", arg0)
}

// Observe indicates an expected call of Observe
func (mr *HistogramMockRecorder) Observe(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Observe", reflect.TypeOf((*Histogram)(nil).Observe), arg0)
}

// With mocks base method
func (m *Histogram) With(arg0 ...string) metrics.Histogram {
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(metrics.Histogram)
	return ret0
}

// With indicates an expected call of With
func (mr *HistogramMockRecorder) With(arg0 ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*Histogram)(nil).With), arg0...)
}