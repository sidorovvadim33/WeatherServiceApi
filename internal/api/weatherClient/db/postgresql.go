package weather

import (
	"WeatherServiceAPI/internal/api/weatherClient"
	"WeatherServiceAPI/pkg/client/postgresql"
	"WeatherServiceAPI/pkg/logging"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"time"
)

type db struct {
	client postgresql.Client
	logger *logging.Logger
}

func (d db) FindInfoByCityAndDate(ctx context.Context, city string, date time.Time) (weatherDataJson string, err error) {
	q := `SELECT w.data_json FROM weather as w join cities c on c.id = w.city_id where c.name = $1 AND w.date = $2;`

	d.logger.Debug(fmt.Sprintf("SQL Query: %s", q))

	rows := d.client.QueryRow(ctx, q, city, date.Format("2006-01-02 15:04:05"))

	err = rows.Scan(&weatherDataJson)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			d.logger.Error(newErr)
			return weatherDataJson, newErr
		}
		return weatherDataJson, err
	}

	return weatherDataJson, nil
}

func (d db) FindBriefInfo(ctx context.Context, city string) (wthr weatherClient.BriefWeatherCity, err error) {
	q := `SELECT c.country, c.name, AVG(w.temp), ARRAY(select innerW.date from weather as innerW where innerW.city_id = w.city_id order by innerW.date) FROM weather as w join cities c on c.id = w.city_id group by c.name, c.country, w.city_id having c.name = $1;`

	d.logger.Debug(fmt.Sprintf("SQL Query: %s", q))

	rows := d.client.QueryRow(ctx, q, city)

	err = rows.Scan(&wthr.Country, &wthr.Name, &wthr.AvgTemp, &wthr.DateTimeArray)
	if err != nil {
		return wthr, err
	}

	return wthr, nil
}

func (d db) Create(ctx context.Context, cityId string, weather weatherClient.WeatherData) error {
	q := `INSERT INTO weather (city_id, temp, date, data_json) VALUES ($1, $2, $3, $4) ON CONFLICT (city_id, date) DO UPDATE SET city_id = excluded.city_id,temp = $2, data_json = $4;`

	d.logger.Debug(fmt.Sprintf("SQL Query: %s", q))
	for _, w := range weather.List {
		bytes, err := json.Marshal(w)
		if err != nil {
			return err
		}
		_, err = d.client.Exec(ctx, q, cityId, w.Main.Temp, w.DtTxt, bytes)
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
	}

	return nil
}

func NewStorage(client postgresql.Client, logger *logging.Logger) weatherClient.Storage {
	return &db{
		client: client,
		logger: logger,
	}
}
