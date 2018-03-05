package flaki

//go:generate mockgen -destination=./mock/component.go -package=mock -mock_names=Component=Component github.com/cloudtrust/flaki-service/pkg/flaki Component

import (
	"context"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
)

type key int

const (
	// CorrelationIDKey is the key for the correlation ID in the context.
	CorrelationIDKey key = iota

	// LoggingCorrelationIDKey is the logs correlation ID key
	LoggingCorrelationIDKey = "correlation_id"
	// MetricCorrelationIDKey is the metrics correlation ID key.
	MetricCorrelationIDKey = "correlation_id"
	// TracingCorrelationIDKey is the trace correlation ID key.
	TracingCorrelationIDKey = "correlation_id"
	// TrackingCorrelationIDKey is the error correlation ID key.
	TrackingCorrelationIDKey = "correlation_id"
)

// Component is the flaki component interface.
type Component interface {
	NextID(context.Context, *fb.FlakiRequest) (*fb.FlakiReply, error)
	NextValidID(context.Context, *fb.FlakiRequest) *fb.FlakiReply
}

// Component is the flaki component.
type component struct {
	module Module
}

// NewComponent returns a flaki component.
func NewComponent(module Module) Component {
	return &component{
		module: module,
	}
}

// NextID generates a unique string ID.
func (c *component) NextID(ctx context.Context, req *fb.FlakiRequest) (*fb.FlakiReply, error) {
	var id, err = c.module.NextID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "module could not generate ID")
	}

	return encodeFlakiReply(id), nil
}

// NextValidID generates a unique string ID.
func (c *component) NextValidID(ctx context.Context, req *fb.FlakiRequest) *fb.FlakiReply {
	var id = c.module.NextValidID(ctx)
	return encodeFlakiReply(id)
}

// encodeFlakiReply encode the flatbuffer reply.
func encodeFlakiReply(id string) *fb.FlakiReply {
	var b = flatbuffers.NewBuilder(0)
	var str = b.CreateString(id)

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, str)
	b.Finish(fb.FlakiReplyEnd(b))

	return fb.GetRootAsFlakiReply(b.FinishedBytes(), 0)
}
