package cookie

import (
	"diplom/pkg/models"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const SessionIDHeader = "X-Session-Id"

func GetSessionCookie(c echo.Context) models.SessionID {
	cookie, err := c.Cookie(SessionIDHeader)
	if err != nil {
		return ""
	}

	return models.SessionID(cookie.Value)
}

func SetSessionCookie(c echo.Context, SessionID models.SessionID) {
	cookie := &http.Cookie{
		Name:    SessionIDHeader,
		Value:   string(SessionID),
		Path:    "/",
		Expires: time.Now().Add(7 * 24 * time.Hour),
	}

	c.SetCookie(cookie)
}
