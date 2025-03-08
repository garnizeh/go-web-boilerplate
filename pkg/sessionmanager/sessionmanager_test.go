package sessionmanager_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/go-web-boilerplate/pkg/sessionmanager"
	"github.com/garnizeh/go-web-boilerplate/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	db, err := storage.NewDBSqlite(":memory:")
	if err != nil {
		t.Fatalf("failed to create sqlite database in memory: %v", err)
	}
	defer db.Close()

	if err := storage.MigrateSessions(db); err != nil {
		t.Fatalf("failed to migrate sessions table: %v", err)
	}

	sm := sessionmanager.New(db)

	e := echo.New()

	// Call /put to set the message in the session manager
	req := httptest.NewRequest(http.MethodGet, "/put", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session := sm.Echo()

	h := session(func(c echo.Context) error {
		sm.SessionManager().Put(c.Request().Context(), "message", "Hello from a session!")
		return c.String(http.StatusOK, "")
	})

	if err := h(c); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}

	assert.Equal(t, rec.Result().StatusCode, 200)
	assert.Equal(t, len(rec.Result().Cookies()), 1)

	sessionCookie := rec.Result().Cookies()[0]

	assert.Equal(t, sessionCookie.Name, "_s")

	// Make a request to /get to see if the message is still there
	req = httptest.NewRequest(http.MethodGet, "/get", nil)
	req.Header.Set("Cookie", sessionCookie.Name+"="+sessionCookie.Value)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	h = session(func(c echo.Context) error {
		msg := sm.SessionManager().GetString(c.Request().Context(), "message")
		return c.String(http.StatusOK, msg)
	})

	if err := h(c); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}

	assert.Equal(t, rec.Result().StatusCode, 200)
	assert.Equal(t, rec.Body.String(), "Hello from a session!")
}
