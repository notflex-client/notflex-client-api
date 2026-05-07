package api

import (
	"encoding/json"
	"net/http"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func ListRecommendations(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "ListRecommendations", "userID", user.ID}

	movies := make([]models.Movie, 0, 12)
	query := database.DB.WithContext(r.Context()).Model(&models.Movie{}).Preload("Genres").Preload("Tags")

	var watchedMovieIDs []string
	database.DB.WithContext(r.Context()).Model(&models.WatchHistory{}).
		Where("user_id = ?", user.ID).
		Pluck("movie_id", &watchedMovieIDs)

	if len(watchedMovieIDs) > 0 {
		var genreIDs []int
		database.DB.WithContext(r.Context()).Table("movie_genres").
			Where("movie_id IN ?", watchedMovieIDs).
			Pluck("genre_id", &genreIDs)
		if len(genreIDs) > 0 {
			query = query.Joins("JOIN movie_genres mg ON mg.movie_id = movies.id").Where("mg.genre_id IN ?", genreIDs)
		}
		query = query.Where("movies.id NOT IN ?", watchedMovieIDs)
	}

	if err := query.Order("movies.avg_rating DESC, movies.created_at DESC").Limit(12).Find(&movies).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing recommendations", err, logParams...))
		return
	}

	if len(movies) == 0 {
		if err := database.DB.WithContext(r.Context()).Preload("Genres").Preload("Tags").Order("avg_rating DESC, created_at DESC").Limit(12).Find(&movies).Error; err != nil {
			HandleResponseError(w, r, NewInternalServerError("listing fallback recommendations", err, logParams...))
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"items":  movies,
		"source": "mock-content-based",
	})
}
