package flakic

import (
	"context"
)

// Component is the Flaki component.
type Component interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// Module is the Flaki module interface.
type Module interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// Component is the Flaki component.
type component struct {
	module Module
}

// New returns a Flaki component.
func New(module Module) Component {
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
