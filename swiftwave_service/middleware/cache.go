package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// CacheMiddleware > Cache middleware
// Cache JS, CSS and PNG files for 1 year, as if static content changes, the uri also changes
// So setting cache-control header to max-age to 1 year
// + it will also set etag header to the file name
func CacheMiddleware() func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasSuffix(c.Request().RequestURI, ".js") || strings.HasSuffix(c.Request().RequestURI, ".css") || strings.HasSuffix(c.Request().RequestURI, ".png") || strings.HasSuffix(c.Request().RequestURI, ".ttf") {
				s := strings.Split(c.Request().RequestURI, "/")
				etag := s[len(s)-1]
				c.Response().Header().Set("Etag", etag)
				c.Response().Header().Set("Cache-Control", "max-age=31536000")
				if match := c.Request().Header.Get("If-None-Match"); match != "" {
					if strings.Contains(match, etag) {
						return c.NoContent(http.StatusNotModified)
					}
				}
			}
			return next(c)
		}
	}
}
