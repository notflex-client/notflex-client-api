package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func GenreRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/",        ListGenres)
	r.Post("/",       CreateGenre)
	r.Delete("/{id}", DeleteGenre)
	return r
}

func ListGenres(w http.ResponseWriter, r *http.Request) {
	lp := []any{"handler", "ListGenres"}

	var genres []models.Genre
	if err := database.DB.Order("name ASC").Find(&genres).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("list genres", err, lp...))
		return
	}
	render.JSON(w, r, render.M{"data": genres})
}

func CreateGenre(w http.ResponseWriter, r *http.Request) {
	lp := []any{"handler", "CreateGenre"}

	var body struct {
		Name string `json:"name" validate:"required,max=50"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", lp...))
		return
	}
	body.Name = strings.TrimSpace(body.Name)

	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	genre := models.Genre{Name: body.Name}
	if err := database.DB.Create(&genre).Error; err != nil {
		if strings.Contains(err.Error(), "unique") {
			HandleResponseError(w, r, NewValidationError(map[string]string{"name": "Genre already exists"}))
			return
		}
		HandleResponseError(w, r, NewInternalServerError("create genre", err, lp...))
		return
	}
	slog.Info("genre created", "id", genre.ID, "name", genre.Name)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, genre)
}

func DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	lp := []any{"handler", "DeleteGenre", "id", id}
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidID", lp...))
		return
	}
	if err := database.DB.Delete(&models.Genre{}, id).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("delete genre", err, lp...))
		return
	}
	slog.Info("genre deleted", "id", id)
	render.JSON(w, r, render.M{"success": true})
}
