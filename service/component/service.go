package component

import (
	"context"
	"github.com/JohanDroz/flaki-service/service/module"
)

// Service is the interface that the services implement.
type Service interface {
	NextID(context.Context) (uint64, error)
	NextValidID(context.Context) uint64
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

func (s *basicService) NextID(ctx context.Context) (uint64, error) {
	return s.module.NextID(ctx)
}

func (s *basicService) NextValidID(ctx context.Context) uint64 {
	return s.module.NextValidID(ctx)
}
