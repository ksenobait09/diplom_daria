package context

import (
	"context"
	"diplom/pkg/models"

	"github.com/pkg/errors"
)

type userKey struct{}

func MustGetUser(ctx context.Context) *models.User {
	user, err := GetUser(ctx)
	if err != nil {
		panic(err)
	}

	return user
}

func GetUser(ctx context.Context) (*models.User, error) {
	rawUser := ctx.Value(userKey{})
	if rawUser == nil {
		return nil, errors.New("no user in context")
	}

	user, ok := rawUser.(*models.User)
	if !ok {
		return nil, errors.Errorf("bad type of stored user in context, got type %T", rawUser)
	}

	return user, nil
}

func StoreUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}
