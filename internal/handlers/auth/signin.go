package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/garnizeh/go-web-boilerplate/internal/middleware"
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

func (a *Auth) GetSignin(c echo.Context) error {
	token := c.Get("sec").(string)
	return a.engine.Render(c, auth.Signin(token), true)
}

func (a *Auth) PostSignin(c echo.Context) error {
	req := new(signinRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	badCredentialsFunc := func() error {
		token := c.Get("sec").(string)
		c.Response().WriteHeader(http.StatusUnauthorized)
		return a.engine.Render(c, auth.SigninError(token, req.Email, req.Password, req.Remember), false)
	}

	if err := req.validate(); err != nil {
		return badCredentialsFunc()
	}

	ctx := c.Request().Context()
	user, err := a.userService.Signin(ctx, req.Email, req.Password)
	if err != nil {
		// TODO: check for email not verified case first
		return badCredentialsFunc()
	}

	if !req.isRemember() {
		if err := a.sessionManager.RenewToken(ctx); err != nil {
			return c.Redirect(http.StatusSeeOther, "/error-500")
		}
	}

	a.sessionManager.Put(ctx, middleware.ContextKeyEmail, user.Email)
	a.sessionManager.RememberMe(ctx, req.isRemember())

	w := c.Response()
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
	return nil
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
