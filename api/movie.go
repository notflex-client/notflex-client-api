package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"notflex_client_api/common/database"
	"notflex_client_api/models"
)

func MovieRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/genre/{genreId}", GetMoviesByGenre)
	return r
}

// GetMoviesByGenre tuân thủ 100% Query mẫu số 2 của SQL Script
// Lấy danh sách phim thuộc cùng 1 category, phân trang LIMIT OFFSET
func GetMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	genreIDStr := chi.URLParam(r, "genreId")
	genreID, err := strconv.Atoi(genreIDStr)
	if err != nil {
		http.Error(w, "Invalid genre ID", http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	limit := 20
	offset := (page - 1) * limit

	var movies []models.Movie
	
	// Thực thi câu lệnh JOIN lấy danh sách phim
	err = database.DB.
		Joins("JOIN movie_genres mg ON mg.movie_id = movies.id").
		Where("mg.genre_id = ?", genreID).
		Order("movies.avg_rating DESC").
		Limit(limit).
		Offset(offset).
		Find(&movies).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}
