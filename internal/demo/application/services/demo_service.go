package services

import (
	"context"

	"github.com/radius/radius-backend/internal/demo/application/dto"
	"go.uber.org/zap"
)

type DemoService struct {
	logger *zap.Logger
}

func NewDemoService(logger *zap.Logger) *DemoService {
	return &DemoService{
		logger: logger,
	}
}

func (s *DemoService) HandleHello(_ context.Context) (dto.HelloResponse, error) {
	// TODO: replace with real use case logic.
	return dto.HelloResponse{Message: "hello world"}, nil
}
