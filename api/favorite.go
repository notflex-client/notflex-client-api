package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func ListFavorite(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListFavorite"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	pageParam := r.URL.Query().Get("page")
	pageSizeParam := r.URL.Query().Get("pageSize")
	page := helpers.StringToInt64(pageParam, 1)
	pageSize := helpers.StringToInt64(pageSizeParam, 20)
	if pageSize < 1 || pageSize > 50 {
		HandleResponseError(w, r, NewBadRequestError("InvalidPageSize", logParams...))
		return
	}

	query := database.DB.WithContext(r.Context()).
		Model(&models.Favorite{}).
		Where("user_id = ?", user.ID)

	itemCount := int64(0)
	err = query.Count(&itemCount).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting favorites", err, logParams...))
		return
	}

	offset := (page - 1) * pageSize
	favorites := make([]models.Favorite, 0, pageSize)
	err = query.
		Preload("Movie.Genres").
		Preload("Movie.Tags").
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&favorites).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing favorites", err, logParams...))
		return
	}

	pageCount := (itemCount + pageSize - 1) / pageSize
	json.NewEncoder(w).Encode(map[string]any{
		"items":     favorites,
		"page":      page,
		"itemCount": itemCount,
		"pageCount": pageCount,
	})
}

func CreateFavorite(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "CreateFavorite"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	var movie models.Movie
	err = database.DB.WithContext(r.Context()).Where("id = ?", movieIDParam).First(&movie).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("MovieNotFound", logParams...))
		return
	}

	favorite := models.Favorite{UserID: user.ID, MovieID: movieIDParam}
	err = database.DB.WithContext(r.Context()).
		Where("user_id = ? AND movie_id = ?", user.ID, movieIDParam).
		First(&models.Favorite{}).Error
	if err == nil {
		json.NewEncoder(w).Encode(favorite)
		return
	}
	if err != gorm.ErrRecordNotFound {
		HandleResponseError(w, r, NewInternalServerError("checking favorite", err, logParams...))
		return
	}

	err = database.DB.WithContext(r.Context()).Create(&favorite).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating favorite", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(favorite)
}

func DeleteFavorite(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "DeleteFavorite"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	movieIDParam := chi.URLParam(r, "movieId")
	logParams = append(logParams, "movieID", movieIDParam)

	err = database.DB.WithContext(r.Context()).
		Where("user_id = ? AND movie_id = ?", user.ID, movieIDParam).
		Delete(&models.Favorite{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting favorite", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
