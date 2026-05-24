package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func AdminListUser(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminListUser"}

	pageParam := strings.TrimSpace(r.URL.Query().Get("page"))
	pageSizeParam := strings.TrimSpace(r.URL.Query().Get("pageSize"))
	keywordParam := strings.TrimSpace(r.URL.Query().Get("keyword"))
	roleParam := strings.TrimSpace(r.URL.Query().Get("role"))

	page := helpers.StringToInt64(pageParam, 1)
	pageSize := helpers.StringToInt64(pageSizeParam, 20)
	if pageSize < 1 || pageSize > 100 {
		HandleResponseError(w, r, NewBadRequestError("InvalidPageSize", logParams...))
		return
	}

	query := database.DB.WithContext(r.Context()).Model(&models.User{})
	if keywordParam != "" {
		like := "%" + strings.ToLower(keywordParam) + "%"
		query = query.Where("LOWER(email) LIKE ? OR LOWER(full_name) LIKE ?", like, like)
	}
	if roleParam != "" {
		query = query.Where("role = ?", roleParam)
	}

	itemCount := int64(0)
	err := query.Count(&itemCount).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting users", err, logParams...))
		return
	}

	offset := (page - 1) * pageSize
	users := make([]models.User, 0, pageSize)
	err = query.
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&users).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing users", err, logParams...))
		return
	}

	pageCount := (itemCount + pageSize - 1) / pageSize
	json.NewEncoder(w).Encode(map[string]any{
		"items":     users,
		"page":      page,
		"itemCount": itemCount,
		"pageCount": pageCount,
	})
}

func AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminUpdateUser"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	var body struct {
		FullName  string  `json:"full_name" validate:"max=100"`
		Role      string  `json:"role" validate:"oneof=guest subscriber admin"`
		IsActive  *bool   `json:"is_active"`
		AvatarURL *string `json:"avatar_url"`
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

	var user models.User
	err = database.DB.WithContext(r.Context()).Where("id = ?", idParam).First(&user).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("UserNotFound", logParams...))
		return
	}

	updates := map[string]any{
		"full_name": body.FullName,
		"role":      body.Role,
	}
	if body.IsActive != nil {
		updates["is_active"] = *body.IsActive
	}
	if body.AvatarURL != nil {
		updates["avatar_url"] = body.AvatarURL
	}

	err = database.DB.WithContext(r.Context()).Model(&user).Updates(updates).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating user", err, logParams...))
		return
	}

	err = database.DB.WithContext(r.Context()).Where("id = ?", idParam).First(&user).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("reloading user", err, logParams...))
		return
	}

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(user)
}

func AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteUser"}

	idParam := chi.URLParam(r, "id")
	logParams = append(logParams, "id", idParam)

	currentUser, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	if currentUser.ID == idParam {
		HandleResponseError(w, r, NewBadRequestError("CannotDeleteSelf", logParams...))
		return
	}

	err = database.DB.WithContext(r.Context()).
		Model(&models.User{}).
		Where("id = ?", idParam).
		Update("is_active", false).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deactivating user", err, logParams...))
		return
	}

	database.DB.WithContext(r.Context()).Where("user_id = ?", idParam).Delete(&models.UserToken{})

	json.NewEncoder(w).Encode(true)
}
