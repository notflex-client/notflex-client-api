package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"notflex_client_api/common/database"
	"notflex_client_api/enum"
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

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "UpdateProfile"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	var body struct {
		FullName  string  `json:"full_name" validate:"required,min=1,max=100"`
		AvatarURL *string `json:"avatar_url"`
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

	updates := map[string]any{
		"full_name":  strings.TrimSpace(body.FullName),
		"avatar_url": body.AvatarURL,
	}
	err = database.DB.WithContext(r.Context()).Model(&user).Updates(updates).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating profile", err, logParams...))
		return
	}

	err = database.DB.WithContext(r.Context()).Where("id = ?", user.ID).First(&user).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("reloading user", err, logParams...))
		return
	}
	user.PasswordHash = ""
	json.NewEncoder(w).Encode(user)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ChangePassword"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	var body struct {
		CurrentPassword string `json:"currentPassword" validate:"required"`
		NewPassword     string `json:"newPassword" validate:"required,min=8,max=50"`
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

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.CurrentPassword))
	if err != nil {
		HandleResponseError(w, r, NewValidationError(map[string]string{"currentPassword": helpers.Translate(r.Context(), "WrongCurrentPassword")}))
		return
	}

	if body.CurrentPassword == body.NewPassword {
		HandleResponseError(w, r, NewValidationError(map[string]string{"newPassword": helpers.Translate(r.Context(), "SamePassword")}))
		return
	}

	for _, pattern := range []string{`[a-z]`, `[A-Z]`, `[0-9]`, `[^a-zA-Z0-9]`} {
		matched, _ := regexp.MatchString(pattern, body.NewPassword)
		if !matched {
			HandleResponseError(w, r, NewValidationError(map[string]string{"newPassword": helpers.Translate(r.Context(), "StrongPassword")}))
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("hashing password", err, logParams...))
		return
	}

	err = database.DB.WithContext(r.Context()).Model(&user).Update("password_hash", string(hash)).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating password", err, logParams...))
		return
	}

	tokenID, _ := r.Context().Value(enum.ContextKeyToken).(string)
	if tokenID != "" {
		database.DB.WithContext(r.Context()).Where("user_id = ? AND id <> ?", user.ID, tokenID).Delete(&models.UserToken{})
	}

	json.NewEncoder(w).Encode(map[string]any{"success": true})
}
