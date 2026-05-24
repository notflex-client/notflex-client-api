package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func ListSubtitleByMovie(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListSubtitleByMovie"}

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	subtitles := make([]models.Subtitle, 0)
	err := database.DB.WithContext(r.Context()).
		Where("movie_id = ?", movieIDParam).
		Order("language ASC").
		Find(&subtitles).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing subtitles", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"items": subtitles, "itemCount": len(subtitles)})
}

func AdminCreateSubtitle(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateSubtitle"}

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	var body struct {
		Language    string `json:"language" validate:"required,max=20"`
		Label       string `json:"label" validate:"max=50"`
		SubtitleURL string `json:"subtitle_url" validate:"required,max=1000"`
		Format      string `json:"format" validate:"max=10"`
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

	var movie models.Movie
	err = database.DB.WithContext(r.Context()).Where("id = ?", movieIDParam).First(&movie).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("MovieNotFound", logParams...))
		return
	}

	if body.Format == "" {
		body.Format = "vtt"
	}

	subtitle := models.Subtitle{
		MovieID:     movieIDParam,
		Language:    body.Language,
		Label:       body.Label,
		SubtitleURL: body.SubtitleURL,
		Format:      body.Format,
	}
	err = database.DB.WithContext(r.Context()).Create(&subtitle).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating subtitle", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subtitle)
}

func AdminDeleteSubtitle(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteSubtitle"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	err := database.DB.WithContext(r.Context()).Where("id = ?", idParam).Delete(&models.Subtitle{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting subtitle", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
