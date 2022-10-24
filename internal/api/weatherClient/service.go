package weatherClient

import (
	"WeatherServiceAPI/pkg/logging"
	"context"
	"time"
)

type service struct {
	storage Storage
	logger  *logging.Logger
}

func (s service) FindInfoByCityAndDate(ctx context.Context, city string, date time.Time) (weatherDataJson string, err error) {
	return s.storage.FindInfoByCityAndDate(ctx, city, date)
}

func (s service) FindBriefInfo(ctx context.Context, city string) (BriefWeatherCity, error) {
	return s.storage.FindBriefInfo(ctx, city)
}

func (s service) Create(ctx context.Context, cityID string, data WeatherData) error {
	return s.storage.Create(ctx, cityID, data)
}

func NewService(storage Storage, logger *logging.Logger) (Service, error) {
	return &service{
		storage: storage,
		logger:  logger,
	}, nil
}

type Service interface {
	Create(ctx context.Context, cityID string, data WeatherData) error
	FindBriefInfo(ctx context.Context, city string) (BriefWeatherCity, error)
	FindInfoByCityAndDate(ctx context.Context, city string, date time.Time) (weatherDataJson string, err error)
}
