package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func PrepareCSP(isDebug bool) echo.MiddlewareFunc {
	//var csp string
	csp := "default-src 'self'; base-uri 'self'; form-action 'self'; script-src 'self'; script-src-elem 'self' %s; script-src-attr 'self'; object-src 'self'; style-src 'self'; style-src-elem 'self' %s; style-src-attr 'self'; img-src 'self'; font-src 'self'; connect-src 'self' %s; media-src 'self'; frame-ancestors 'self'; frame-src 'self'; child-src 'self';"
	if isDebug {
		signinJS := "'sha256-CI/zle8B36PcsnRpJFDskeKWwPR91xYlzFCXfK2ri2Y='"
		script_src_elem := signinJS

		iconifyCSS := "'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=' 'sha256-G2K9ENiXaTIc4pmzLEOJB962ySgP2gMolWCZ6HJpU4I=' 'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=' 'sha256-G2K9ENiXaTIc4pmzLEOJB962ySgP2gMolWCZ6HJpU4I='"
		htmxCSS := "'sha256-2Zmme+3cWvmG8lapM3WvEkAyYA3671LVoN107gkAU4g='"
		style_src_elem := iconifyCSS + " " + htmxCSS

		iconifySrc := "https://api.iconify.design https://api.simplesvg.com https://api.unisvg.com https://api.iconify.design https://api.simplesvg.com https://api.unisvg.com"
		connect_src := iconifySrc

		csp = fmt.Sprintf(csp, script_src_elem, style_src_elem, connect_src)
	}
	//const cspBase = "default-src 'self'; base-uri 'self'; form-action 'self'; script-src 'self'; script-src-elem 'self' 'sha256-CI/zle8B36PcsnRpJFDskeKWwPR91xYlzFCXfK2ri2Y='; script-src-attr 'self'; object-src 'self'; style-src 'self'; style-src-elem 'self' 'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=' 'sha256-G2K9ENiXaTIc4pmzLEOJB962ySgP2gMolWCZ6HJpU4I=' 'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=' 'sha256-G2K9ENiXaTIc4pmzLEOJB962ySgP2gMolWCZ6HJpU4I=' 'sha256-2Zmme+3cWvmG8lapM3WvEkAyYA3671LVoN107gkAU4g='; style-src-attr 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https://api.iconify.design; media-src 'self'; frame-ancestors 'self'; frame-src 'self'; child-src 'self';"

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Insert the hashes into the content security policy header
			c.Response().Header().Set(echo.HeaderContentSecurityPolicy, csp)

			return next(c)
		}
	}
}
