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
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockComponent = &mockComponent{fail: false, id: flakiID}
	var mockLogger = &mockLogger{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeComponentLoggingMW(mockLogger)(mockComponent)

	// NextID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockLogger.called)
	assert.Equal(t, corrID, mockLogger.correlationID)

	// NextValidID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockLogger.called)
	assert.Equal(t, corrID, mockLogger.correlationID)

	// NextID without correlation ID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextID(context.Background())
	assert.True(t, mockLogger.called)
	assert.Equal(t, flakiID, mockLogger.correlationID)

	// NextValidID without correlation ID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextValidID(context.Background())
	assert.True(t, mockLogger.called)
	assert.Equal(t, flakiID, mockLogger.correlationID)
}

func TestModuleLoggingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockModule = &mockModule{fail: false, id: flakiID}
	var mockLogger = &mockLogger{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeModuleLoggingMW(mockLogger)(mockModule)

	// NextID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockLogger.called)
	assert.Equal(t, corrID, mockLogger.correlationID)

	// NextValidID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockLogger.called)
	assert.Equal(t, corrID, mockLogger.correlationID)

	// NextID without correlation ID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextID(context.Background())
	assert.True(t, mockLogger.called)
	assert.Equal(t, flakiID, mockLogger.correlationID)

	// NextValidID without correlation ID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m.NextValidID(context.Background())
	assert.True(t, mockLogger.called)
	assert.Equal(t, flakiID, mockLogger.correlationID)
}

// Mock Logger.
type mockLogger struct {
	called        bool
	correlationID string
}

func (l *mockLogger) Log(keyvals ...interface{}) error {
	l.called = true

	for i, kv := range keyvals {
		if kv == "correlation_id" {
			l.correlationID = keyvals[i+1].(string)
		}
	}
	return nil
}
