package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

// POST /auth/login
func (server *Server) login(c echo.Context) error {
	// Get params
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Check if user exists
	user, err := core.FindUserByUsername(c.Request().Context(), server.ServiceManager.DbClient, username)
	if err != nil {
		return c.JSON(400, map[string]string{
			"message": "user does not exist",
		})
	}

	// Check password
	if !user.CheckPassword(password) {
		return c.JSON(400, map[string]string{
			"message": "incorrect password",
		})
	}

	// Generate jwt token
	token, err := user.GenerateJWT(server.Config.ServiceConfig.JwtSecretKey)
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "failed to generate jwt token",
		})
	}

	// Return token
	return c.JSON(200, map[string]string{
		"token": token,
	})
}
