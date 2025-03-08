package templates

import (
	"github.com/a-h/templ"
	"github.com/garnizeh/go-web-boilerplate/internal/middleware"
	viewlayout "github.com/garnizeh/go-web-boilerplate/internal/templates/layout"
	"github.com/labstack/echo/v4"
)

type Engine struct {
	title   string
	isDebug bool
}

func New(title string, isDebug bool) *Engine {
	return &Engine{
		title:   title,
		isDebug: isDebug,
	}
}

func (e *Engine) Render(c echo.Context, contents templ.Component, full bool) error {
	ctx := c.Request().Context()
	w := c.Response().Writer
	if full {
		nonce := c.Get(middleware.NonceKey).(string)
		return viewlayout.Layout(contents, e.title, nonce, e.isDebug).Render(ctx, w)
	}

	return contents.Render(ctx, w)
}
