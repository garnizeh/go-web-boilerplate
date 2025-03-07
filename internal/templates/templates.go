package templates

import (
	"github.com/a-h/templ"
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

func (e *Engine) Render(c echo.Context, contents templ.Component) error {
	ctx := c.Request().Context()
	w := c.Response().Writer

	return viewlayout.Layout(contents, e.title, e.isDebug).Render(ctx, w)
}
