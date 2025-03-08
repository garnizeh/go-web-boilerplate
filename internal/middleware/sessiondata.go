package middleware

import (
	"fmt"
	"strings"

	"github.com/garnizeh/go-web-boilerplate/service/user"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyEmail = "email"
	sessiondataKey  = "sessionData"
)

type sessionData struct {
	AppName   string
	UserEmail string
	UserName  string
	UserRoles []string
	CSRFToken string
}

func (s sessionData) SignedIn() bool {
	return s.UserEmail != ""
}

func PrepareSessionData(sessionManager *scs.SessionManager, users *user.Service, appName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			// Skip static endpoints.
			if strings.HasPrefix(req.URL.Path, "/static") {
				return next(c)
			}

			ctx := req.Context()
			email := sessionManager.GetString(ctx, ContextKeyEmail)
			sessionData := sessionData{AppName: appName}
			if email != "" {
				user, err := users.GetUser(ctx, email)
				if err != nil {
					// TODO: need to clear the session/cookie and redirect to signin.
					panic(fmt.Sprintf("failed to get user with email %q: %v", email, err))
				}

				sessionData.UserEmail = user.Email
				sessionData.UserName = user.Name
				sessionData.UserRoles = user.Roles
			}

			tk, ok := c.Get("csc").(string)
			if ok {
				sessionData.CSRFToken = tk
			}

			c.Set(sessiondataKey, sessionData)
			return next(c)
		}
	}
}
