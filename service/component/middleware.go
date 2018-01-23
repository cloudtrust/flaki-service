package component

import (
	"context"
	"time"

	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/log"
)

// Middleware on Service
type Middleware func(Service) Service

// Logging Middleware
type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// loggingMiddleware implements Service
func (m *loggingMiddleware) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation-id"), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

func (m *loggingMiddleware) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation-id"), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(log log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			logger: log,
			next:   next,
		}
	}
}

// Sentry interface
type sentryClient interface {
	SetDSN(dsn string) error
	SetRelease(release string)
	SetEnvironment(environment string)
	SetDefaultLoggerName(name string)
	Capture(packet *sentry.Packet, captureTags map[string]string) (eventID string, ch chan error)
	CaptureMessage(message string, tags map[string]string, interfaces ...sentry.Interface) string
	CaptureMessageAndWait(message string, tags map[string]string, interfaces ...sentry.Interface) string
	CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
	CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string
	CapturePanic(f func(), tags map[string]string, interfaces ...sentry.Interface) (err interface{}, errorID string)
	CapturePanicAndWait(f func(), tags map[string]string, interfaces ...sentry.Interface) (err interface{}, errorID string)
	Close()
	Wait()
	URL() string
	ProjectID() string
	Release() string
	IncludePaths() []string
	SetIncludePaths(p []string)
}

// Error Middleware
type errorMiddleware struct {
	client sentryClient
	next   Service
}

func (s *errorMiddleware) NextID(ctx context.Context) (string, error) {
	var id, err = s.next.NextID(ctx)
	if err != nil {
		s.client.CaptureErrorAndWait(err, map[string]string{"correlation-id": getStrIDFromContext(ctx)})
	}
	return id, err
}

func getStrIDFromContext(ctx context.Context) string {
	var id = ctx.Value("correlation-id")
	if id == nil {
		return ""
	}
	return id.(string)
}

func (s *errorMiddleware) NextValidID(ctx context.Context) string {
	return s.next.NextValidID(ctx)
}

// MakeErrorMiddleware makes an error middleware, where the errors are send to Sentry.
func MakeErrorMiddleware(client sentryClient) Middleware {
	return func(next Service) Service {
		return &errorMiddleware{
			client: client,
			next:   next,
		}
	}
}
