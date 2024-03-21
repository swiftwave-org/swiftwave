package rest

import "github.com/labstack/echo/v4"

// GET /version
func (server *Server) version(c echo.Context) error {
	return c.String(200, server.Config.LocalConfig.Version)
}
