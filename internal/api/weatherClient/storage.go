package weatherClient

import (
	"context"
	"time"
)

type Storage interface {
	Create(ctx context.Context, cityID string, data WeatherData) error
	FindBriefInfo(ctx context.Context, city string) (BriefWeatherCity, error)
	FindInfoByCityAndDate(ctx context.Context, city string, date time.Time) (weatherDataJson string, err error)
}
