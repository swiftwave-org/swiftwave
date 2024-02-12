package rest

import "github.com/labstack/echo/v4"

// GET /healthcheck
func (server *Server) healthcheck(c echo.Context) error {
	return c.String(200, "OK")
}
