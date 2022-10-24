package db

import (
	"WeatherServiceAPI/internal/api/cityClient"
	"WeatherServiceAPI/internal/user"
	"WeatherServiceAPI/pkg/client/postgresql"
	"WeatherServiceAPI/pkg/logging"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
)

var _ user.Storage = &db{}

type db struct {
	client postgresql.Client
	logger *logging.Logger
}

func (d db) Create(ctx context.Context, user user.User) (string, error) {
	q := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING uuid;`

	if err := d.client.QueryRow(ctx, q, user.Email, user.Password).Scan(&user.UUID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			d.logger.Error(newErr)
			return "", newErr
		}

		return "", err
	}

	return user.UUID, nil
}

func (d db) CreateFavourite(ctx context.Context, user user.User, cityId string) error {
	q := `INSERT INTO user_favorites (user_id, city_id) VALUES ($1, $2) ON CONFLICT (user_id, city_id) DO NOTHING;`

	d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	_, err := d.client.Exec(ctx, q, user.UUID, cityId)
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

func (d db) FindFavourites(ctx context.Context, user user.User) ([]cityClient.CityData, error) {
	q := `SELECT c.* FROM user_favorites JOIN cities c on c.id = user_favorites.city_id WHERE user_id = $1;`

	rows, err := d.client.Query(ctx, q, user.UUID)
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

func (d db) FindByEmail(ctx context.Context, email string) (user user.User, err error) {
	q := `SELECT * FROM users WHERE email = $1;`

	d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	if err = d.client.QueryRow(ctx, q, email).Scan(&user.UUID, &user.Email, &user.Password); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			d.logger.Error(newErr)
			return user, newErr
		}
		return user, err
	}

	return user, nil

}

func (d db) FindOne(ctx context.Context, uuid string) (user user.User, err error) {
	q := `SELECT * FROM users WHERE uuid  = $1;`

	d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
	if err = d.client.QueryRow(ctx, q, uuid).Scan(&user.UUID, &user.Email, &user.Password); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			d.logger.Error(newErr)
			return user, newErr
		}
		return user, err
	}

	return user, nil
}

func (d db) Update(ctx context.Context, user user.User) (err error) {

	if user.Email != "" && user.Password != "" {
		q := `UPDATE users SET email = $2, password = $3 WHERE uuid = $1;`

		d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
		_, err = d.client.Exec(ctx, q, user.UUID, user.Email, user.Password)
	} else if user.Email != "" {
		q := `UPDATE users SET email = $2 WHERE uuid = $1;`

		d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
		_, err = d.client.Exec(ctx, q, user.UUID, user.Email)
	} else if user.Password != "" {
		q := `UPDATE users SET password = $2 WHERE uuid = $1;`

		d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
		_, err = d.client.Exec(ctx, q, user.UUID, user.Password)
	}

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

func (d db) Delete(ctx context.Context, uuid string) error {
	q := `DELETE FROM users WHERE uuid = $1;`
	d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	_, err := d.client.Exec(ctx, q, uuid)
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

func (d db) DeleteFavourite(ctx context.Context, user user.User, cityId string) error {
	q := `DELETE FROM user_favorites WHERE user_id = $1 AND city_id = $2;`

	d.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	_, err := d.client.Exec(ctx, q, user.UUID, cityId)
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

func NewStorage(client postgresql.Client, logger *logging.Logger) user.Storage {
	return &db{
		client: client,
		logger: logger,
	}
}
