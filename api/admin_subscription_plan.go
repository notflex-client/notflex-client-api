package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func AdminListSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminListSubscriptionPlan"}

	var plans []models.SubscriptionPlan
	err := database.DB.WithContext(r.Context()).Order("price ASC").Find(&plans).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing plans", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(plans)
}

func AdminCreateSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminCreateSubscriptionPlan"}

	var body struct {
		Name         string  `json:"name" validate:"required,max=50"`
		Price        float64 `json:"price" validate:"required,min=0"`
		DurationDays int     `json:"duration_days" validate:"required,min=1"`
		Description  *string `json:"description"`
		IsActive     bool    `json:"is_active"`
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

	plan := models.SubscriptionPlan{
		Name:         body.Name,
		Price:        body.Price,
		DurationDays: body.DurationDays,
		Description:  body.Description,
		IsActive:     body.IsActive,
	}
	err = database.DB.WithContext(r.Context()).Create(&plan).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating plan", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(plan)
}

func AdminUpdateSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminUpdateSubscriptionPlan"}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidID", logParams...))
		return
	}

	var body struct {
		Name         string  `json:"name" validate:"required,max=50"`
		Price        float64 `json:"price" validate:"required,min=0"`
		DurationDays int     `json:"duration_days" validate:"required,min=1"`
		Description  *string `json:"description"`
		IsActive     bool    `json:"is_active"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	var plan models.SubscriptionPlan
	err = database.DB.WithContext(r.Context()).Where("id = ?", id).First(&plan).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("PlanNotFound", logParams...))
		return
	}

	plan.Name = body.Name
	plan.Price = body.Price
	plan.DurationDays = body.DurationDays
	plan.Description = body.Description
	plan.IsActive = body.IsActive
	err = database.DB.WithContext(r.Context()).Save(&plan).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating plan", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(plan)
}

func AdminDeleteSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "AdminDeleteSubscriptionPlan"}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidID", logParams...))
		return
	}

	err = database.DB.WithContext(r.Context()).Where("id = ?", id).Delete(&models.SubscriptionPlan{}).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting plan", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
}
