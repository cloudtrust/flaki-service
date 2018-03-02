package flaki

//go:generate mockgen -destination=./mock/tracking.go -package=mock -mock_names=Sentry=Sentry  github.com/cloudtrust/flaki-service/pkg/flaki Sentry

import (
	"context"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
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
func (m *trackingComponentMW) NextID(ctx context.Context, req *fb.FlakiRequest) (*fb.FlakiReply, error) {
	var reply, err = m.next.NextID(ctx, req)
	if err != nil {
		var tags = map[string]string{}
		if id := ctx.Value("correlation_id"); id != nil {
			tags["correlation_id"] = id.(string)
		}
		m.client.CaptureError(err, tags)
	}
	return reply, err
}

// trackingComponentMW implements Component.
func (m *trackingComponentMW) NextValidID(ctx context.Context, req *fb.FlakiRequest) *fb.FlakiReply {
	return m.next.NextValidID(ctx, req)
}
