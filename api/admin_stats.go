package api

import (
	"encoding/json"
	"net/http"
	"time"

	"notflex_client_api/common/database"
	"notflex_client_api/models"
)

func AdminGetStats(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminGetStats"}

	var totalMovies int64
	err := database.DB.WithContext(r.Context()).Model(&models.Movie{}).Count(&totalMovies).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting movies", err, logParams...))
		return
	}

	var totalUsers int64
	err = database.DB.WithContext(r.Context()).Model(&models.User{}).Count(&totalUsers).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting users", err, logParams...))
		return
	}

	var activeSubscriptions int64
	err = database.DB.WithContext(r.Context()).
		Model(&models.UserSubscription{}).
		Where("status = ? AND end_date > ?", "active", time.Now()).
		Count(&activeSubscriptions).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting subscriptions", err, logParams...))
		return
	}

	monthStart := time.Now().AddDate(0, 0, -30)
	var revenue float64
	err = database.DB.WithContext(r.Context()).
		Model(&models.Payment{}).
		Where("status = ? AND created_at >= ?", "success", monthStart).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&revenue).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("summing revenue", err, logParams...))
		return
	}

	recentPayments := make([]models.Payment, 0, 5)
	err = database.DB.WithContext(r.Context()).
		Preload("Subscription.Plan").
		Order("created_at DESC").
		Limit(5).
		Find(&recentPayments).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing recent payments", err, logParams...))
		return
	}

	recentUsers := make([]models.User, 0, 5)
	err = database.DB.WithContext(r.Context()).
		Order("created_at DESC").
		Limit(5).
		Find(&recentUsers).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing recent users", err, logParams...))
		return
	}
	for i := range recentUsers {
		recentUsers[i].PasswordHash = ""
	}

	json.NewEncoder(w).Encode(map[string]any{
		"total_movies":         totalMovies,
		"total_users":          totalUsers,
		"active_subscriptions": activeSubscriptions,
		"revenue_last_30_days": revenue,
		"recent_payments":      recentPayments,
		"recent_users":         recentUsers,
	})
}
