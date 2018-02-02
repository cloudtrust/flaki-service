package component

import (
	"context"

	"github.com/cloudtrust/flaki-service/pkg/flaki/module"
)

// Service is the interface that the service implements.
type Service interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// basicService contains the flaki distributed unique IDs generator.
type basicService struct {
	module flaki.Service
}

// NewBasicService returns the basic service.
func NewBasicService(flakiModule flaki.Service) Service {
	return &basicService{
		module: flakiModule,
	}
}

// NextID uses the flaki component to generate a unique ID.
func (s *basicService) NextID(ctx context.Context) (string, error) {
	return s.module.NextID(ctx)
}

// NextValidID uses the flaki component to generate a valid unique ID.
func (s *basicService) NextValidID(ctx context.Context) string {
	return s.module.NextValidID(ctx)
}
