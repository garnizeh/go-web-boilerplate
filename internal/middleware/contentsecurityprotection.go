package middleware

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

func PrepareCSP(isDebug bool) echo.MiddlewareFunc {
	// TODO: change to use nonces

	//var csp string
	csp := "default-src 'self'; base-uri 'self'; form-action 'self'; script-src 'self'; script-src-elem 'self' %s; script-src-attr 'self'; object-src 'self'; style-src 'self'; style-src-elem 'self' %s; style-src-attr 'self'; img-src 'self'; font-src 'self'; connect-src 'self' %s; media-src 'self'; frame-ancestors 'self'; frame-src 'self'; child-src 'self';"
	if isDebug {
		commonPasswordToggleJS := "'sha256-A5Awe5vXZ6juDgrEBJU49pdHLmRRxShFbH+gF4R5JkM='"   // #nosec G101
		commonFormValidationJS := "'sha256-F4Y3UcBwjxA7lOwycXHk3C3VDEScdK/IO5gDa5Iz82Q='"   // #nosec G101
		authSignin := "'sha256-rmLJ+0mwI9lyyqu3NXcfT5CF557j2UbKMa4gS6HJZTw='"               // #nosec G101
		authSigninFormValidation := "'sha256-lb3uDNz5d+s9bEe0k7UhTdA7l3reFXNxdOwcvxrKMJw='" // #nosec G101
		scriptSrcElem := strings.Join(
			[]string{
				commonPasswordToggleJS,
				commonFormValidationJS,
				authSignin,
				authSigninFormValidation,
			},
			" ",
		)

		iconifyCSS := "'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=' 'sha256-G2K9ENiXaTIc4pmzLEOJB962ySgP2gMolWCZ6HJpU4I=' 'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=' 'sha256-G2K9ENiXaTIc4pmzLEOJB962ySgP2gMolWCZ6HJpU4I='" // #nosec G101
		htmxCSS := "'sha256-2Zmme+3cWvmG8lapM3WvEkAyYA3671LVoN107gkAU4g='"                                                                                                                                                                      // #nosec G101
		styleSrcElem := strings.Join(
			[]string{
				iconifyCSS,
				htmxCSS,
			},
			" ",
		)

		iconifySrc := "https://api.iconify.design https://api.simplesvg.com https://api.unisvg.com https://api.iconify.design https://api.simplesvg.com https://api.unisvg.com" // #nosec G101
		connectSrc := strings.Join(
			[]string{
				iconifySrc,
			},
			" ",
		)

		csp = fmt.Sprintf(csp, scriptSrcElem, styleSrcElem, connectSrc)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Insert the hashes into the content security policy header
			if !isDebug {
				c.Response().Header().Set(echo.HeaderContentSecurityPolicy, csp)
			}

			return next(c)
		}
	}
}
