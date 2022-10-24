package weatherClient

import (
	"WeatherServiceAPI/internal/api/cityClient"
	"WeatherServiceAPI/internal/config"
	"WeatherServiceAPI/pkg/logging"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type client struct {
	logger *logging.Logger
	cfg    config.Config
}

func NewClient(logger *logging.Logger, cfg config.Config) *client {
	return &client{
		logger: logger,
		cfg:    cfg,
	}
}

type cwStruct struct {
	cityId  string
	weather WeatherData
}

func (c *client) RefreshWeatherDataAsync(cities []cityClient.CityData, wService Service) error {
	cwChan := make(chan cwStruct, len(cities))

	for _, city := range cities {
		city := city
		go func() {
			var weather WeatherData
			url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric", city.Lat, city.Lon, c.cfg.ApiID)

			r, err := http.Get(url)
			if err != nil {
				c.logger.Fatal(err)
			}
			defer r.Body.Close()

			err = json.NewDecoder(r.Body).Decode(&weather)
			if err != nil {
				c.logger.Fatalf("%v", err)
			}

			cwChan <- cwStruct{
				cityId:  city.Id,
				weather: weather,
			}
		}()
	}

	counter := 0
	for cw := range cwChan {
		counter++

		if counter == len(cities) {
			close(cwChan)
		}

		err := wService.Create(context.TODO(), cw.cityId, cw.weather)
		if err != nil {
			return err
		}
	}

	c.logger.Info("weather data refreshed")

	return nil
}
