package rest

import "github.com/labstack/echo/v4"

// GET /verify-auth
func (server *Server) verifyAuth(c echo.Context) error {
	return c.String(200, "OK")
}
