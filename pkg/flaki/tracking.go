package flaki

//go:generate mockgen -destination=./mock/tracking.go -package=mock -mock_names=Sentry=Sentry  github.com/cloudtrust/flaki-service/pkg/flaki Sentry

import (
	"context"

	"github.com/cloudtrust/flaki-service/api/fb"
	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/log"
)

// Sentry interface.
type Sentry interface {
	CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
}

// Tracking middleware at component level.
type trackingComponentMW struct {
	sentry Sentry
	logger log.Logger
	next   IDGeneratorComponent
}

// MakeComponentTrackingMW makes an error tracking middleware, where the errors are logged and sent to Sentry.
func MakeComponentTrackingMW(sentry Sentry, logger log.Logger) func(IDGeneratorComponent) IDGeneratorComponent {
	return func(next IDGeneratorComponent) IDGeneratorComponent {
		return &trackingComponentMW{
			sentry: sentry,
			logger: logger,
			next:   next,
		}
	}
}

// trackingComponentMW implements Component.
func (m *trackingComponentMW) NextID(ctx context.Context, req *fb.FlakiRequest) (*fb.FlakiReply, error) {
	var reply, err = m.next.NextID(ctx, req)
	if err != nil {
		var corrID = ""
		if id := ctx.Value("correlation_id"); id != nil {
			corrID = id.(string)
		}
		m.sentry.CaptureError(err, map[string]string{"correlation_id": corrID})
		m.logger.Log("unit", "NextID", "correlation_id", corrID, "error", err.Error())
	}
	return reply, err
}

// trackingComponentMW implements Component.
func (m *trackingComponentMW) NextValidID(ctx context.Context, req *fb.FlakiRequest) *fb.FlakiReply {
	return m.next.NextValidID(ctx, req)
}
