package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComponentLoggingMW(t *testing.T) {
	var mockLogger = &mockLogger{}

	var m = MakeComponentLoggingMW(mockLogger)(&mockModule{fail: false})

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	// NextID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
	m.NextID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

	// NextValidID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

	// NextID without correlation ID.
	var f = func() {
		m.NextID(context.Background())
	}
	assert.Panics(t, f)

	// NextValidID without correlation ID.
	f = func() {
		m.NextValidID(context.Background())
	}
	assert.Panics(t, f)
}

func TestModuleLoggingMW(t *testing.T) {
	var mockLogger = &mockLogger{}
	var mockFlaki = &mockFlaki{}

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	var m = NewModule(mockFlaki)
	m = MakeModuleLoggingMW(mockLogger)(m)

	// NextID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
	m.NextID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

	// NextValidID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

	// NextID without correlation ID.
	var f = func() {
		m.NextID(context.Background())
	}
	assert.Panics(t, f)

	// NextValidID without correlation ID.
	f = func() {
		m.NextValidID(context.Background())
	}
	assert.Panics(t, f)
}

// Mock Logger.
type mockLogger struct {
	Called        bool
	CorrelationID string
}

func (l *mockLogger) Log(keyvals ...interface{}) error {
	l.Called = true

	for i, kv := range keyvals {
		if kv == "correlation_id" {
			l.CorrelationID = keyvals[i+1].(string)
		}
	}
	return nil
}
