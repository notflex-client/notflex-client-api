package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func ListBanner(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListBanner"}

	banners := make([]models.Banner, 0)
	err := database.DB.WithContext(r.Context()).
		Where("is_active = TRUE").
		Preload("Movie").
		Order("position ASC, created_at DESC").
		Find(&banners).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing banners", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(banners)
}

func AdminListBanner(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminListBanner"}

	banners := make([]models.Banner, 0)
	err := database.DB.WithContext(r.Context()).
		Preload("Movie").
		Order("position ASC, created_at DESC").
		Find(&banners).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing banners", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(banners)
}

func AdminCreateBanner(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateBanner"}

	var body struct {
		MovieID     *string `json:"movie_id"`
		Title       string  `json:"title" validate:"required,max=255"`
		Description string  `json:"description" validate:"max=2000"`
		ImageURL    string  `json:"image_url" validate:"required,max=1000"`
		LinkURL     string  `json:"link_url" validate:"max=1000"`
		Position    int     `json:"position"`
		IsActive    bool    `json:"is_active"`
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

	banner := models.Banner{
		MovieID:     body.MovieID,
		Title:       body.Title,
		Description: body.Description,
		ImageURL:    body.ImageURL,
		LinkURL:     body.LinkURL,
		Position:    body.Position,
		IsActive:    body.IsActive,
	}
	err = database.DB.WithContext(r.Context()).Create(&banner).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating banner", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(banner)
}

func AdminUpdateBanner(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminUpdateBanner"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	var body struct {
		MovieID     *string `json:"movie_id"`
		Title       string  `json:"title" validate:"required,max=255"`
		Description string  `json:"description" validate:"max=2000"`
		ImageURL    string  `json:"image_url" validate:"required,max=1000"`
		LinkURL     string  `json:"link_url" validate:"max=1000"`
		Position    int     `json:"position"`
		IsActive    bool    `json:"is_active"`
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

	var banner models.Banner
	err = database.DB.WithContext(r.Context()).Where("id = ?", idParam).First(&banner).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("BannerNotFound", logParams...))
		return
	}

	banner.MovieID = body.MovieID
	banner.Title = body.Title
	banner.Description = body.Description
	banner.ImageURL = body.ImageURL
	banner.LinkURL = body.LinkURL
	banner.Position = body.Position
	banner.IsActive = body.IsActive
	err = database.DB.WithContext(r.Context()).Save(&banner).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating banner", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(banner)
}

func AdminDeleteBanner(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteBanner"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	err := database.DB.WithContext(r.Context()).Where("id = ?", idParam).Delete(&models.Banner{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting banner", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
