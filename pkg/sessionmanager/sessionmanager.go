package sessionmanager

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
)

// SessionManager takes care of the opened sessions of the app.
type SessionManager struct {
	sm *scs.SessionManager
	db *sql.DB
}

// New creates a new session manager. It expects a sqlite database as
// argument to manage the sessions.
func New(db *sql.DB) *SessionManager {
	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 24 * time.Hour * 7
	sessionManager.IdleTimeout = 24 * time.Hour
	sessionManager.Cookie.Name = "_s"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.Persist = false
	sessionManager.Cookie.Secure = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	return &SessionManager{
		sm: sessionManager,
		db: db,
	}
}

func (sm *SessionManager) Close() error {
	return sm.db.Close()
}

func (sm *SessionManager) SessionManager() *scs.SessionManager {
	return sm.sm
}

func (sm *SessionManager) Echo() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			var token string
			cookie, err := c.Cookie(sm.sm.Cookie.Name)
			if err == nil {
				token = cookie.Value
			}

			ctx, err = sm.sm.Load(ctx, token)
			if err != nil {
				return err
			}

			c.SetRequest(c.Request().WithContext(ctx))

			c.Response().Before(func() {
				if sm.sm.Status(ctx) != scs.Unmodified {
					responseCookie := &http.Cookie{
						Name:     sm.sm.Cookie.Name,
						Path:     sm.sm.Cookie.Path,
						Domain:   sm.sm.Cookie.Domain,
						Secure:   sm.sm.Cookie.Secure,
						HttpOnly: sm.sm.Cookie.HttpOnly,
						SameSite: sm.sm.Cookie.SameSite,
					}

					switch sm.sm.Status(ctx) {
					case scs.Modified:
						token, expiry, err := sm.sm.Commit(ctx)
						if err != nil {
							panic(err)
						}

						if sm.sm.GetBool(ctx, "__rememberMe") {
							responseCookie.Expires = time.Unix(expiry.Unix()+1, 0)        // Round up to the nearest second.
							responseCookie.MaxAge = int(time.Until(expiry).Seconds() + 1) // Round up to the nearest second.
						}

						responseCookie.Value = token

					case scs.Destroyed:
						responseCookie.Expires = time.Unix(1, 0)
						responseCookie.MaxAge = -1
					}

					c.SetCookie(responseCookie)
					addHeaderIfMissing(c.Response(), "Cache-Control", `no-cache="Set-Cookie"`)
					addHeaderIfMissing(c.Response(), "Vary", "Cookie")
				}
			})

			return next(c)
		}
	}
}

func addHeaderIfMissing(w http.ResponseWriter, key, value string) {
	for _, h := range w.Header()[key] {
		if h == value {
			return
		}
	}
	w.Header().Add(key, value)
}
