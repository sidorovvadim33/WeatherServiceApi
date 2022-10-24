package weatherApiClient

import (
	"WeatherServiceAPI/internal/api/cityClient"
	"WeatherServiceAPI/pkg/client/postgresql"
	"WeatherServiceAPI/pkg/logging"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgconn"
)

type db struct {
	client postgresql.Client
	logger *logging.Logger
}

func (d db) Create(ctx context.Context, data cityClient.CityData) error {
	q := `INSERT INTO cities (name, lat, lon, country) VALUES ($1, $2, $3, $4) ON CONFLICT (name, country) DO NOTHING;`

	d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
	_, err := d.client.Exec(ctx, q, data.Name, data.Lat, data.Lon, data.Country)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			d.logger.Error(newErr)
			return newErr
		}
		return err
	}

	return nil
}

func (d db) FindAll(ctx context.Context) ([]cityClient.CityData, error) {
	q := `SELECT id, name, lat, lon, country FROM cities;`

	rows, err := d.client.Query(ctx, q)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			d.logger.Error(newErr)
			return nil, newErr
		}
		return nil, err
	}

	cities := make([]cityClient.CityData, 0)

	for rows.Next() {
		var cty cityClient.CityData

		err = rows.Scan(&cty.Id, &cty.Name, &cty.Lat, &cty.Lon, &cty.Country)
		if err != nil {
			return nil, err
		}

		cities = append(cities, cty)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cities, nil

}

func NewStorage(client postgresql.Client, logger *logging.Logger) cityClient.Storage {
	return &db{
		client: client,
		logger: logger,
	}
}
