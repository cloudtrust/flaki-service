package flakim

import (
	"context"
)

// Module is the interface of the flaki Module
type Module interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// Flaki is the interface of the distributed unique IDs generator.
type Flaki interface {
	NextIDString() (string, error)
	NextValidIDString() string
}

// Module is the module using the Flaki generator to generate unique IDs.
type module struct {
	flaki Flaki
}

// New returns a Flaki module.
func New(flaki Flaki) Module {
	return &module{
		flaki: flaki,
	}
}

// NextID generates a unique string ID.
func (s *module) NextID(_ context.Context) (string, error) {
	return s.flaki.NextIDString()
}

// NextValidID generates a unique string ID.
func (s *module) NextValidID(_ context.Context) string {
	return s.flaki.NextValidIDString()
}
