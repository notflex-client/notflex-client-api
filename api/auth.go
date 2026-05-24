package api

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"notflex_client_api/common/database"
	"notflex_client_api/enum"
	"notflex_client_api/helpers"
	"notflex_client_api/mailer"
	"notflex_client_api/models"
	"notflex_client_api/templates"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
		Remember bool   `json:"remember"`
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

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		HandleResponseError(w, r, NewBadRequestError("AccountLocked", logParams...))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		newCount := user.FailedLoginCount + 1
		updates := map[string]any{"failed_login_count": newCount}
		if newCount >= 5 {
			lockedUntil := time.Now().Add(15 * time.Minute)
			updates["locked_until"] = lockedUntil
			updates["failed_login_count"] = 0
		}
		database.DB.WithContext(r.Context()).Model(&user).Updates(updates)
		HandleResponseError(w, r, NewBadRequestError("InvalidCredentials", logParams...))
		return
	}

	// Reset lockout state on successful login
	if user.FailedLoginCount > 0 || user.LockedUntil != nil {
		database.DB.WithContext(r.Context()).Model(&user).Updates(map[string]any{
			"failed_login_count": 0,
			"locked_until":       nil,
		})
	}

	expireAt := time.Now().Add(24 * time.Hour)
	if body.Remember {
		expireAt = time.Now().Add(30 * 24 * time.Hour)
	}

	token := models.UserToken{UserAgent: r.UserAgent(), UserID: user.ID, ExpireAt: expireAt}
	if err := database.DB.WithContext(r.Context()).Create(&token).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating user token", err, logParams...))
		return
	}

	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"token": token.ID, "user": user})
}

func CreateLoginCodeRequest(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "CreateLoginCodeRequest"}

	var body struct {
		Email string `json:"email" validate:"required,email,max=100"`
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

	email := strings.ToLower(body.Email)
	logParams = append(logParams, "email", email)

	var user models.User
	err = database.DB.WithContext(r.Context()).
		Where("email = ? AND is_active = TRUE", email).
		First(&user).Error
	if err != nil {
		HandleResponseError(w, r, NewValidationError(map[string]string{"email": helpers.Translate(r.Context(), "AccountNotFound")}))
		return
	}

	verifyCode := helpers.RandomString(6, "0123456789")
	request := models.LoginRequest{
		Email:            email,
		ConfirmationCode: verifyCode,
		ExpireAt:         time.Now().Add(5 * time.Minute),
	}

	database.DB.WithContext(r.Context()).Where("email = ?", email).Delete(&models.LoginRequest{})

	err = database.DB.WithContext(r.Context()).Create(&request).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating login request", err, logParams...))
		return
	}

	locale, _ := r.Context().Value(enum.ContextKeyLocale).(string)
	err = mailer.Send(r.Context(), mailer.Message{
		To:      email,
		Subject: helpers.Translate(r.Context(), "LoginCodeSubject"),
		Body:    templates.LoginCode(locale, templates.LoginCodeParam{Code: verifyCode, Email: email}),
	})
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("sending login code email", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"id": request.ID})
}

