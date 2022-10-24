package main

import (
	weather3 "WeatherServiceAPI/internal/api"
	"WeatherServiceAPI/internal/api/cityClient"
	weatherApiClient2 "WeatherServiceAPI/internal/api/cityClient/db"
	"WeatherServiceAPI/internal/api/weatherClient"
	weather2 "WeatherServiceAPI/internal/api/weatherClient/db"
	"WeatherServiceAPI/internal/config"
	"WeatherServiceAPI/internal/user"
	"WeatherServiceAPI/internal/user/db"
	"WeatherServiceAPI/pkg/client/postgresql"
	"WeatherServiceAPI/pkg/logging"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"time"
)

func main() {
	logger := logging.GetLogger()

	logger.Info("create router")
	router := httprouter.New()

	cfg := config.GetConfig()
	postgresSQLClient, err := postgresql.NewClient(context.TODO(), 3, cfg.Storage)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	citiesService := AddCitiesData(postgresSQLClient, logger, cfg)
	weatherService := AddWeatherData(postgresSQLClient, logger, cfg, citiesService)

	logger.Info("register user handler")
	handler := weather3.NewHandler(logger, citiesService, weatherService)
	handler.Register(router)

	userStorage := db.NewStorage(postgresSQLClient, logger)
	userService, err := user.NewService(userStorage, logger)
	if err != nil {
		logger.Fatal(err)
	}

	usersHandler := user.NewHandler(logger, userService)
	usersHandler.Register(router)

	start(router, cfg)
}

func start(router *httprouter.Router, cfg *config.Config) {
	logger := logging.GetLogger()
	logger.Info("start application")

	var listener net.Listener
	var listenErr error

	logger.Info("listen tcp")
	listener, listenErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.Port))
	logger.Infof("server is listening port %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)

	if listenErr != nil {
		logger.Fatal(listenErr)
	}

	server := &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}

func AddCitiesData(postgreSQLClient *pgxpool.Pool, logger *logging.Logger, cfg *config.Config) cityClient.Service {
	logger.Info("getting cities data from api source")

	cClient := cityClient.NewClient(logger, *cfg)

	citiesStorage := weatherApiClient2.NewStorage(postgreSQLClient, logger)
	citiesService, err := cityClient.NewService(citiesStorage, logger)
	if err != nil {
		panic(err)
	}

	logger.Info("refresh cities data in database")
	err = cClient.RefreshCitiesCoordinatesAsync(citiesService)
	if err != nil {
		logger.Fatalf("failed to refresh cities data. due to error: %v", err)
	}

	return citiesService
}

func AddWeatherData(postgreSQLClient *pgxpool.Pool, logger *logging.Logger, cfg *config.Config, citiesService cityClient.Service) weatherClient.Service {
	wClient := weatherClient.NewClient(logger, *cfg)
	wStorage := weather2.NewStorage(postgreSQLClient, logger)
	wService, err := weatherClient.NewService(wStorage, logger)
	if err != nil {
		panic(err)
	}
	logger.Info("getting weather data from api source")

	refreshFunc := func() {
		cities, err := citiesService.FindAll(context.TODO())
		if err != nil {
			logger.Fatalf("failed to get cities from database. due to error: %v", err)
		}

		err = wClient.RefreshWeatherDataAsync(cities, wService)
		if err != nil {
			logger.Fatalf("failed to refresh weather data. due to error: %v", err)
		}
	}
	refreshFunc()

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			refreshFunc()
		}
	}()

	return wService
}
