package Manager

import (
	"github.com/labstack/echo/v4"
	"io"
)

// Required for http-01 verification
// - Path /.well-known/acme-challenge/{token}
func (s Manager) acmeHttpHandler(c echo.Context) error {
	token := c.Param("token")
	fullToken := s.fetchKeyAuthorization(token)
	_, err := io.WriteString(c.Response().Writer, fullToken)
	return err
}

// Required for pre-authorization
// Check if the domain is pointing to the server
// - Path /.well-known/pre-authorize/
func (s Manager) dnsConfigurationPreAuthorizeHttpHandler(c echo.Context) error {
	_, err := io.WriteString(c.Response().Writer, "OK")
	return err
}

// Init http handlers
func (s Manager) InitHttpHandlers(e *echo.Echo) {
	e.GET("/.well-known/acme-challenge/:token", s.acmeHttpHandler)
	e.GET("/.well-known/pre-authorize", s.dnsConfigurationPreAuthorizeHttpHandler)
}
