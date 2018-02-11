package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	sentry "github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"
)

func TestComponentTrackingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var mockComponent = &mockComponent{fail: true}
	var mockSentry = &mockSentry{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeComponentTrackingMW(mockSentry)(mockComponent)

	// NextID.
	mockSentry.called = false
	mockSentry.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockSentry.called)
	assert.Equal(t, corrID, mockSentry.correlationID)

	// NextValidID never returns an error.
	mockSentry.called = false
	mockSentry.correlationID = ""
	m.NextValidID(ctx)
	assert.False(t, mockSentry.called)

	// NextID without correlation ID.
	mockSentry.called = false
	mockSentry.correlationID = ""
	m.NextID(context.Background())
	assert.True(t, mockSentry.called)
	assert.Zero(t, mockSentry.correlationID)
}

// Mock Sentry.
type mockSentry struct {
	called        bool
	correlationID string
}

func (client *mockSentry) CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	client.called = true
	client.correlationID = tags["correlation_id"]
	return ""
}
