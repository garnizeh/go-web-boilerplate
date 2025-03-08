package middleware

import (
	"context"
	"time"

	"github.com/garnizeh/go-web-boilerplate/pkg/logger"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const valuesKey = "values"

func newValues() *logger.Values {
	return &logger.Values{
		TraceID: uuid.NewString(),
		Now:     time.Now().UTC(),
	}
}

func getValues(c echo.Context) *logger.Values {
	v, ok := c.Get(valuesKey).(*logger.Values)
	if !ok {
		return &logger.Values{
			TraceID:    "00000000-0000-0000-0000-000000000000",
			Now:        time.Now(),
			StatusCode: 0,
		}
	}

	return v
}

func getPath(c echo.Context) string {
	path := c.Request().URL.Path
	rawQuery := c.Request().URL.Query()
	if len(rawQuery) > 0 {
		path += "?" + rawQuery.Encode()
	}

	return path
}

func PrepareLogger(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			v := getValues(c)

			path := getPath(c)
			method := c.Request().Method
			remoteAddr := c.Request().RemoteAddr

			ctx := ContextWithValues(c)
			log.Info(ctx, "request started", "method", method, "path", path, "remote_addr", remoteAddr)

			err := next(c)

			log.Info(ctx, "request completed", "method", method, "path", path, "remote_addr", remoteAddr,
				"status_code", v.StatusCode, "since", time.Since(v.Now).String())

			return err
		}
	}
}

func PrepareLoggerValues(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			v := newValues()

			c.Set(valuesKey, v)

			err := next(c)
			if err != nil {
				ctx := ContextWithValues(c)
				log.Error(ctx, "web request", "ERROR", err)
			}

			return err
		}
	}
}

func ContextWithValues(e echo.Context) context.Context {
	ctx := e.Request().Context()

	v := getValues(e)
	ctx = context.WithValue(ctx, logger.Key, v)

	return ctx
}
