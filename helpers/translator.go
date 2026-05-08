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
	"RegisterCodeSubject":         "Your Notflex verification code",
	"LoginCodeSubject":            "Your Notflex sign-in code",
	"LoginRequestNotFound":        "Login request not found",
	"PasswordMismatch":            "Passwords do not match",
	"WrongOldPassword":            "Current password is incorrect",
}

func Translate(_ context.Context, tag string) string {
	if msg, ok := messages[tag]; ok {
		return msg
	}
	return tag
}
