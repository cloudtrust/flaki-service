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

func TestErrorMiddleware(t *testing.T) {
	var mockSentry = &mockSentry{}

	var srv = MakeErrorMiddleware(mockSentry)(&mockFlakiService{
		fail: true,
	})

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	// NextID.
	mockSentry.Called = false
	mockSentry.CorrelationID = ""
	srv.NextID(ctx)
	assert.True(t, mockSentry.Called)
	assert.Equal(t, id, mockSentry.CorrelationID)

	// NextValidID never returns an error.
	mockSentry.Called = false
	mockSentry.CorrelationID = ""
	srv.NextValidID(ctx)
	assert.False(t, mockSentry.Called)

	// NextID without correlation ID.
	var f = func() {
		srv.NextID(context.Background())
	}
	assert.Panics(t, f)
}

// Mock Sentry.
type mockSentry struct {
	Called        bool
	CorrelationID string
}

func (client *mockSentry) CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	client.Called = true
	client.CorrelationID = tags["correlation_id"]
	return ""
}
