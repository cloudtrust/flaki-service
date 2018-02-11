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

func TestInstrumentingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockModule = &mockModule{fail: false, id: flakiID}
	var mockCounter = &mockCounter{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeModuleInstrumentingMW(mockCounter)(mockModule)

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
