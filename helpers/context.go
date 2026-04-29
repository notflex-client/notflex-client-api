package helpers

import (
	"context"
	"errors"

	"notflex_client_api/enum"
	"notflex_client_api/models"
)

func GetUserFromContext(ctx context.Context) (models.User, error) {
	userAny := ctx.Value(enum.ContextKeyUser)
	if user, ok := userAny.(models.User); ok {
		return user, nil
	}
	return models.User{}, errors.New("user not in context")
}
