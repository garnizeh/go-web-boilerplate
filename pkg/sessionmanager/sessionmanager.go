package sessionmanager

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
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
	sessionManager.Cookie.Persist = true
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
