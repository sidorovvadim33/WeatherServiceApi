package api

import (
	"WeatherServiceAPI/internal/api/cityClient"
	"WeatherServiceAPI/internal/api/weatherClient"
	"WeatherServiceAPI/internal/apperror"
	"WeatherServiceAPI/internal/handlers"
	"WeatherServiceAPI/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"math"
	"net/http"
	"sort"
	"time"
)

const (
	citiesUrl       = "/api/cities"
	cityInfoUrl     = "/api/cities/:cityClient"
	cityDateInfoURL = "/api/cities/:cityClient/:date"
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
}

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

func (h *handler) GetBriefWeatherInfo(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET BRIEF WEATHER INFO FOR CITY")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get cityClient from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	cityName := params.ByName("cityClient")

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

func (h *handler) GetCityTimeInfo(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("GET DETAILED WEATHER INFO FOR CITY ON DATE")
	w.Header().Set("Content-Type", "application/json")

	h.logger.Debug("get cityClient and date from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	cityName := params.ByName("cityClient")
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
