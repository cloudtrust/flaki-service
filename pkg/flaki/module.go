package flaki

//go:generate mockgen -destination=./mock/module.go -package=mock -mock_names=Module=Module,Flaki=Flaki github.com/cloudtrust/flaki-service/pkg/flaki Module,Flaki

import (
	"context"

	"github.com/pkg/errors"
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

// NewModule returns a Flaki module.
func NewModule(flaki Flaki) Module {
	return &module{
		flaki: flaki,
	}
}

// NextID generates a unique string ID.
func (m *module) NextID(_ context.Context) (string, error) {
	var id, err = m.flaki.NextIDString()
	if err != nil {
		return "", errors.Wrap(err, "flaki could not generate ID")
	}
	return id, nil
}

// NextValidID generates a unique string ID.
func (m *module) NextValidID(_ context.Context) string {
	return m.flaki.NextValidIDString()
}
