package cityClient

import (
	"WeatherServiceAPI/pkg/logging"
	"context"
)

type service struct {
	storage Storage
	logger  *logging.Logger
}

func NewService(storage Storage, logger *logging.Logger) (Service, error) {
	return &service{
		storage: storage,
		logger:  logger,
	}, nil
}

type Service interface {
	Create(ctx context.Context, data CityData) error
	FindAll(ctx context.Context) ([]CityData, error)
}

func (s service) Create(ctx context.Context, data CityData) error {
	err := s.storage.Create(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (s service) FindAll(ctx context.Context) ([]CityData, error) {
	return s.storage.FindAll(ctx)
}
