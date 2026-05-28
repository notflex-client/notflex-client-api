package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/models"
)

func GetSimilarMovies(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "GetSimilarMovies"}

	movieID := chi.URLParam(r, "id")
	logParams = append(logParams, "movieID", movieID)

	// Verify movie exists
	var movie models.Movie
	if err := database.DB.WithContext(r.Context()).
		Where("id = ?", movieID).First(&movie).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("MovieNotFound", logParams...))
		return
	}

	var sims []models.MovieSimilarity
	err := database.DB.WithContext(r.Context()).
		Where("movie_id = ?", movieID).
		Preload("SimilarMovie.Genres").
		Preload("SimilarMovie.Tags").
		Order("rank ASC").
		Limit(12).
		Find(&sims).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing similar movies", err, logParams...))
		return
	}

	movies := make([]models.Movie, 0, len(sims))
	for _, s := range sims {
		if s.SimilarMovie != nil {
			movies = append(movies, *s.SimilarMovie)
		}
	}

	// Fallback: if similarity table empty, return genre-based
	if len(movies) == 0 {
		database.DB.WithContext(r.Context()).
			Joins("JOIN movie_genres mg ON mg.movie_id = movies.id").
			Where("mg.genre_id IN (SELECT genre_id FROM movie_genres WHERE movie_id = ?)", movieID).
			Where("movies.id != ?", movieID).
			Preload("Genres").Preload("Tags").
			Distinct("movies.*").
			Order("movies.avg_rating DESC").
			Limit(12).
			Find(&movies)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"items":     movies,
		"itemCount": len(movies),
	})
}
