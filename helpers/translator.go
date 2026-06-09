package helpers

import "context"

var messages = map[string]string{
	"Unauthorized":                "Unauthorized",
	"Forbidden":                   "Access denied",
	"InvalidBodyStructure":        "Invalid request body",
	"InvalidID":                   "Invalid ID",
	"EmailExisted":                "Email already in use",
	"InvalidEmail":                "Invalid email address",
	"WrongConfirmationCode":       "Wrong confirmation code",
	"ExpiredConfirmationCode":     "Confirmation code has expired",
	"RegistrationRequestNotFound": "Registration request not found",
	"RequestNotVerified":          "Email has not been verified",
	"StrongPassword":              "Password must contain uppercase, lowercase, number, and special character",
	"InvalidCredentials":          "Invalid email or password",
	"UserNotFound":                "User not found",
	"MovieNotFound":               "Movie not found",
	"GenreNotFound":               "Genre not found",
	"InvalidGenreID":              "Invalid genre ID",
	"RegisterCodeSubject":          "Your Notflex verification code",
	"AccountNotFound":              "No account found with this email address",
	"ResetPasswordEmailSubject":    "Reset your Notflex password",
	"InvalidResetToken":            "This reset link is invalid or has already been used",
	"ExpiredResetToken":            "This reset link has expired. Please request a new one",
	"AccountLocked":                "Too many failed login attempts. Your account is locked for 15 minutes.",
	"WrongCurrentPassword":         "Current password is incorrect",
	"SamePassword":                 "New password must be different from the current password",
	"LoginCodeSubject":             "Your Notflex sign-in code",
	"LoginRequestNotFound":         "Login request not found",
	"InvalidPageSize":              "Invalid page size",
	"TagNotFound":                  "Tag not found",
	"TagSlugExisted":               "Tag slug already exists",
	"PlanNotFound":                 "Subscription plan not found",
	"BannerNotFound":               "Banner not found",
	"GenreNameExisted":             "Genre name already exists",
	"CannotDeleteSelf":             "You cannot delete your own account",
	"SubscriptionPlanNotFound":     "Subscription plan not found",
	"InvalidWebhookSignature":      "Invalid webhook signature",
	"UnsupportedImageType":         "Unsupported image type",
	"InvalidMultipartForm":         "Invalid multipart form",
	"MissingFile":                  "File is required",
	"ProfileNotFound":              "Profile not found",
	"MaxProfilesReached":           "You have reached the maximum number of profiles",
	"CannotDeleteLastProfile":      "You cannot delete your last profile",
	"CannotTransferLastProfile":    "You cannot transfer your last profile",
	"TargetUserNotFound":           "No account found with this email address",
	"CannotTransferToSelf":         "You cannot transfer a profile to your own account",
	"TargetMaxProfilesReached":     "The target account has reached its profile limit",
}

func Translate(_ context.Context, tag string) string {
	if msg, ok := messages[tag]; ok {
		return msg
	}
	return tag
}
