package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/xlzd/gotp"
	"strings"
	"time"
)

// POST /auth/login
func (server *Server) login(c echo.Context) error {
	// Get params
	username := c.FormValue("username")
	password := c.FormValue("password")
	totp := c.FormValue("totp")

	// Check if user exists
	user, err := core.FindUserByUsername(c.Request().Context(), server.ServiceManager.DbClient, username)
	if err != nil {
		return c.JSON(400, &LoginResponse{
			Message:      "user does not exist",
			Token:        "",
			TotpRequired: false,
		})
	}

	// check if totp is enabled
	if user.TotpEnabled && strings.Compare(totp, "") != 0 {
		return c.JSON(400, &LoginResponse{
			Message:      "two factor authentication is enabled, but totp is not provided",
			Token:        "",
			TotpRequired: true,
		})
	}

	// Check password
	if !user.CheckPassword(password) {
		return c.JSON(400, &LoginResponse{
			Message:      "invalid password",
			Token:        "",
			TotpRequired: false,
		})
	}

	// Check totp
	if user.TotpEnabled {
		totpRecord := gotp.NewDefaultTOTP(user.TotpSecret)
		if totpRecord.Verify(totp, time.Now().Unix()) == false {
			return c.JSON(400, &LoginResponse{
				Message:      "invalid totp",
				Token:        "",
				TotpRequired: true,
			})
		}
	}

	// Generate jwt token
	token, err := user.GenerateJWT(server.Config.SystemConfig.JWTSecretKey)
	if err != nil {
		return c.JSON(500, &LoginResponse{
			Message:      "failed to generate jwt token",
			Token:        "",
			TotpRequired: false,
		})
	}

	// Return token
	return c.JSON(200, &LoginResponse{
		Message:      "success",
		Token:        token,
		TotpRequired: false,
	})
}
