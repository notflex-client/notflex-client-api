package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

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
