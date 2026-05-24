package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func ListAudioTrackByMovie(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListAudioTrackByMovie"}

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	tracks := make([]models.AudioTrack, 0)
	err := database.DB.WithContext(r.Context()).
		Where("movie_id = ?", movieIDParam).
		Order("is_default DESC, language ASC").
		Find(&tracks).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing audio tracks", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"items": tracks, "itemCount": len(tracks)})
}

func AdminCreateAudioTrack(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateAudioTrack"}

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	var body struct {
		Language  string `json:"language" validate:"required,max=20"`
		Label     string `json:"label" validate:"max=50"`
		AudioURL  string `json:"audio_url" validate:"required,max=1000"`
		IsDefault bool   `json:"is_default"`
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

	if body.IsDefault {
		database.DB.WithContext(r.Context()).
			Model(&models.AudioTrack{}).
			Where("movie_id = ?", movieIDParam).
			Update("is_default", false)
	}

	track := models.AudioTrack{
		MovieID:   movieIDParam,
		Language:  body.Language,
		Label:     body.Label,
		AudioURL:  body.AudioURL,
		IsDefault: body.IsDefault,
	}
	err = database.DB.WithContext(r.Context()).Create(&track).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating audio track", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(track)
}

func AdminDeleteAudioTrack(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteAudioTrack"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	err := database.DB.WithContext(r.Context()).Where("id = ?", idParam).Delete(&models.AudioTrack{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting audio track", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
