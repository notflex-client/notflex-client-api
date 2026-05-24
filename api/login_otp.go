package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"notflex_client_api/common/database"
	"notflex_client_api/enum"
	"notflex_client_api/helpers"
	"notflex_client_api/mailer"
	"notflex_client_api/models"
	"notflex_client_api/templates"
)

func LoginOtpRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email" validate:"required,email,max=100"`
	}
	logParams := []any{"handler", "LoginOtpRequest"}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	var user models.User
	if err := database.DB.WithContext(r.Context()).
		Where("email = ? AND is_active = TRUE", strings.ToLower(body.Email)).
		First(&user).Error; err != nil {
		HandleResponseError(w, r, NewValidationError(map[string]string{"email": helpers.Translate(r.Context(), "UserNotFound")}))
		return
	}

	verifyCode := helpers.RandomString(6, "0123456789")
	request := models.LoginRequest{
		Email:            strings.ToLower(body.Email),
		ConfirmationCode: verifyCode,
		ExpireAt:         time.Now().Add(5 * time.Minute),
	}

	database.DB.WithContext(r.Context()).Where("email = ?", request.Email).Delete(&models.LoginRequest{})

	if err := database.DB.WithContext(r.Context()).Create(&request).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating login request", err, logParams...))
		return
	}

	locale, _ := r.Context().Value(enum.ContextKeyLocale).(string)
	err := mailer.Send(r.Context(), mailer.Message{
		To:      request.Email,
		Subject: helpers.Translate(r.Context(), "LoginCodeSubject"),
		Body:    templates.LoginOtp(locale, templates.LoginOtpParam{Code: verifyCode, Email: request.Email}),
	})
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("sending login otp email", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"id": request.ID})
}

func LoginOtpVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID         string `json:"id" validate:"required"`
		VerifyCode string `json:"verifyCode" validate:"required"`
	}
	logParams := []any{"handler", "LoginOtpVerify"}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	var request models.LoginRequest
	if err := database.DB.WithContext(r.Context()).Where("id = ?", body.ID).First(&request).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("LoginRequestNotFound", logParams...))
		return
	}
	if strings.TrimSpace(body.VerifyCode) != request.ConfirmationCode {
		HandleResponseError(w, r, NewValidationError(map[string]string{"verifyCode": helpers.Translate(r.Context(), "WrongConfirmationCode")}))
		return
	}
	if time.Now().After(request.ExpireAt) {
		HandleResponseError(w, r, NewValidationError(map[string]string{"verifyCode": helpers.Translate(r.Context(), "ExpiredConfirmationCode")}))
		return
	}

	var user models.User
	if err := database.DB.WithContext(r.Context()).
		Where("email = ? AND is_active = TRUE", request.Email).
		First(&user).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("UserNotFound", logParams...))
		return
	}

	token := models.UserToken{UserAgent: r.UserAgent(), UserID: user.ID}
	if err := database.DB.WithContext(r.Context()).Create(&token).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating user token", err, logParams...))
		return
	}

	database.DB.WithContext(r.Context()).Delete(&request)

	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"token": token.ID, "user": user})
}
