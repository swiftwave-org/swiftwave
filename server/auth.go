package server

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Init function
func (server *Server) InitAuthRestAPI() {
	server.ECHO_SERVER.POST("/auth/login", server.authLogin)
	server.ECHO_SERVER.POST("/auth/verify", server.authVerify)
	server.ECHO_SERVER.POST("/auth/logout", server.authLogout)
	server.ECHO_SERVER.GET("/auth/ws/token", server.authWsToken)
}

// REST API functions

// GET /auth/ws/token
func (server *Server) authWsToken(c echo.Context) error {
	// generate websocket token
	token, err := server.generateWebsocketToken()
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   "internal server error",
			"message": "internal server error",
		})
	}
	// return token
	return c.JSON(200, map[string]interface{}{
		"token":   token,
		"message": "token generated successfully",
	})
}

// POST /auth/login
func (server *Server) authLogin(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	if username == "" || password == "" {
		return c.JSON(400, map[string]interface{}{
			"error":   "username and password are required",
			"message": "username and password are required",
		})
	}
	// fetch username and hashed password from env
	envUsername := os.Getenv("ADMIN_USERNAME")
	envHashedBase64Password := os.Getenv("ADMIN_PASSWORD")
	envHashedPasswordBytes, err := base64.StdEncoding.DecodeString(envHashedBase64Password)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   "internal server error",
			"message": "internal server error",
		})
	}
	if err := bcrypt.CompareHashAndPassword(envHashedPasswordBytes, []byte(password)); err != nil || username != envUsername {
		return c.JSON(401, map[string]interface{}{
			"error":   "invalid username or password",
			"message": "invalid username or password",
		})
	}
	// generate random session token
	randomToken, err := generateLongRandomString(64)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   "internal server error",
			"message": "internal server error",
		})
	}
	// store session token in memory
	server.SESSION_TOKENS[randomToken] = time.Time.Add(time.Now(), time.Duration(server.SESSION_TOKEN_EXPIRY_MINUTES)*time.Minute)
	// Try to set a cookie with authorization token
	cookie := new(http.Cookie)
	cookie.Name = "authorization"
	cookie.Value = randomToken
	cookie.Expires = time.Now().Add(time.Duration(server.SESSION_TOKEN_EXPIRY_MINUTES) * time.Minute)
	cookie.Path = "/"
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	// return session token
	return c.JSON(200, map[string]interface{}{
		"token":   randomToken,
		"message": "login successful",
	})
}

// GET /auth/verify
func (server *Server) authVerify(c echo.Context) error {
	// does not required verification as middleware is already applied
	return c.JSON(200, map[string]interface{}{
		"message": "token is valid",
	})
}

// POST /auth/logout
func (server *Server) authLogout(c echo.Context) error {
	token := ""
	cookie, err := c.Cookie("authorization")
	if err == nil {
		token = cookie.Value
	} else {
		token = c.Request().Header.Get("authorization")
	}
	// delete session token from memory
	delete(server.SESSION_TOKENS, token)
	// return success message
	return c.JSON(200, map[string]interface{}{
		"message": "logout successful",
	})
}

// MIDDLWARES
func (server *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// get token from cookie
		token := ""
		cookie, err := c.Cookie("authorization")
		if err == nil {
			token = cookie.Value
		} else {
			token = c.Request().Header.Get("authorization")
		}
		// whitelist routes
		// - /auth/login
		// - /.well-known
		path := c.Request().URL.Path
		if path == "/auth/login" || strings.HasPrefix(path, "/.well-known") {
			return next(c)
		}
		// whitelist OPTIONS method
		if c.Request().Method == "OPTIONS" {
			return next(c)
		}
		// check if token is valid
		if _, ok := server.SESSION_TOKENS[token]; ok {
			return next(c)
		}
		// check if websocket endpoint
		if strings.HasPrefix(path, "/ws") {
			// fetch token from query params
			token := c.QueryParam("token")
			// validate token
			if server.verifyWebsocketToken(token) {
				return next(c)
			}
		}
		// return error
		return c.JSON(401, map[string]interface{}{
			"error":   "unauthorized access",
			"message": "unauthorized access",
		})
	}
}

// private functions
func generateLongRandomString(length int) (string, error) {
	numUUIDs := (length + 32) / 33 // Number of UUIDs needed to achieve desired length
	randomString := ""

	for i := 0; i < numUUIDs; i++ {
		uuidObj, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}

		randomString += strings.Replace(uuidObj.String(), "-", "", -1)
	}

	return randomString[:length], nil
}

// Verify if Websocket token is valid
func (server *Server) verifyWebsocketToken(token string) bool {
	if _, ok := server.WEBSOCKET_TOKENS[token]; ok {
		// check if token is expired
		if time.Now().After(server.WEBSOCKET_TOKENS[token]) {
			// delete token
			delete(server.WEBSOCKET_TOKENS, token)
			return false
		}
		// delete token
		delete(server.WEBSOCKET_TOKENS, token)
		return true
	}
	return false
}

// Generate websocket token
func (server *Server) generateWebsocketToken() (string, error) {
	randomToken, err := generateLongRandomString(32)
	if err != nil {
		return "", err
	}
	// store session token in memory
	server.WEBSOCKET_TOKENS[randomToken] = time.Time.Add(time.Now(), time.Duration(server.WEBSOCKET_TOKEN_EXPIRY_MINUTES)*time.Minute)
	return randomToken, nil
}
