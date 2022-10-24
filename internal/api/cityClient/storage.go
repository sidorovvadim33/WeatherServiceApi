package cityClient

import (
	"context"
)

type Storage interface {
	Create(ctx context.Context, city CityData) error
	FindAll(ctx context.Context) ([]CityData, error)
}
