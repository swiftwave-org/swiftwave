package rest

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/xlzd/gotp"
)

// there is one http server per process, so the throttle is shared by every login handler
var loginAttempts = newLoginThrottle()

// POST /auth/login
func (server *Server) login(c echo.Context) error {
	// Get params
	username := c.FormValue("username")
	password := c.FormValue("password")
	totp := c.FormValue("totp")

	// Throttle keys, scoped per account and per source so neither one alone lets an
	// attacker keep guessing
	userKey := "user:" + strings.ToLower(strings.TrimSpace(username))
	ipKey := "ip:" + c.RealIP()
	totpUserKey := "totp:" + userKey
	totpIPKey := "totp:" + ipKey

	if wait := loginAttempts.lockedFor(userKey, ipKey, totpUserKey, totpIPKey); wait > 0 {
		return tooManyAttempts(c, wait)
	}

	// Check if user exists
	user, err := core.FindUserByUsername(c.Request().Context(), server.ServiceManager.DbClient, username)
	if err != nil {
		loginAttempts.fail(userKey, loginFailureThreshold, loginBaseLockout, loginMaxLockout)
		loginAttempts.fail(ipKey, loginFailureThreshold, loginBaseLockout, loginMaxLockout)
		return invalidCredentials(c)
	}

	// Check password
	if !user.CheckPassword(password) {
		loginAttempts.fail(userKey, loginFailureThreshold, loginBaseLockout, loginMaxLockout)
		loginAttempts.fail(ipKey, loginFailureThreshold, loginBaseLockout, loginMaxLockout)
		return invalidCredentials(c)
	}

	// Check totp
	// asked for only after the password is verified, so an unauthenticated caller cannot
	// discover which usernames exist or which of them have two factor enabled
	if user.TotpEnabled {
		if strings.Compare(totp, "") == 0 {
			return c.JSON(400, &LoginResponse{
				Message:      "two factor authentication is enabled, but totp is not provided",
				Token:        "",
				TotpRequired: true,
			})
		}
		totpRecord := gotp.NewDefaultTOTP(user.TotpSecret)
		if !totpRecord.Verify(totp, time.Now().Unix()) {
			// a six digit code is small enough to exhaust within one validity window,
			// so failures here are budgeted far more tightly than password failures
			loginAttempts.fail(totpUserKey, totpFailureThreshold, totpBaseLockout, totpMaxLockout)
			loginAttempts.fail(totpIPKey, totpFailureThreshold, totpBaseLockout, totpMaxLockout)
			return c.JSON(400, &LoginResponse{
				Message:      "invalid totp",
				Token:        "",
				TotpRequired: false,
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

	loginAttempts.clear(userKey, ipKey, totpUserKey, totpIPKey)

	// Return token
	return c.JSON(200, &LoginResponse{
		Message:      "success",
		Token:        token,
		TotpRequired: false,
	})
}

// private functions
// invalidCredentials keeps the response identical for an unknown user and a wrong
// password, so the endpoint cannot be used to enumerate usernames
func invalidCredentials(c echo.Context) error {
	return c.JSON(400, &LoginResponse{
		Message:      "invalid credentials",
		Token:        "",
		TotpRequired: false,
	})
}

func tooManyAttempts(c echo.Context, wait time.Duration) error {
	seconds := int(math.Ceil(wait.Seconds()))
	c.Response().Header().Set("Retry-After", strconv.Itoa(seconds))
	return c.JSON(429, &LoginResponse{
		Message:      "too many failed login attempts, try again in " + strconv.Itoa(seconds) + " seconds",
		Token:        "",
		TotpRequired: false,
	})
}
