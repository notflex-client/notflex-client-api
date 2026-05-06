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

	"gorm.io/gorm"
)

func ListGenre(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListGenre"}

	var genres []models.Genre
	if err := database.DB.WithContext(r.Context()).Order("name").Find(&genres).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing genres", err, logParams...))
		return
	}
	json.NewEncoder(w).Encode(genres)
}

func ListTag(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListTag"}

	var tags []models.Tag
	if err := database.DB.WithContext(r.Context()).Order("name").Find(&tags).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing tags", err, logParams...))
		return
	}
	json.NewEncoder(w).Encode(tags)
}

func ListMovie(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListMovie"}
	pageParam := strings.TrimSpace(r.URL.Query().Get("page"))
	pageSizeParam := strings.TrimSpace(r.URL.Query().Get("pageSize"))
	limitParam := strings.TrimSpace(r.URL.Query().Get("limit"))
	genreIDParam := strings.TrimSpace(r.URL.Query().Get("genre_id"))
	keywordParam := strings.TrimSpace(r.URL.Query().Get("keyword"))
	typeParam := strings.TrimSpace(r.URL.Query().Get("type"))
	tagParam := strings.TrimSpace(r.URL.Query().Get("tag"))
	sortParam := strings.TrimSpace(r.URL.Query().Get("sort"))

	if pageParam == "" {
		pageParam = "1"
	}
	if pageSizeParam == "" {
		pageSizeParam = limitParam
	}
	page := helpers.StringToInt64(pageParam, 1)
	pageSize := helpers.StringToInt64(pageSizeParam, 20)
	if pageSize < 1 || pageSize > 50 {
		HandleResponseError(w, r, NewBadRequestError("InvalidPageSize", logParams...))
		return
	}

	query := database.DB.WithContext(r.Context()).Model(&models.Movie{}).Preload("Genres").Preload("Tags")
	if typeParam != "" {
		query = query.Where("movies.type = ?", typeParam)
	}
	if genreIDParam != "" {
		query = query.Joins("JOIN movie_genres mg ON mg.movie_id = movies.id").Where("mg.genre_id = ?", genreIDParam)
	}
	if tagParam != "" {
		query = query.Joins("JOIN movie_tags mt ON mt.movie_id = movies.id").
			Joins("JOIN tags t ON t.id = mt.tag_id").
			Where("t.slug = ? OR LOWER(t.name) = LOWER(?)", tagParam, tagParam)
	}
	if keywordParam != "" {
		query = query.Where("LOWER(movies.title) LIKE ?", "%"+strings.ToLower(keywordParam)+"%")
	}

	itemCount := int64(0)
	err := query.Count(&itemCount).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting movies", err, logParams...))
		return
	}

	order := "movies.avg_rating DESC"
	switch sortParam {
	case "new":
		order = "movies.created_at DESC"
	case "top", "rating":
		order = "movies.avg_rating DESC"
	}

	offset := (page - 1) * pageSize
	movies := make([]models.Movie, 0, pageSize)
	err = query.Order(order).Limit(int(pageSize)).Offset(int(offset)).Find(&movies).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing movies", err, logParams...))
		return
	}

	pageCount := (itemCount + pageSize - 1) / pageSize
	json.NewEncoder(w).Encode(map[string]any{
		"items":     movies,
		"page":      page,
		"itemCount": itemCount,
		"pageCount": pageCount,
	})
}

func GetMovie(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "GetMovie"}

	idParam := chi.URLParam(r, "id")

	var movie models.Movie
	if err := database.DB.WithContext(r.Context()).
		Preload("Genres").
		Preload("Tags").
		Preload("Episodes", func(db *gorm.DB) *gorm.DB {
			return db.Order("season_number ASC, episode_number ASC")
		}).
		Where("id = ?", idParam).
		First(&movie).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("MovieNotFound", logParams...))
		return
	}

	json.NewEncoder(w).Encode(movie)
}

func GetMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "GetMoviesByGenre"}

	genreIDParam := chi.URLParam(r, "genreId")
	if _, err := strconv.Atoi(genreIDParam); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidGenreID", logParams...))
		return
	}

	page := helpers.StringToInt64(r.URL.Query().Get("page"), 1)
	pageSize := int64(20)
	offset := (page - 1) * pageSize

	var movies []models.Movie
	if err := database.DB.WithContext(r.Context()).
		Preload("Genres").
		Joins("JOIN movie_genres mg ON mg.movie_id = movies.id").
		Where("mg.genre_id = ?", genreIDParam).
		Order("movies.avg_rating DESC").
		Limit(int(pageSize)).Offset(int(offset)).
		Find(&movies).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing movies by genre", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(movies)
}

func CreateWatchHistory(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "CreateWatchHistory", "userID", user.ID}

	var body struct {
		MovieID       string `json:"movieId" validate:"required"`
		WatchDuration int    `json:"watchDuration"`
		IsCompleted   bool   `json:"isCompleted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	history := models.WatchHistory{
		UserID:        user.ID,
		MovieID:       body.MovieID,
		WatchDuration: body.WatchDuration,
		IsCompleted:   body.IsCompleted,
	}
	if err := database.DB.WithContext(r.Context()).Create(&history).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating watch history", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(history)
}

func ListWatchHistory(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "ListWatchHistory", "userID", user.ID}

	var history []models.WatchHistory
	if err := database.DB.WithContext(r.Context()).
		Where("user_id = ?", user.ID).
		Preload("Movie").
		Order("watched_at DESC").
		Find(&history).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing watch history", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(history)
}

func CreateRating(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "CreateRating", "userID", user.ID}

	var body struct {
		MovieID string `json:"movieId" validate:"required"`
		Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	rating := models.MovieRating{UserID: user.ID, MovieID: body.MovieID, Rating: body.Rating}
	err := database.DB.WithContext(r.Context()).
		Where(models.MovieRating{UserID: user.ID, MovieID: body.MovieID}).
		Assign(models.MovieRating{Rating: body.Rating}).
		FirstOrCreate(&rating).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("upserting rating", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(rating)
}
