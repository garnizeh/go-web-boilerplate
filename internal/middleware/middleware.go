package middleware

import (
	"fmt"
	"strings"

	"github.com/garnizeh/go-web-boilerplate/service/user"

    "github.com/alexedwards/scs/v2"
    "github.com/labstack/echo/v4"
)

const contextKeyEmail = "email"

type SessionData struct {
	AppName   string
	Email     string
	Name      string
	ErrMsg    string
	FlashMsg  string
	CSRFToken string
	Fields    any
}

func (sd SessionData) SignedIn() bool {
	return sd.Email != ""
}

func PrepareSessionData(sessionManager *scs.SessionManager, users *user.Service, appName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			// Skip static endpoints.
			if strings.HasPrefix(req.URL.Path, "/static") {
				return next(c)
			}

			sessionData := SessionData{AppName: appName}

			ctx := req.Context()
			email := sessionManager.GetString(ctx, contextKeyEmail)
			if email != "" {
				user, err := users.GetUser(ctx, email)
				if err != nil {
					// TODO: need to clear the session/cookie and redirect to signin.
					panic(fmt.Sprintf("failed to get user with email %q: %v", email, err))
				}

				sessionData.Email = user.Email
				sessionData.Name = user.Name
			}
			tk, ok := c.Get("csc").(string)
			if ok {
				sessionData.CSRFToken = tk
			}

			c.Set("sessionData", sessionData)
			return next(c)
		}
	}
}
