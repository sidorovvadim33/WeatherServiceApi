package api

import (
	_ "WeatherServiceAPI/docs"
	"WeatherServiceAPI/internal/api/cityClient"
	"WeatherServiceAPI/internal/api/weatherClient"
	"WeatherServiceAPI/internal/apperror"
	"WeatherServiceAPI/internal/handlers"
	"WeatherServiceAPI/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
	"math"
	"net/http"
	"sort"
	"time"
)

const (
	citiesUrl       = "/api/cities"
	cityInfoUrl     = "/api/cities/:city"
	cityDateInfoURL = "/api/cities/:city/:date"
)

type handler struct {
	logger         *logging.Logger
	cityService    cityClient.Service
	weatherService weatherClient.Service
}

func NewHandler(logger *logging.Logger, cityService cityClient.Service, weatherService weatherClient.Service) handlers.Handler {
	return &handler{
		logger:         logger,
		cityService:    cityService,
		weatherService: weatherService,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, citiesUrl, apperror.Middleware(h.GetAvailableCities))
	router.HandlerFunc(http.MethodGet, cityInfoUrl, apperror.Middleware(h.GetBriefWeatherInfo))
	router.HandlerFunc(http.MethodGet, cityDateInfoURL, apperror.Middleware(h.GetCityTimeInfo))

	router.HandlerFunc(http.MethodGet, "/doc/:any", swaggerHandler)
}

func swaggerHandler(res http.ResponseWriter, req *http.Request) {
	httpSwagger.WrapHandler(res, req)
}

// GetAvailableCities godoc
// @Summary      Available cities list
// @Description  get cities
// @Tags         Weather
// @Accept       json
// @Produce      json
// @Success      200  {array}   []cityClient.CityData
// @Router       /cities [get]
func (h *handler) GetAvailableCities(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET CITIES")
	w.Header().Set("Content-Type", "application/json")

	citiesData, err := h.cityService.FindAll(r.Context())
	if err != nil {
		return err
	}

	sort.Slice(citiesData, func(i, j int) bool {
		return citiesData[i].Name < citiesData[j].Name
	})

	h.logger.Debug("marshal cities")
	citiesBytes, err := json.Marshal(citiesData)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(citiesBytes)

	return nil
}

// GetBriefWeatherInfo godoc
// @Summary      City brief weather info and dates with more details
// @Description  Get brief weather info for city
// @Tags         Weather
// @Accept       json
// @Produce      json
// @Param        city    path     string  true  "weather info for city"  "City name"
// @Success      200  {array}    weatherClient.BriefWeatherCity
// @Router       /cities/{city} [get]
func (h *handler) GetBriefWeatherInfo(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET BRIEF WEATHER INFO FOR CITY")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get city from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	cityName := params.ByName("city")

	briefInfo, err := h.weatherService.FindBriefInfo(r.Context(), cityName)
	if err != nil {
		return err
	}

	briefInfo.AvgTemp = math.Round(briefInfo.AvgTemp*100) / 100

	h.logger.Debug("marshal api brief info")
	briefInfoBytes, err := json.Marshal(briefInfo)
	if err != nil {
		return fmt.Errorf("failed to marshall api brief info. error: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(briefInfoBytes)
	return nil

}

// GetCityTimeInfo godoc
// @Summary      City detail weather info for date
// @Description  Get city detailed weather by date
// @Tags         Weather
// @Accept       json
// @Produce      json
// @Param        city    path     string  true  "weather info for city"  "City name"
// @Param        date    path     string  true  "date expected 2006-01-02 15:04:05 or 2006-01-02T15:04:05Z format"  "Date"
// @Success      200  {array}    string
// @Router       /cities/{city}/{date} [get]
func (h *handler) GetCityTimeInfo(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET DETAILED WEATHER INFO FOR CITY ON DATE")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get city and date from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	cityName := params.ByName("city")
	dateString := params.ByName("date")

	date, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		date, err = time.Parse("2006-01-02 15:04:05", dateString)
		if err != nil {
			return fmt.Errorf("failed to parse date. expected: 2006-01-02 15:04:05 or 2006-01-02T15:04:05Z format. error: %w", err)
		}
	}

	weatherByCityAndDate, err := h.weatherService.FindInfoByCityAndDate(r.Context(), cityName, date)
	if err != nil {
		return err
	}

	h.logger.Debug("marshal api brief info")
	if err != nil {
		return fmt.Errorf("failed to marshall api brief info. error: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(weatherByCityAndDate))

	return nil
}
