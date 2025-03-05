package auth

import (
	"github.com/garnizeh/go-web-boilerplate/internal/templates"
	viewauth "github.com/garnizeh/go-web-boilerplate/internal/templates/views/auth"
	"github.com/labstack/echo/v4"
)

func GetSignin(engine *templates.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		return engine.Render(viewauth.Signin(), c)
	}
}
