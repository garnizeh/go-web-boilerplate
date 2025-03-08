package auth

import (
	"github.com/alexedwards/scs/v2"
	"github.com/garnizeh/go-web-boilerplate/internal/templates"
	"github.com/garnizeh/go-web-boilerplate/service/user"
)

type Auth struct {
	sessionManager *scs.SessionManager
	engine         *templates.Engine
	userService    *user.Service
}

func New(
	sessionManager *scs.SessionManager,
	engine *templates.Engine,
	userService *user.Service,
) *Auth {
	return &Auth{
		sessionManager: sessionManager,
		engine:         engine,
		userService:    userService,
	}
}
