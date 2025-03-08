package auth

import (
	"github.com/alexedwards/scs/v2"
	"github.com/garnizeh/go-web-boilerplate/internal/middleware"
	"github.com/garnizeh/go-web-boilerplate/internal/templates"
	"github.com/garnizeh/go-web-boilerplate/pkg/logger"
	"github.com/garnizeh/go-web-boilerplate/service"
	"github.com/labstack/echo/v4"
)

const CSRFKey = "csrf"

type auth struct {
	engine         *templates.Engine
	service        *service.Service
	sessionManager *scs.SessionManager
}

func Mount(
	log *logger.Logger,
	g *echo.Group,
	sessionManager *scs.SessionManager,
	engine *templates.Engine,
	service *service.Service,
) {
	a := &auth{
		engine:         engine,
		service:        service,
		sessionManager: sessionManager,
	}

	g.GET("/signin", a.getSignin, middleware.IsSignedOutMiddleware)
	g.POST("/signin", a.postSignin, middleware.IsSignedOutMiddleware)
}
