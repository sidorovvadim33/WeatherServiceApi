package cityClient

import (
	"WeatherServiceAPI/internal/config"
	"WeatherServiceAPI/pkg/logging"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var cities = [20]string{
	"London",
	"Moscow",
	"Kazan",
	"Naberezhnye Chelny",
	"Warsaw",
	"Lisbon",
	"Beijing",
	"Nizhny Novgorod",
	"Batumi",
	"Oslo",
	"Helsinki",
	"Riga",
	"Berlin",
	"Prague",
	"Paris",
	"Milan",
	"Barcelona",
	"Rome",
	"Kosovo",
	"Istanbul",
}

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

func (c *client) RefreshCitiesCoordinatesAsync(citiesService Service) error {
	responseChan := make(chan *http.Response, len(cities))

	for _, city := range cities {
		city := city
		go func() {
			url := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", city, c.cfg.ApiID)

			r, err := http.Get(url)
			if err != nil {
				c.logger.Fatal(err)
			}

			responseChan <- r
		}()
	}

	var s []CityData
	count := 0

	for response := range responseChan {
		err := json.NewDecoder(response.Body).Decode(&s)
		if err != nil {
			log.Fatal(err)
		}
		err = response.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		count++

		if count == len(cities) {
			close(responseChan)
		}

		err = citiesService.Create(context.TODO(), s[0])
		if err != nil {
			panic(err)
		}
	}

	c.logger.Info("cities data refreshed")
	return nil
}
