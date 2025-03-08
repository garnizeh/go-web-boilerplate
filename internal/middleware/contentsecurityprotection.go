package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

const NonceKey = "nonce"

var (
	styleSrcElems = []string{
		"'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='",
		"'sha256-G2K9ENiXaTIc4pmzLEOJB962ySgP2gMolWCZ6HJpU4I='",
		"'sha256-2Zmme+3cWvmG8lapM3WvEkAyYA3671LVoN107gkAU4g='",
	}
)

// https://templ.guide/security/content-security-policy
func PrepareCSP() echo.MiddlewareFunc {
	styleSrcElem := strings.Join(styleSrcElems, " ")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		nonce := generateRandomString(16)
		csp := fmt.Sprintf("default-src 'self'; base-uri 'self'; form-action 'self'; script-src 'self'; script-src-elem 'nonce-%s'; script-src-attr 'self'; object-src 'self'; style-src 'self'; style-src-elem 'self' %s; style-src-attr 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https://api.iconify.design  https://api.simplesvg.com https://api.unisvg.com; media-src 'self'; frame-ancestors 'self'; frame-src 'self'; child-src 'self';", nonce, styleSrcElem)

		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderContentSecurityPolicy, csp)

			c.Set(NonceKey, nonce)

			ctx := c.Request().Context()
			ctx = templ.WithNonce(ctx, nonce)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return strings.Repeat("0", length)
	}

	return hex.EncodeToString(bytes)
}
