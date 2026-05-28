package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func AdminListEpisode(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminListEpisode"}

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	episodes := make([]models.Episode, 0)
	err := database.DB.WithContext(r.Context()).
		Where("movie_id = ?", movieIDParam).
		Order("season_number ASC, episode_number ASC").
		Find(&episodes).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing episodes", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"items": episodes, "itemCount": len(episodes)})
}

func AdminCreateEpisode(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateEpisode"}

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	var body struct {
		SeasonNumber  int     `json:"season_number" validate:"required,min=1"`
		EpisodeNumber int     `json:"episode_number" validate:"required,min=1"`
		Title         string  `json:"title" validate:"required,max=255"`
		Description   *string `json:"description"`
		VideoURL      *string `json:"video_url"`
		DurationMins  *int    `json:"duration_mins"`
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

	episode := models.Episode{
		MovieID:       movieIDParam,
		SeasonNumber:  body.SeasonNumber,
		EpisodeNumber: body.EpisodeNumber,
		Title:         body.Title,
		Description:   body.Description,
		VideoURL:      body.VideoURL,
		DurationMins:  body.DurationMins,
	}
	err = database.DB.WithContext(r.Context()).Create(&episode).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating episode", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(episode)
}

func AdminUpdateEpisode(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminUpdateEpisode"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	var body struct {
		SeasonNumber  int     `json:"season_number" validate:"required,min=1"`
		EpisodeNumber int     `json:"episode_number" validate:"required,min=1"`
		Title         string  `json:"title" validate:"required,max=255"`
		Description   *string `json:"description"`
		VideoURL      *string `json:"video_url"`
		DurationMins  *int    `json:"duration_mins"`
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

	var episode models.Episode
	err = database.DB.WithContext(r.Context()).Where("id = ?", idParam).First(&episode).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("EpisodeNotFound", logParams...))
		return
	}

	episode.SeasonNumber = body.SeasonNumber
	episode.EpisodeNumber = body.EpisodeNumber
	episode.Title = body.Title
	episode.Description = body.Description
	episode.VideoURL = body.VideoURL
	episode.DurationMins = body.DurationMins

	err = database.DB.WithContext(r.Context()).Save(&episode).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating episode", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(episode)
}

func AdminDeleteEpisode(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteEpisode"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	err := database.DB.WithContext(r.Context()).Where("id = ?", idParam).Delete(&models.Episode{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting episode", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
