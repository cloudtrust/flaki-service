package flaki

//go:generate mockgen -destination=./mock/flaki.go -package=mock -mock_names=Flaki=Flaki github.com/cloudtrust/flaki-service/pkg/flaki Flaki

import (
	"context"

	"github.com/pkg/errors"
)

// Flaki is the interface of the distributed unique IDs generator.
type Flaki interface {
	NextIDString() (string, error)
	NextValidIDString() string
}

// Module is the module using the Flaki generator to generate unique IDs.
type Module struct {
	flaki Flaki
}

// NewModule returns a Flaki module.
func NewModule(flaki Flaki) *Module {
	return &Module{
		flaki: flaki,
	}
}

// NextID generates a unique string ID.
func (m *Module) NextID(_ context.Context) (string, error) {
	var id, err = m.flaki.NextIDString()
	if err != nil {
		return "", errors.Wrap(err, "flaki could not generate ID")
	}
	return id, nil
}

// NextValidID generates a unique string ID.
func (m *Module) NextValidID(_ context.Context) string {
	return m.flaki.NextValidIDString()
}
