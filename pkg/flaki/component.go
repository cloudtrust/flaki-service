package flaki

//go:generate mockgen -destination=./mock/component.go -package=mock -mock_names=Component=Component github.com/cloudtrust/flaki-service/pkg/flaki Component

import (
	"context"

	"github.com/pkg/errors"
)

// Component is the Flaki component interface.
type Component interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// Component is the Flaki component.
type component struct {
	module Module
}

// NewComponent returns a Flaki component.
func NewComponent(module Module) Component {
	return &component{
		module: module,
	}
}

// NextID generates a unique string ID.
func (c *component) NextID(ctx context.Context) (string, error) {
	var id, err = c.module.NextID(ctx)
	if err != nil {
		return "", errors.Wrap(err, "module could not generate ID")
	}
	return id, nil
}

// NextValidID generates a unique string ID.
func (c *component) NextValidID(ctx context.Context) string {
	return c.module.NextValidID(ctx)
}
