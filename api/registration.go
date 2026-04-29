package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"notflex_client_api/common/database"
	"notflex_client_api/enum"
	"notflex_client_api/helpers"
	"notflex_client_api/mailer"
	"notflex_client_api/models"
	"notflex_client_api/templates"
)

func CreateRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email" validate:"required,email,max=100"`
	}
	logParams := []any{"handler", "CreateRegistrationRequest"}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}
	if strings.Contains(body.Email, "+") {
		HandleResponseError(w, r, NewValidationError(map[string]string{"email": helpers.Translate(r.Context(), "InvalidEmail")}))
		return
	}

	var count int64
	if err := database.DB.WithContext(r.Context()).Model(&models.User{}).
		Where("email = ?", strings.ToLower(body.Email)).Count(&count).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("checking email uniqueness", err, logParams...))
		return
	}
	if count > 0 {
		HandleResponseError(w, r, NewValidationError(map[string]string{"email": helpers.Translate(r.Context(), "EmailExisted")}))
		return
	}

	verifyCode := helpers.RandomString(6, "0123456789")
	request := models.RegisterRequest{
		Email:            strings.ToLower(body.Email),
		ConfirmationCode: verifyCode,
		ExpireAt:         time.Now().Add(5 * time.Minute),
	}

	// Xóa request cũ để tránh rác
	database.DB.WithContext(r.Context()).Where("email = ?", request.Email).Delete(&models.RegisterRequest{})

	if err := database.DB.WithContext(r.Context()).Create(&request).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating registration request", err, logParams...))
		return
	}

	locale, _ := r.Context().Value(enum.ContextKeyLocale).(string)
	err := mailer.Send(r.Context(), mailer.Message{
		To:      request.Email,
		Subject: helpers.Translate(r.Context(), "RegisterCodeSubject"),
		Body:    templates.Register(locale, templates.RegisterParam{Code: verifyCode, Email: request.Email}),
	})
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("sending registration email", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"id": request.ID})
}

func RegistrationVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID         string `json:"id" validate:"required"`
		VerifyCode string `json:"verifyCode" validate:"required"`
	}
	logParams := []any{"handler", "RegistrationVerify"}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	var request models.RegisterRequest
	if err := database.DB.WithContext(r.Context()).Where("id = ?", body.ID).First(&request).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("RegistrationRequestNotFound", logParams...))
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

	if err := database.DB.WithContext(r.Context()).Model(&request).Update("verified", true).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating verify status", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func ConfirmRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID       string `json:"id" validate:"required"`
		FullName string `json:"fullName" validate:"required,min=2,max=50"`
		Password string `json:"password" validate:"required,min=8,max=50"`
	}
	logParams := []any{"handler", "ConfirmRegistrationRequest"}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if errs := helpers.ValidateStruct(body); errs != nil {
		HandleResponseError(w, r, NewValidationError(errs))
		return
	}

	for _, pattern := range []string{`[a-z]`, `[A-Z]`, `[0-9]`, `[^a-zA-Z0-9]`} {
		matched, _ := regexp.MatchString(pattern, body.Password)
		if !matched {
			HandleResponseError(w, r, NewValidationError(map[string]string{"password": helpers.Translate(r.Context(), "StrongPassword")}))
			return
		}
	}

	var request models.RegisterRequest
	if err := database.DB.WithContext(r.Context()).Where("id = ?", body.ID).First(&request).Error; err != nil {
		HandleResponseError(w, r, NewNotFoundError("RegistrationRequestNotFound", logParams...))
		return
	}
	if !request.Verified {
		HandleResponseError(w, r, NewBadRequestError("RequestNotVerified", logParams...))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("hashing password", err, logParams...))
		return
	}

	user := models.User{
		Email:        request.Email,
		PasswordHash: string(hash),
		FullName:     body.FullName,
		Role:         "guest",
	}
	if err := database.DB.WithContext(r.Context()).Create(&user).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating user", err, logParams...))
		return
	}

	token := models.UserToken{UserID: user.ID, UserAgent: r.UserAgent()}
	if err := database.DB.WithContext(r.Context()).Create(&token).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating user token", err, logParams...))
		return
	}

	database.DB.WithContext(r.Context()).Delete(&request)

	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"token": token.ID, "user": user})
}
