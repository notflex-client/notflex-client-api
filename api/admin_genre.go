package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func AdminCreateGenre(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateGenre"}

	var body struct {
		Name string `json:"name" validate:"required,max=50"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	body.Name = strings.TrimSpace(body.Name)

	var count int64
	err = database.DB.WithContext(r.Context()).Model(&models.Genre{}).
		Where("LOWER(name) = LOWER(?)", body.Name).Count(&count).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("checking genre uniqueness", err, logParams...))
		return
	}
	if count > 0 {
		HandleResponseError(w, r, NewValidationError(map[string]string{"name": helpers.Translate(r.Context(), "GenreNameExisted")}))
		return
	}

	genre := models.Genre{Name: body.Name}
	err = database.DB.WithContext(r.Context()).Create(&genre).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating genre", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre)
}

func AdminUpdateGenre(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminUpdateGenre"}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidID", logParams...))
		return
	}

	var body struct {
		Name string `json:"name" validate:"required,max=50"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	var genre models.Genre
	err = database.DB.WithContext(r.Context()).Where("id = ?", id).First(&genre).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("GenreNotFound", logParams...))
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	var count int64
	err = database.DB.WithContext(r.Context()).Model(&models.Genre{}).
		Where("LOWER(name) = LOWER(?) AND id <> ?", body.Name, id).Count(&count).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("checking genre uniqueness", err, logParams...))
		return
	}
	if count > 0 {
		HandleResponseError(w, r, NewValidationError(map[string]string{"name": helpers.Translate(r.Context(), "GenreNameExisted")}))
		return
	}

	genre.Name = body.Name
	err = database.DB.WithContext(r.Context()).Save(&genre).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating genre", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(genre)
}

func AdminDeleteGenre(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteGenre"}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidID", logParams...))
		return
	}

	database.DB.WithContext(r.Context()).Exec("DELETE FROM movie_genres WHERE genre_id = ?", id)

	err = database.DB.WithContext(r.Context()).Where("id = ?", id).Delete(&models.Genre{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting genre", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