func ConfirmLoginCode(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ConfirmLoginCode"}

	var body struct {
		ID         string `json:"id" validate:"required"`
		VerifyCode string `json:"verifyCode" validate:"required"`
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

	var request models.LoginRequest
	err = database.DB.WithContext(r.Context()).Where("id = ?", body.ID).First(&request).Error
	if err != nil {
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
	err = database.DB.WithContext(r.Context()).
		Where("email = ? AND is_active = TRUE", request.Email).
		First(&user).Error
	if err != nil {
		HandleResponseError(w, r, NewValidationError(map[string]string{"email": helpers.Translate(r.Context(), "AccountNotFound")}))
		return
	}

	logParams = append(logParams, "userID", user.ID)

	token := models.UserToken{UserAgent: r.UserAgent(), UserID: user.ID, ExpireAt: time.Now().Add(24 * time.Hour)}
	err = database.DB.WithContext(r.Context()).Create(&token).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating user token", err, logParams...))
		return
	}

	database.DB.WithContext(r.Context()).Delete(&request)

	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"token": token.ID, "user": user})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "Logout"}

	tokenID, _ := r.Context().Value(enum.ContextKeyToken).(string)
	if err := database.DB.WithContext(r.Context()).Where("id = ?", tokenID).Delete(&models.UserToken{}).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting user token", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ForgotPassword"}

	var body struct {
		Email string `json:"email" validate:"required,email"`
	}
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
		HandleResponseError(w, r, NewValidationError(map[string]string{"email": helpers.Translate(r.Context(), "AccountNotFound")}))
		return
	}

	logParams = append(logParams, "userID", user.ID)

	// Rate limit: chỉ gửi email mới nếu đã qua 5 phút kể từ lần trước
	if user.ResetPasswordTime > 0 && user.ResetPasswordTime+5*60 > time.Now().Unix() {
		json.NewEncoder(w).Encode(map[string]any{
			"success":           true,
			"resetPasswordTime": user.ResetPasswordTime,
		})
		return
	}

	token := uuid.NewString()
	now := time.Now().Unix()
	if err := database.DB.WithContext(r.Context()).Model(&user).Updates(map[string]any{
		"reset_password_token": token,
		"reset_password_time":  now,
	}).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating reset password token", err, logParams...))
		return
	}

	resetURL := strings.TrimRight(os.Getenv("CLIENT_APP_URL"), "/") + "/reset-password/" + token
	locale, _ := r.Context().Value(enum.ContextKeyLocale).(string)

	if err := mailer.Send(r.Context(), mailer.Message{
		To:      user.Email,
		Subject: helpers.Translate(r.Context(), "ResetPasswordEmailSubject"),
		Body:    templates.ResetPassword(locale, resetURL),
	}); err != nil {
		HandleResponseError(w, r, NewInternalServerError("sending reset password email", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success":           true,
		"resetPasswordTime": now,
	})
}

func GetForgotPasswordToken(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "GetForgotPasswordToken"}

	token := chi.URLParam(r, "token")
	if token == "" {
		HandleResponseError(w, r, NewBadRequestError("InvalidResetToken", logParams...))
		return
	}

	var user models.User
	err := database.DB.WithContext(r.Context()).
		Where("reset_password_token = ? AND is_active = TRUE", token).
		First(&user).Error

	isValid := err == nil && user.ResetPasswordTime > 0 && user.ResetPasswordTime+15*60 > time.Now().Unix()
	json.NewEncoder(w).Encode(map[string]any{"isValid": isValid})
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ResetPassword"}

	var body struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"newPassword" validate:"required,min=8,max=50"`
	}
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
		Where("reset_password_token = ? AND is_active = TRUE", body.Token).
		First(&user).Error
	if err != nil {
		HandleResponseError(w, r, NewValidationError(map[string]string{"token": helpers.Translate(r.Context(), "InvalidResetToken")}))
		return
	}
	if user.ResetPasswordTime == 0 || user.ResetPasswordTime+15*60 < time.Now().Unix() {
		HandleResponseError(w, r, NewValidationError(map[string]string{"token": helpers.Translate(r.Context(), "ExpiredResetToken")}))
		return
	}

	logParams = append(logParams, "userID", user.ID)

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

	if err := database.DB.WithContext(r.Context()).Model(&user).Updates(map[string]any{
		"password_hash":        string(hash),
		"reset_password_token": "",
		"reset_password_time":  0,
	}).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating password", err, logParams...))
		return
	}

	// Đăng xuất tất cả thiết bị sau khi đổi mật khẩu
	if err := database.DB.WithContext(r.Context()).Where("user_id = ?", user.ID).Delete(&models.UserToken{}).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting user tokens", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"success": true})
}
