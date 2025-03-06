package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/labstack/echo/v4"
)

var nonceKey = "nonces"

type Nonces struct {
	Htmx            string
	ResponseTargets string
	Tw              string
	HtmxCSSHash     string
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func CSP(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create a new Nonces struct for every request when here.
		// Move to outside the handler to use the same nonces in all responses
		nonceSet := Nonces{
			Htmx:            generateRandomString(16),
			ResponseTargets: generateRandomString(16),
			Tw:              generateRandomString(16),
			HtmxCSSHash:     "sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg=",
		}

		// Set nonces in context
		c.Set(nonceKey, nonceSet)

		//ContentSecurityPolicy: "default-src 'self'; script-src 'self'; object-src 'self'; style-src 'self'; img-src 'self'; media-src 'self'; frame-ancestors 'self'; frame-src 'self'; connect-src 'self'",
		// Insert the nonces into the content security policy header
		csp := fmt.Sprintf("default-src 'self'; script-src 'nonce-%s' 'nonce-%s' ; style-src 'nonce-%s' '%s';",
			nonceSet.Htmx,
			nonceSet.ResponseTargets,
			nonceSet.Tw,
			nonceSet.HtmxCSSHash,
		)
		c.Response().Header().Set(echo.HeaderContentSecurityPolicy, csp)

		return next(c)
	}
}
