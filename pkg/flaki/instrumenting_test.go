package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
)

func TestComponentInstrumentingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockComponent = &mockComponent{fail: false, id: flakiID}
	var mockHistogram = &mockHistogram{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeComponentInstrumentingMW(mockHistogram)(mockComponent)

	// NextID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockHistogram.called)
	assert.Equal(t, corrID, mockHistogram.correlationID)

	// NextValidID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockHistogram.called)
	assert.Equal(t, corrID, mockHistogram.correlationID)

	// NextID without correlation ID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextID(context.Background())
	assert.True(t, mockHistogram.called)
	assert.Equal(t, flakiID, mockHistogram.correlationID)

	// NextValidID without correlation ID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextValidID(context.Background())
	assert.True(t, mockHistogram.called)
	assert.Equal(t, flakiID, mockHistogram.correlationID)
}

func TestModuleInstrumentingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockModule = &mockModule{fail: false, id: flakiID}
	var mockHistogram = &mockHistogram{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeModuleInstrumentingMW(mockHistogram)(mockModule)

	// NextID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockHistogram.called)
	assert.Equal(t, corrID, mockHistogram.correlationID)

	// NextValidID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockHistogram.called)
	assert.Equal(t, corrID, mockHistogram.correlationID)

	// NextID without correlation ID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextID(context.Background())
	assert.True(t, mockHistogram.called)
	assert.Equal(t, flakiID, mockHistogram.correlationID)

	// NextValidID without correlation ID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m.NextValidID(context.Background())
	assert.True(t, mockHistogram.called)
	assert.Equal(t, flakiID, mockHistogram.correlationID)
}

func TestModuleInstrumentingCounterMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockModule = &mockModule{fail: false, id: flakiID}
	var mockCounter = &mockCounter{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeModuleInstrumentingCounterMW(mockCounter)(mockModule)

	// NextID.
	mockCounter.called = false
	mockCounter.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockCounter.called)
	assert.Equal(t, corrID, mockCounter.correlationID)

	// NextValidID.
	mockCounter.called = false
	mockCounter.correlationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockCounter.called)
	assert.Equal(t, corrID, mockCounter.correlationID)

	// NextID without correlation ID.
	mockCounter.called = false
	mockCounter.correlationID = ""
	m.NextID(context.Background())
	assert.True(t, mockCounter.called)
	assert.Equal(t, flakiID, mockCounter.correlationID)

	// NextValidID without correlation ID.
	mockCounter.called = false
	mockCounter.correlationID = ""
	m.NextValidID(context.Background())
	assert.True(t, mockCounter.called)
	assert.Equal(t, flakiID, mockCounter.correlationID)
}

// Mock counter.
type mockCounter struct {
	called        bool
	correlationID string
}

func (c *mockCounter) With(labelValues ...string) metrics.Counter {
	for i, kv := range labelValues {
		if kv == "correlation_id" {
			c.correlationID = labelValues[i+1]
		}
	}
	return c
}

func (c *mockCounter) Add(delta float64) {
	c.called = true
}

// Mock histogram.
type mockHistogram struct {
	called        bool
	correlationID string
}

func (h *mockHistogram) With(labelValues ...string) metrics.Histogram {
	for i, kv := range labelValues {
		if kv == "correlation_id" {
			h.correlationID = labelValues[i+1]
		}
	}
	return h
}
func (h *mockHistogram) Observe(value float64) {
	h.called = true
}
