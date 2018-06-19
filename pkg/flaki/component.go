package flaki

//go:generate mockgen -destination=./mock/module.go -package=mock -mock_names=IDGeneratorModule=IDGeneratorModule github.com/cloudtrust/flaki-service/pkg/flaki IDGeneratorModule

import (
	"context"

	"github.com/cloudtrust/flaki-service/api/fb"
	"github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
)

// IDGeneratorModule is the interface of the flaki Module.
type IDGeneratorModule interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// Component is the flaki component.
type Component struct {
	module IDGeneratorModule
}

// NewComponent returns a flaki component.
func NewComponent(module IDGeneratorModule) *Component {
	return &Component{
		module: module,
	}
}

// NextID generates a unique string ID.
func (c *Component) NextID(ctx context.Context, req *fb.FlakiRequest) (*fb.FlakiReply, error) {
	var id, err = c.module.NextID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "module could not generate ID")
	}

	return encodeFlakiReply(id), nil
}

// NextValidID generates a unique string ID.
func (c *Component) NextValidID(ctx context.Context, req *fb.FlakiRequest) *fb.FlakiReply {
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
