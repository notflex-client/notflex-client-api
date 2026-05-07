package middlewares

import (
	"context"
	"net/http"
	"strings"

	"notflex_client_api/api"
	"notflex_client_api/common/database"
	"notflex_client_api/enum"
	"notflex_client_api/models"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromBearerToken(r)
		if !ok {
			api.HandleResponseError(w, r, api.NewUnauthorizedError())
			return
		}

		ctx := context.WithValue(r.Context(), enum.ContextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(enum.ContextKeyUser).(models.User)
			if !ok || !allowed[user.Role] {
				api.HandleResponseError(w, r, api.NewForbiddenError())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserFromBearerToken(r *http.Request) (models.User, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return models.User{}, false
	}

	tokenID := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenID == authHeader {
		return models.User{}, false
	}

	var token models.UserToken
	if err := database.DB.WithContext(r.Context()).Where("id = ?", tokenID).First(&token).Error; err != nil {
		return models.User{}, false
	}

	var user models.User
	if err := database.DB.WithContext(r.Context()).
		Where("id = ? AND is_active = TRUE", token.UserID).
		First(&user).Error; err != nil {
		return models.User{}, false
	}

	return user, true
}
