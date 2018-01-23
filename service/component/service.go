package component

import (
	"context"

	"github.com/cloudtrust/flaki-service/service/module"
)

// Service is the interface that the services implement.
type Service interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

type basicService struct {
	module flaki.Service
}

// NewBasicService returns the basic service
func NewBasicService(flakiModule flaki.Service) Service {
	return &basicService{
		module: flakiModule,
	}
}

func (s *basicService) NextID(ctx context.Context) (string, error) {
	return s.module.NextID(ctx)
}

func (s *basicService) NextValidID(ctx context.Context) string {
	return s.module.NextValidID(ctx)
}
