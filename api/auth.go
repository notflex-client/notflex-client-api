package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OldPassword string `json:"oldPassword" validate:"required"`
		NewPassword string `json:"newPassword" validate:"required,min=8,max=50"`
		RePassword  string `json:"rePassword" validate:"required"`
	}
	logParams := []any{"handler", "ChangePassword"}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	if body.NewPassword != body.RePassword {
		HandleResponseError(w, r, NewValidationError(map[string]string{"rePassword": helpers.Translate(r.Context(), "PasswordMismatch")}))
		return
	}

	for _, pattern := range []string{`[a-z]`, `[A-Z]`, `[0-9]`, `[^a-zA-Z0-9]`} {
		matched, _ := regexp.MatchString(pattern, body.NewPassword)
		if !matched {
			HandleResponseError(w, r, NewValidationError(map[string]string{"newPassword": helpers.Translate(r.Context(), "StrongPassword")}))
			return
		}
	}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.OldPassword)); err != nil {
		HandleResponseError(w, r, NewValidationError(map[string]string{"oldPassword": helpers.Translate(r.Context(), "WrongOldPassword")}))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("hashing password", err, logParams...))
		return
	}

	if err := database.DB.WithContext(r.Context()).Model(&models.User{}).Where("id = ?", user.ID).Update("password_hash", string(hash)).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating password", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	logParams := []any{"handler", "Login"}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	var user models.User
	err := database.DB.WithContext(r.Context()).
		Where("email = ? AND is_active = TRUE", strings.ToLower(body.Email)).
		First(&user).Error
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidCredentials", logParams...))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidCredentials", logParams...))
		return
	}

	token := models.UserToken{UserAgent: r.UserAgent(), UserID: user.ID}
	if err := database.DB.WithContext(r.Context()).Create(&token).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating user token", err, logParams...))
		return
	}

	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"token": token.ID, "user": user})
}
