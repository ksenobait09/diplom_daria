package middleware

import (
	"diplom/pkg/auth"
	"diplom/pkg/context"
	"diplom/pkg/cookie"
	"diplom/pkg/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
)

func checkSession(authRepo *auth.Repo, SessionID models.SessionID) (*models.User, error) {
	if SessionID == "" {
		return nil, nil
	}

	user, err := authRepo.CheckSession(SessionID)
	if errors.Is(err, auth.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to check user's session")
	}

	return user, nil
}

func Auth(authRepo *auth.Repo, logger *log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			user, err := checkSession(authRepo, cookie.GetSessionCookie(c))
			if err != nil {
				return err
			}

			ctx = context.StoreUser(ctx, user)

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
