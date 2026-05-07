package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func ListSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	var plans []models.SubscriptionPlan
	if err := database.DB.WithContext(r.Context()).Where("is_active = TRUE").Order("price ASC").Find(&plans).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing subscription plans", err, "handler", "ListSubscriptionPlans"))
		return
	}
	json.NewEncoder(w).Encode(plans)
}

func GetMySubscription(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	subscription, ok, err := findActiveSubscription(r, user.ID)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("loading subscription", err, "handler", "GetMySubscription", "userID", user.ID))
		return
	}
	if !ok {
		json.NewEncoder(w).Encode(map[string]any{"subscription": nil, "status": "free"})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"subscription": subscription, "status": subscription.Status})
}

func CheckoutSubscription(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "CheckoutSubscription", "userID", user.ID}

	var body struct {
		PlanID        int    `json:"plan_id" validate:"required"`
		PaymentMethod string `json:"payment_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}
	if body.PaymentMethod == "" {
		body.PaymentMethod = "mock-card"
	}

	var plan models.SubscriptionPlan
	if err := database.DB.WithContext(r.Context()).First(&plan, "id = ? AND is_active = TRUE", body.PlanID).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("SubscriptionPlanNotFound", logParams...))
		return
	}

	now := time.Now()
	subscription := models.UserSubscription{
		UserID:    user.ID,
		PlanID:    plan.ID,
		StartDate: now,
		EndDate:   now.AddDate(0, 0, plan.DurationDays),
		Status:    "active",
	}

	err := database.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.UserSubscription{}).Where("user_id = ? AND status = ?", user.ID, "active").Update("status", "cancelled").Error; err != nil {
			return err
		}
		if err := tx.Create(&subscription).Error; err != nil {
			return err
		}
		payment := models.Payment{
			UserID:         user.ID,
			SubscriptionID: &subscription.ID,
			Amount:         plan.Price,
			PaymentMethod:  body.PaymentMethod,
			Status:         "success",
			TransactionID:  "MOCK-" + uuid.NewString(),
		}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).Where("id = ?", user.ID).Update("role", "subscriber").Error
	})
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("checkout subscription", err, logParams...))
		return
	}

	database.DB.WithContext(r.Context()).Preload("Plan").First(&subscription, "id = ?", subscription.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subscription)
}

func ListPayments(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	var payments []models.Payment
	if err := database.DB.WithContext(r.Context()).Where("user_id = ?", user.ID).Preload("Subscription.Plan").Order("created_at DESC").Find(&payments).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing payments", err, "handler", "ListPayments", "userID", user.ID))
		return
	}
	json.NewEncoder(w).Encode(payments)
}

func userFromBearerToken(r *http.Request) (models.User, bool) {
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
	if err := database.DB.WithContext(r.Context()).Where("id = ? AND is_active = TRUE", token.UserID).First(&user).Error; err != nil {
		return models.User{}, false
	}
	return user, true
}

func findActiveSubscription(r *http.Request, userID string) (models.UserSubscription, bool, error) {
	var subscription models.UserSubscription
	err := database.DB.WithContext(r.Context()).
		Preload("Plan").
		Where("user_id = ? AND status = ? AND end_date > ?", userID, "active", time.Now()).
		Order("end_date DESC").
		First(&subscription).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.UserSubscription{}, false, nil
		}
		return models.UserSubscription{}, false, err
	}
	return subscription, true, nil
}
