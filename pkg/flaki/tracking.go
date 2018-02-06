package flaki

import (
	"context"

	sentry "github.com/getsentry/raven-go"
)

// Sentry interface.
type Sentry interface {
	CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
}

// Tracking middleware at component level.
type trackingComponentMW struct {
	client Sentry
	next   Component
}

// MakeComponentTrackingMW makes an error tracking middleware, where the errors are sent to Sentry.
func MakeComponentTrackingMW(client Sentry) func(Component) Component {
	return func(next Component) Component {
		return &trackingComponentMW{
			client: client,
			next:   next,
		}
	}
}

// trackingComponentMW implements Component.
func (m *trackingComponentMW) NextID(ctx context.Context) (string, error) {
	var id, err = m.next.NextID(ctx)
	if err != nil {
		m.client.CaptureError(err, map[string]string{"correlation_id": ctx.Value("correlation_id").(string)})
	}
	return id, err
}

// trackingComponentMW implements Component.
func (m *trackingComponentMW) NextValidID(ctx context.Context) string {
	return m.next.NextValidID(ctx)
}
