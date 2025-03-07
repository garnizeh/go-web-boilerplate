package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/garnizeh/go-web-boilerplate/internal/templates"
	"github.com/garnizeh/go-web-boilerplate/internal/templates/views/auth"
	"github.com/garnizeh/go-web-boilerplate/pkg/validator"
	"github.com/labstack/echo/v4"
)

// TODO: get these values from config, maybe even some rules
const (
	passwordMinSize = 4
	passwordMaxSize = 32
)

var (
	errInvalidEmail    = errors.New("invalid email")
	errInvalidPassword = errors.New("invalid password")
)

func GetSignin(engine *templates.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Get("sec").(string)
		return engine.Render(c, auth.Signin(token))
	}
}

type signinRequest struct {
	Email    string `form:"email"`
	Password string `form:"password"`
	Remember string `form:"remember"`
}

func (s *signinRequest) validate() error {
	s.Email = validator.SanitizeString(strings.ToLower(s.Email))
	if !validator.IsValidEmail(s.Email) {
		return errInvalidEmail
	}

	s.Password = validator.SanitizeString(s.Password)
	if len(s.Password) < passwordMinSize || len(s.Password) > passwordMaxSize {
		return errInvalidPassword
	}

	s.Remember = validator.SanitizeString(s.Remember)

	return nil
}

func (s *signinRequest) isRemember() bool {
	return s.Remember == "true"
}

func PostSignin(engine *templates.Engine) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(signinRequest)
		if err := c.Bind(req); err != nil {
			return err
		}

		if err := req.validate(); err != nil {
			token := c.Get("sec").(string)
			c.Response().WriteHeader(http.StatusUnauthorized)
			return auth.SigninError(token).Render(c.Request().Context(), c.Response().Writer)
		}

		return c.JSON(200, req)
	}
}
