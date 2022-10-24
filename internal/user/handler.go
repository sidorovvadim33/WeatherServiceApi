package user

import (
	"WeatherServiceAPI/internal/apperror"
	"WeatherServiceAPI/internal/handlers"
	"WeatherServiceAPI/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	usersURL    = "/api/users"
	userURL     = "/api/users/:uuid"
	usersFavURL = "/api/userfavs"
	userFavURL  = "/api/userfavs/:uuid"
)

type handler struct {
	Logger      *logging.Logger
	UserService Service
}

func NewHandler(logger *logging.Logger, userService Service) handlers.Handler {
	return &handler{
		Logger:      logger,
		UserService: userService,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, userURL, apperror.Middleware(h.GetUser))
	router.HandlerFunc(http.MethodGet, usersURL, apperror.Middleware(h.GetUserByEmailAndPassword))
	router.HandlerFunc(http.MethodPost, usersURL, apperror.Middleware(h.CreateUser))
	router.HandlerFunc(http.MethodPatch, userURL, apperror.Middleware(h.PartiallyUpdateUser))
	router.HandlerFunc(http.MethodDelete, userURL, apperror.Middleware(h.DeleteUser))

	router.HandlerFunc(http.MethodPost, userFavURL, apperror.Middleware(h.CreateFavourite))
	router.HandlerFunc(http.MethodDelete, userFavURL, apperror.Middleware(h.DeleteFromFavourites))
	router.HandlerFunc(http.MethodGet, usersFavURL, apperror.Middleware(h.GetUserFavourites))
}

// GetUser godoc
// @Summary      Get user by uuid
// @Description  Get user by uuid
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        uuid    path     string  true  "User uuid"
// @Success      200  {array}  User
// @Router       /users/{uuid} [get]
func (h *handler) GetUser(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("GET USER")
	w.Header().Set("Content-Type", "application/json")

	h.Logger.Debug("get uuid from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	user, err := h.UserService.GetOne(r.Context(), userUUID)
	if err != nil {
		return err
	}

	h.Logger.Debug("marshal user")
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshall user. error: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(userBytes)
	return nil
}

// GetUserByEmailAndPassword godoc
// @Summary      Get user by email and password
// @Description  Get user by email and password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        email    query     string  true  "User email"
// @Param        password    query     string  true  "User password"
// @Success      200  {array}  User
// @Router       /users [get]
func (h *handler) GetUserByEmailAndPassword(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("GET USER BY EMAIL AND PASSWORD")
	w.Header().Set("Content-Type", "application/json")

	h.Logger.Debug("get email and password from URL")
	email := r.URL.Query().Get("email")
	password := r.URL.Query().Get("password")
	if email == "" || password == "" {
		return fmt.Errorf("invalid query parameters email or password")
	}

	user, err := h.UserService.GetByEmailAndPassword(r.Context(), email, password)
	if err != nil {
		return err
	}

	h.Logger.Debug("marshal user")
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(userBytes)

	return nil
}

// GetUserFavourites godoc
// @Summary      Get user favourite cities
// @Description  Get user favourite cities by email and password
// @Tags         Users Favourite Cities
// @Accept       json
// @Produce      json
// @Param        email    query     string  true  "User email"
// @Param        password    query     string  true  "User password"
// @Success      200  {array}  []cityClient.CityData
// @Router       /userfavs [get]
func (h *handler) GetUserFavourites(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("GET USER FAVOURITE CITIES BY EMAIL AND PASSWORD")
	w.Header().Set("Content-Type", "application/json")

	h.Logger.Debug("get email and password from URL")
	email := r.URL.Query().Get("email")
	password := r.URL.Query().Get("password")
	if email == "" || password == "" {
		return fmt.Errorf("invalid query parameters email or password")
	}

	cities, err := h.UserService.GetFavourites(r.Context(), email, password)
	if err != nil {
		return err
	}

	if cities != nil {
		h.Logger.Debug("marshal cities data")
		userBytes, err := json.Marshal(cities)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		w.Write(userBytes)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("user favourites list is empty"))
	}

	return nil
}

// CreateUser godoc
// @Summary      Create new user
// @Description  Create new user by email and password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        crUser    body     CreateUserDTO  true  "New user"
// @Success      200  {array}  []cityClient.CityData
// @Router       /users [post]
func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("CREATE USER")
	w.Header().Set("Content-Type", "application/json")

	h.Logger.Debug("decode create user dto")
	var crUser CreateUserDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&crUser); err != nil {
		return fmt.Errorf("invalid JSON scheme. check swagger API")
	}

	userUUID, err := h.UserService.Create(r.Context(), crUser)
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("%s/%s", usersURL, userUUID))
	w.WriteHeader(http.StatusCreated)

	return nil
}

// CreateFavourite godoc
// @Summary      Create new user favourite city
// @Description  Create new user favourite city by email and password
// @Tags         Users Favourite Cities
// @Accept       json
// @Produce      json
// @Param        uuid    path     string  true  "User uuid"
// @Param        crUser    body      UserFavouriteCityDTO  true  "User with city"
// @Success      200  {array}  []cityClient.CityData
// @Router       /userfavs/{uid} [post]
func (h *handler) CreateFavourite(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("ADD CITY TO USER FAVOURITES")
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	h.Logger.Debug("decode user fav city dto")
	var userFavCity UserFavouriteCityDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&userFavCity); err != nil {
		return fmt.Errorf("invalid JSON scheme")
	}
	userFavCity.UUID = userUUID

	err := h.UserService.CreateFavourite(r.Context(), userFavCity, userFavCity.CityID)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}

// PartiallyUpdateUser godoc
// @Summary      Partially user update
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        uuid    path     string  true  "User uuid"
// @Param        crUser    body     UpdateUserDTO  true  "Updated user"
// @Success      204
// @Router       /users/{uuid} [patch]
func (h *handler) PartiallyUpdateUser(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("PARTIALLY UPDATE USER")
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	h.Logger.Debug("decode update user dto")
	var updUser UpdateUserDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&updUser); err != nil {
		return fmt.Errorf("invalid JSON scheme")
	}
	updUser.UUID = userUUID

	err := h.UserService.Update(r.Context(), updUser)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}

// DeleteUser godoc
// @Summary      Delete user by uuid param
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        uuid    path     string  true  "User uuid"
// @Success      204
// @Router       /users/{uuid} [delete]
func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("DELETE USER")
	w.Header().Set("Content-Type", "application/json")

	h.Logger.Debug("get uuid from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	err := h.UserService.Delete(r.Context(), userUUID)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}

// DeleteFromFavourites godoc
// @Summary      Delete city from user favourites
// @Tags         Users Favourite Cities
// @Accept       json
// @Produce      json
// @Param        uuid    path     string  true  "User uuid"
// @Param        crUser    body     UserFavouriteCityDTO  true  "User and city"
// @Success      204
// @Router       /userfavs/{uuid} [delete]
func (h *handler) DeleteFromFavourites(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("DELETE CITY FROM USER FAVOURITES")
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	h.Logger.Debug("decode user fav city dto")
	var userFavCity UserFavouriteCityDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&userFavCity); err != nil {
		return fmt.Errorf("invalid JSON scheme")
	}
	userFavCity.UUID = userUUID

	err := h.UserService.DeleteFavourite(r.Context(), userFavCity, userFavCity.CityID)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}
