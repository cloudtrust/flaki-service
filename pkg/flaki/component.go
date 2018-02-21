package flaki

//go:generate mockgen -destination=./mock/component.go -package=mock -mock_names=Component=Component github.com/cloudtrust/flaki-service/pkg/flaki Component

import (
	"context"
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

// New returns a Flaki component.
func NewComponent(module Module) Component {
	return &component{
		module: module,
	}
}

// NextID generates a unique string ID.
func (c *component) NextID(ctx context.Context) (string, error) {
	return c.module.NextID(ctx)
}

// NextValidID generates a unique string ID.
func (c *component) NextValidID(ctx context.Context) string {
	return c.module.NextValidID(ctx)
}
