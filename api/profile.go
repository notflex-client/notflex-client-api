package api

import (
	"encoding/json"
	"net/http"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func TransferProfile(w http.ResponseWriter, r *http.Request) {
	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}

	var body struct {
		ProfileID   string `json:"profileId" validate:"required"`
		TargetEmail string `json:"targetEmail" validate:"required,email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("invalid_request_body"))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	if body.TargetEmail == user.Email {
		HandleResponseError(w, r, NewBadRequestError("cannot_transfer_to_self"))
		return
	}

	var targetUser models.User
	if err := database.DB.WithContext(r.Context()).Where("email = ?", body.TargetEmail).First(&targetUser).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("target_user_not_found"))
		return
	}

	transfer := models.ProfileTransfer{
		FromUserID:  user.ID,
		ToUserID:    targetUser.ID,
		ProfileName: user.FullName,
		Status:      "pending",
	}

	if err := database.DB.WithContext(r.Context()).Create(&transfer).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("failed to create transfer", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"transfer": transfer,
	})
}
