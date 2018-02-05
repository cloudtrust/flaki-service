package flaki

import (
	"context"

	sentry "github.com/getsentry/raven-go"
)

// Sentry interface.
type Sentry interface {
	CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
}

// Error Middleware.
type errorMiddleware struct {
	client Sentry
	next   FlakiComponent
}

// MakeErrorMiddleware makes an error handling middleware, where the errors are sent to Sentry.
func MakeErrorMiddleware(client Sentry) Middleware {
	return func(next FlakiComponent) FlakiComponent {
		return &errorMiddleware{
			client: client,
			next:   next,
		}
	}
}

// errorMiddleware implements FlakiComponent.
func (m *errorMiddleware) NextID(ctx context.Context) (string, error) {
	var id, err = m.next.NextID(ctx)
	if err != nil {
		m.client.CaptureError(err, map[string]string{"correlation_id": ctx.Value("correlation_id").(string)})
	}
	return id, err
}

// errorMiddleware implements FlakiComponent.
func (m *errorMiddleware) NextValidID(ctx context.Context) string {
	return m.next.NextValidID(ctx)
}
