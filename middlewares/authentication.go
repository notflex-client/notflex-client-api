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
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			api.HandleResponseError(w, r, api.NewUnauthorizedError())
			return
		}

		tokenID := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenID == authHeader {
			api.HandleResponseError(w, r, api.NewUnauthorizedError())
			return
		}

		var token models.UserToken
		if err := database.DB.WithContext(r.Context()).Where("id = ?", tokenID).First(&token).Error; err != nil {
			api.HandleResponseError(w, r, api.NewUnauthorizedError())
			return
		}

		var user models.User
		if err := database.DB.WithContext(r.Context()).
			Where("id = ? AND is_active = TRUE", token.UserID).
			First(&user).Error; err != nil {
			api.HandleResponseError(w, r, api.NewUnauthorizedError())
			return
		}

		ctx := context.WithValue(r.Context(), enum.ContextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
