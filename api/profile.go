package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"notflex_client_api/common/database"
	"notflex_client_api/enum"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

// freeMaxProfiles is the profile cap for accounts without an active subscription.
const freeMaxProfiles = 1

// maxProfilesForUser returns how many watch profiles the account may hold, based
// on the active subscription plan (free accounts get freeMaxProfiles).
func maxProfilesForUser(r *http.Request, userID string) int {
	sub, ok, err := findActiveSubscription(r, userID)
	if err != nil || !ok || sub.Plan.MaxProfiles < freeMaxProfiles {
		return freeMaxProfiles
	}
	return sub.Plan.MaxProfiles
}

// ensureProfiles guarantees the account has at least one watch profile and
// loads them onto the user. A brand-new account (or a legacy one created
// before profiles existed) is bootstrapped with a single default profile
// mirroring the account name/avatar.
func ensureProfiles(ctx context.Context, user *models.User) error {
	var count int64
	if err := database.DB.WithContext(ctx).Model(&models.Profile{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		name := strings.TrimSpace(user.FullName)
		if name == "" {
			name = "Profile 1"
		}
		defaultProfile := models.Profile{UserID: user.ID, Name: name, AvatarURL: user.AvatarURL}
		if err := database.DB.WithContext(ctx).Create(&defaultProfile).Error; err != nil {
			return err
		}
		// Attach any legacy rows (created before profiles existed) to this default
		// profile so existing history/favorites/ratings are not orphaned.
		for _, m := range []any{&models.WatchHistory{}, &models.Favorite{}, &models.MovieRating{}} {
			database.DB.WithContext(ctx).Model(m).
				Where("user_id = ? AND profile_id IS NULL", user.ID).
				Update("profile_id", defaultProfile.ID)
		}
	}
	return database.DB.WithContext(ctx).
		Where("user_id = ?", user.ID).
		Order("created_at ASC").
		Find(&user.Profiles).Error
}

// resolveProfileID returns the active watch profile for the request. It reads the
// X-Profile-Id header and validates that it belongs to the account; otherwise it
// falls back to the account's default (earliest) profile, bootstrapping one if
// none exist yet.
func resolveProfileID(r *http.Request, user *models.User) (string, error) {
	headerID := strings.TrimSpace(r.Header.Get("X-Profile-Id"))
	if headerID != "" {
		var count int64
		if err := database.DB.WithContext(r.Context()).Model(&models.Profile{}).
			Where("id = ? AND user_id = ?", headerID, user.ID).Count(&count).Error; err != nil {
			return "", err
		}
		if count > 0 {
			return headerID, nil
		}
	}
	if err := ensureProfiles(r.Context(), user); err != nil {
		return "", err
	}
	if len(user.Profiles) == 0 {
		return "", nil
	}
	return user.Profiles[0].ID, nil
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "GetProfile"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	if err := ensureProfiles(r.Context(), &user); err != nil {
		HandleResponseError(w, r, NewInternalServerError("loading profiles", err, logParams...))
		return
	}

	user.PasswordHash = ""
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func ListProfiles(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "ListProfiles"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	if err := ensureProfiles(r.Context(), &user); err != nil {
		HandleResponseError(w, r, NewInternalServerError("listing profiles", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"items":        user.Profiles,
		"max_profiles": maxProfilesForUser(r, user.ID),
	})
}

func CreateProfile(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "CreateProfile"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	var body struct {
		Name      string  `json:"name" validate:"required,min=1,max=100"`
		AvatarURL *string `json:"avatar_url"`
		IsKids    bool    `json:"is_kids"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	var count int64
	if err = database.DB.WithContext(r.Context()).Model(&models.Profile{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting profiles", err, logParams...))
		return
	}
	if count >= int64(maxProfilesForUser(r, user.ID)) {
		HandleResponseError(w, r, NewValidationError(map[string]string{"name": helpers.Translate(r.Context(), "MaxProfilesReached")}))
		return
	}

	profile := models.Profile{
		UserID:    user.ID,
		Name:      strings.TrimSpace(body.Name),
		AvatarURL: body.AvatarURL,
		IsKids:    body.IsKids,
	}
	if err = database.DB.WithContext(r.Context()).Create(&profile).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating profile", err, logParams...))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "UpdateUserProfile"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	profileID := chi.URLParam(r, "id")
	logParams = append(logParams, "profileID", profileID)

	var profile models.Profile
	err = database.DB.WithContext(r.Context()).Where("id = ? AND user_id = ?", profileID, user.ID).First(&profile).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("ProfileNotFound", logParams...))
		return
	}

	var body struct {
		Name      string  `json:"name" validate:"required,min=1,max=100"`
		AvatarURL *string `json:"avatar_url"`
		IsKids    bool    `json:"is_kids"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	updates := map[string]any{
		"name":       strings.TrimSpace(body.Name),
		"avatar_url": body.AvatarURL,
		"is_kids":    body.IsKids,
	}
	if err = database.DB.WithContext(r.Context()).Model(&profile).Updates(updates).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("updating profile", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func DeleteUserProfile(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "DeleteUserProfile"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	profileID := chi.URLParam(r, "id")
	logParams = append(logParams, "profileID", profileID)

	var profile models.Profile
	err = database.DB.WithContext(r.Context()).Where("id = ? AND user_id = ?", profileID, user.ID).First(&profile).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("ProfileNotFound", logParams...))
		return
	}

	var count int64
	if err = database.DB.WithContext(r.Context()).Model(&models.Profile{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting profiles", err, logParams...))
		return
	}
	if count <= 1 {
		HandleResponseError(w, r, NewValidationError(map[string]string{"id": helpers.Translate(r.Context(), "CannotDeleteLastProfile")}))
		return
	}

	if err = database.DB.WithContext(r.Context()).Delete(&profile).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("deleting profile", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(true)
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

// TransferProfile moves one of the account's watch profiles (with its history,
// favorites and ratings) to another account identified by email.
func TransferProfile(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "TransferProfile"}

	user, err := helpers.GetUserFromContext(r.Context())
	if err != nil {
		HandleResponseError(w, r, NewUnauthorizedError())
		return
	}
	logParams = append(logParams, "userID", user.ID)

	var body struct {
		ProfileID   string `json:"profileId" validate:"required"`
		TargetEmail string `json:"targetEmail" validate:"required,email"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	if formErrors := helpers.ValidateStruct(body); formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	// Profile must belong to the requesting account.
	var profile models.Profile
	err = database.DB.WithContext(r.Context()).Where("id = ? AND user_id = ?", body.ProfileID, user.ID).First(&profile).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("ProfileNotFound", logParams...))
		return
	}

	// An account must keep at least one profile.
	var ownCount int64
	if err = database.DB.WithContext(r.Context()).Model(&models.Profile{}).Where("user_id = ?", user.ID).Count(&ownCount).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting profiles", err, logParams...))
		return
	}
	if ownCount <= 1 {
		HandleResponseError(w, r, NewValidationError(map[string]string{"profileId": helpers.Translate(r.Context(), "CannotTransferLastProfile")}))
		return
	}

	// Target account must exist and not be the requester.
	var target models.User
	err = database.DB.WithContext(r.Context()).Where("email = ? AND is_active = TRUE", strings.ToLower(strings.TrimSpace(body.TargetEmail))).First(&target).Error
	if err != nil {
		HandleResponseError(w, r, NewValidationError(map[string]string{"targetEmail": helpers.Translate(r.Context(), "TargetUserNotFound")}))
		return
	}
	if target.ID == user.ID {
		HandleResponseError(w, r, NewValidationError(map[string]string{"targetEmail": helpers.Translate(r.Context(), "CannotTransferToSelf")}))
		return
	}

	// Target account must have room for another profile under its plan.
	var targetCount int64
	if err = database.DB.WithContext(r.Context()).Model(&models.Profile{}).Where("user_id = ?", target.ID).Count(&targetCount).Error; err != nil {
		HandleResponseError(w, r, NewInternalServerError("counting target profiles", err, logParams...))
		return
	}
	if targetCount >= int64(maxProfilesForUser(r, target.ID)) {
		HandleResponseError(w, r, NewValidationError(map[string]string{"targetEmail": helpers.Translate(r.Context(), "TargetMaxProfilesReached")}))
		return
	}

	// Re-home the profile and all of its scoped data, then log the transfer.
	err = database.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Profile{}).Where("id = ?", profile.ID).Update("user_id", target.ID).Error; err != nil {
			return err
		}
		for _, m := range []any{&models.WatchHistory{}, &models.Favorite{}, &models.MovieRating{}} {
			if err := tx.Model(m).Where("profile_id = ?", profile.ID).Update("user_id", target.ID).Error; err != nil {
				return err
			}
		}
		return tx.Create(&models.ProfileTransfer{
			FromUserID:  user.ID,
			ToUserID:    target.ID,
			ProfileName: profile.Name,
			Status:      "completed",
		}).Error
	})
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("transferring profile", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"success": true})
}
