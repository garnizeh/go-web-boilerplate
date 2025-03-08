package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func IsSignedInMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := c.Get(sessiondataKey).(sessionData)
		if !sess.SignedIn() {
			return c.Redirect(http.StatusSeeOther, "/auth/signin")
		}

		return next(c)
	}
}

func IsSignedOutMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := c.Get(sessiondataKey).(sessionData)
		if sess.SignedIn() {
			return c.Redirect(http.StatusSeeOther, "/")
		}

		return next(c)
	}
}
