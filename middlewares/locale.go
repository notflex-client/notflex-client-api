package middlewares

import (
	"context"
	"net/http"

	"notflex_client_api/enum"
)

func LocaleHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := r.Header.Get("Accept-Language")
		if locale == "" {
			locale = "en"
		}
		ctx := context.WithValue(r.Context(), enum.ContextKeyLocale, locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
