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

	// NextID.
	assert.False(t, mockLogger.Called)
	srv.NextID(context.Background())
	assert.True(t, mockLogger.Called)

	// NextValidID.
	mockLogger.Called = false
	srv.NextValidID(context.Background())
	assert.True(t, mockLogger.Called)

}
func TestErrorMiddleware(t *testing.T) {
	var mockSentry = &mockSentry{Called: false}

	var srv = MakeErrorMiddleware(mockSentry)(&mockFlakiService{
		fail: true,
	})

	// NextID.
	assert.False(t, mockSentry.Called)
	srv.NextID(context.Background())
	assert.True(t, mockSentry.Called)

	// NextValidID.
	mockSentry.Called = false
	srv.NextValidID(context.Background())
	// NextValidID never returns an error.
	assert.False(t, mockSentry.Called)

	// With correlationID.
	rand.Seed(time.Now().UnixNano())
	var id = rand.Uint64()
	var idStr = strconv.FormatUint(id, 10)

	mockSentry.Called = false
	assert.Zero(t, mockSentry.CorrelationID)
	srv.NextID(context.WithValue(context.Background(), "id", id))
	assert.Equal(t, idStr, mockSentry.CorrelationID)
}

// Mock Flaki service. If fail is set to true, it returns an error.
type mockFlakiService struct {
	fail bool
}

func (s *mockFlakiService) NextID(context.Context) (uint64, error) {
	if s.fail {
		return 0, fmt.Errorf("fail")
	}
	return 0, nil
}

func (s *mockFlakiService) NextValidID(context.Context) uint64 {
	return 0
}

// Mock Logger.
type mockLogger struct {
	Called bool
}

func (l *mockLogger) Log(keyvals ...interface{}) error {
	l.Called = true
	return nil
}

// Mock Sentry.
type mockSentry struct {
	Called        bool
	CorrelationID string
}

func (client *mockSentry) CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	client.Called = true
	client.CorrelationID = tags["correlationID"]
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
