package helpers

import (
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

func ValidateStruct(data interface{}) map[string]string {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
	})

	err := validate.Struct(data)
	if err != nil {
		errs := map[string]string{}
		for _, e := range err.(validator.ValidationErrors) {
			field := LowerFirstChar(e.Field())
			errs[field] = msgForTag(e)
		}
		return errs
	}
	return nil
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Must not exceed %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("Failed validation on '%s'", fe.Tag())
	}
}
