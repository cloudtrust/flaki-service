package component

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	sentry "github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	var mockLogger = &mockLogger{Called: false}

	var srv = MakeLoggingMiddleware(mockLogger)(&mockFlakiService{
		fail: false,
	})

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation-id", id)

	// NextID.
	assert.False(t, mockLogger.Called)
	assert.Zero(t, mockLogger.CorrelationID)
	srv.NextID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

	// NextValidID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
	srv.NextValidID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

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
func TestErrorMiddleware(t *testing.T) {
	var mockSentry = &mockSentry{Called: false}

	var srv = MakeErrorMiddleware(mockSentry)(&mockFlakiService{
		fail: true,
	})

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation-id", id)

	// NextID.
	assert.False(t, mockSentry.Called)
	assert.Zero(t, mockSentry.CorrelationID)
	srv.NextID(ctx)
	assert.True(t, mockSentry.Called)
	assert.Equal(t, id, mockSentry.CorrelationID)

	// NextValidID never returns an error.
	mockSentry.Called = false
	srv.NextValidID(ctx)
	assert.False(t, mockSentry.Called)

	// NextID without correlation ID.
	var f = func() {
		srv.NextID(context.Background())
	}
	assert.Panics(t, f)
}

// Mock Flaki service. If fail is set to true, it returns an error.
type mockFlakiService struct {
	fail bool
}

func (s *mockFlakiService) NextID(context.Context) (string, error) {
	if s.fail {
		return "", fmt.Errorf("fail")
	}
	return "", nil
}

func (s *mockFlakiService) NextValidID(context.Context) string {
	return ""
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

// Mock Sentry.
type mockSentry struct {
	Called        bool
	CorrelationID string
}

func (client *mockSentry) CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	client.Called = true
	client.CorrelationID = tags["correlation-id"]
	return ""
}
func (client *mockSentry) SetDSN(dsn string) error           { return nil }
func (client *mockSentry) SetRelease(release string)         {}
func (client *mockSentry) SetEnvironment(environment string) {}
func (client *mockSentry) SetDefaultLoggerName(name string)  {}
func (client *mockSentry) Capture(packet *sentry.Packet, captureTags map[string]string) (eventID string, ch chan error) {
	return "", nil
}
func (client *mockSentry) CaptureMessage(message string, tags map[string]string, interfaces ...sentry.Interface) string {
	return ""
}
func (client *mockSentry) CaptureMessageAndWait(message string, tags map[string]string, interfaces ...sentry.Interface) string {
	return ""
}
func (client *mockSentry) CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	return ""
}
func (client *mockSentry) CapturePanic(f func(), tags map[string]string, interfaces ...sentry.Interface) (err interface{}, errorID string) {
	return nil, ""
}
func (client *mockSentry) CapturePanicAndWait(f func(), tags map[string]string, interfaces ...sentry.Interface) (err interface{}, errorID string) {
	return nil, ""
}
func (client *mockSentry) Close()                     {}
func (client *mockSentry) Wait()                      {}
func (client *mockSentry) URL() string                { return "" }
func (client *mockSentry) ProjectID() string          { return "" }
func (client *mockSentry) Release() string            { return "" }
func (client *mockSentry) IncludePaths() []string     { return nil }
func (client *mockSentry) SetIncludePaths(p []string) {}
