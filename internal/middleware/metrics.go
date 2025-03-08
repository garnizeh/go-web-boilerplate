package middleware

import (
	"expvar"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/garnizeh/go-web-boilerplate/pkg/logger"
	"github.com/labstack/echo/v4"
)

const metricsKey = "metrics"

// metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently thanks to expvar. No extra abstraction is required.
type metrics struct {
	goroutines *expvar.Int
	requests   *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

// PrepareMetrics constructs the metrics the application will track.
func PrepareMetricsWithRecover(log *logger.Logger) echo.MiddlewareFunc {
	m := metrics{
		goroutines: expvar.NewInt("goroutines"),
		requests:   expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					ctx := c.Request().Context()
					trace := debug.Stack()
					log.Error(ctx, "!!!!!!!! PANIC !!!!!!!!", "error", rec, "trace", string(trace))

					m.panics.Add(1)

					err = c.Redirect(http.StatusSeeOther, "/error-500")
				}
			}()

			c.Set(metricsKey, &m)

			err = next(c)

			m.requests.Add(1)
			if m.requests.Value()%1000 == 0 {
				m.goroutines.Set(int64(runtime.NumGoroutine()))
			}

			if err != nil {
				m.errors.Add(1)
			}

			return
		}
	}
}
