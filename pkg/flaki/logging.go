package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

// Middleware on Flaki Module.
type Middleware func(Module) Module

// Logging Middleware.
type loggingMiddleware struct {
	logger log.Logger
	next   Module
}

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(log log.Logger) Middleware {
	return func(next Module) Module {
		return &loggingMiddleware{
			logger: log,
			next:   next,
		}
	}
}

// loggingMiddleware implements Service.
func (m *loggingMiddleware) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

// loggingMiddleware implements Service.
func (m *loggingMiddleware) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

// Middleware on FlakiComponent.
type Middleware func(FlakiComponent) FlakiComponent

// Logging Middleware.
type loggingMiddleware struct {
	logger log.Logger
	next   FlakiComponent
}

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(log log.Logger) Middleware {
	return func(next FlakiComponent) FlakiComponent {
		return &loggingMiddleware{
			logger: log,
			next:   next,
		}
	}
}

// loggingMiddleware implements FlakiComponent.
func (m *loggingMiddleware) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

// loggingMiddleware implements FlakiComponent.
func (m *loggingMiddleware) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

func TestLoggingMiddleware(t *testing.T) {
	var mockLogger = &mockLogger{}

	var srv = MakeLoggingMiddleware(mockLogger)(&mockFlakiService{
		fail: false,
	})

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	// NextID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
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
