package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

type MovieInput struct {
	Title        string  `json:"title"         validate:"required,max=255"`
	Type         string  `json:"type"          validate:"required,oneof=movie series"`
	Description  *string `json:"description"`
	PosterURL    *string `json:"poster_url"`
	BannerURL    *string `json:"banner_url"`
	VideoURL     *string `json:"video_url"`
	TrailerURL   *string `json:"trailer_url"`
	DurationMins *int    `json:"duration_mins"`
	ReleaseYear  *int16  `json:"release_year"`
	Rating       *string `json:"rating"`
	IsPremium    bool    `json:"is_premium"`
	GenreIDs     []int   `json:"genre_ids"`
	TagIDs       []int   `json:"tag_ids"`
}

func AdminCreateMovie(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateMovie"}

	var input MovieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if input.Type == "" {
		input.Type = "movie"
	}
	if errs := helpers.ValidateStruct(input); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	movie := models.Movie{
		Title:        input.Title,
		Type:         input.Type,
		Description:  input.Description,
		PosterURL:    input.PosterURL,
		BannerURL:    input.BannerURL,
		VideoURL:     input.VideoURL,
		TrailerURL:   input.TrailerURL,
		DurationMins: input.DurationMins,
		ReleaseYear:  input.ReleaseYear,
		Rating:       input.Rating,
		IsPremium:    input.IsPremium,
	}
	if err := database.DB.WithContext(r.Context()).Create(&movie).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating movie", err, logParams...))
		return
	}
	adminAssignGenres(r, movie.ID, input.GenreIDs)
	adminAssignTags(r, movie.ID, input.TagIDs)
	database.DB.WithContext(r.Context()).Preload("Genres").Preload("Tags").First(&movie, "id = ?", movie.ID)

	slog.Info("movie created", "id", movie.ID, "title", movie.Title)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie) //nolint:errcheck
}

func AdminUpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logParams := []any{"handler", "AdminUpdateMovie", "id", id}

	var movie models.Movie
	if err := database.DB.WithContext(r.Context()).First(&movie, "id = ?", id).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("MovieNotFound", logParams...))
		return
	}

	var input MovieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(input); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	movie.Title = input.Title
	movie.Type = input.Type
	movie.Description = input.Description
	movie.PosterURL = input.PosterURL
	movie.BannerURL = input.BannerURL
	movie.VideoURL = input.VideoURL
	movie.TrailerURL = input.TrailerURL
	movie.DurationMins = input.DurationMins
	movie.ReleaseYear = input.ReleaseYear
	movie.Rating = input.Rating
	movie.IsPremium = input.IsPremium

	if err := database.DB.WithContext(r.Context()).Save(&movie).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating movie", err, logParams...))
		return
	}
	adminAssignGenres(r, movie.ID, input.GenreIDs)
	adminAssignTags(r, movie.ID, input.TagIDs)
	database.DB.WithContext(r.Context()).Preload("Genres").Preload("Tags").First(&movie, "id = ?", movie.ID)
	json.NewEncoder(w).Encode(movie) //nolint:errcheck
}

func AdminDeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logParams := []any{"handler", "AdminDeleteMovie", "id", id}

	if err := database.DB.WithContext(r.Context()).Delete(&models.Movie{}, "id = ?", id).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting movie", err, logParams...))
		return
	}
	slog.Info("movie deleted", "id", id)
	json.NewEncoder(w).Encode(map[string]bool{"success": true}) //nolint:errcheck
}

func adminAssignGenres(r *http.Request, movieID string, genreIDs []int) {
	var genres []models.Genre
	database.DB.WithContext(r.Context()).Where("id IN ?", genreIDs).Find(&genres)
	database.DB.WithContext(r.Context()).Model(&models.Movie{ID: movieID}).Association("Genres").Replace(genres)
}

func adminAssignTags(r *http.Request, movieID string, tagIDs []int) {
	var tags []models.Tag
	database.DB.WithContext(r.Context()).Where("id IN ?", tagIDs).Find(&tags)
	database.DB.WithContext(r.Context()).Model(&models.Movie{ID: movieID}).Association("Tags").Replace(tags)
}
