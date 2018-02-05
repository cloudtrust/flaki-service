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

func TestMetricMiddleware(t *testing.T) {
	var mockCounter = &mockCounter{}
	var mockFlaki = &mockFlaki{}

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	var srv = New(mockFlaki)
	srv = MakeMetricMiddleware(mockCounter)(srv)

	// NextID.
	mockCounter.Called = false
	mockCounter.CorrelationID = ""
	srv.NextID(ctx)
	assert.True(t, mockCounter.Called)
	assert.Equal(t, id, mockCounter.CorrelationID)

	// NextValidID.
	mockCounter.Called = false
	mockCounter.CorrelationID = ""
	srv.NextValidID(ctx)
	assert.True(t, mockCounter.Called)
	assert.Equal(t, id, mockCounter.CorrelationID)

	// NextID without correlation ID.
	var f = func() {
		srv.NextID(context.Background())
	}
	assert.Panics(t, f)

	// NextValidID without correlation ID.
	f = func() {
		srv.NextValidID(context.Background())
	}
	assert.Panics(t, f)
}

// Mock counter.
type mockCounter struct {
	Called        bool
	CorrelationID string
}

func (h *mockCounter) With(labelValues ...string) metrics.Counter {
	for i, kv := range labelValues {
		if kv == "correlation_id" {
			h.CorrelationID = labelValues[i+1]
		}
	}
	return h
}

func (h *mockCounter) Add(delta float64) {
	h.Called = true
}
