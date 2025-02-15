package middleware

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"strings"
)

type AuthInfo struct {
	authorized bool
	userID     uint
	context    context.Context
	db         *gorm.DB
}

func (a AuthInfo) IsAuthorized() bool {
	return a.authorized
}

func (a AuthInfo) GetUserID() uint {
	return a.userID
}

func (a AuthInfo) GetUser() (core.User, error) {
	if !a.authorized || a.userID == 0 || a.db == nil {
		return core.User{}, errors.New("unauthorized")
	}
	user, err := core.FindUserByID(a.context, *a.db, a.userID)
	if err != nil {
		return core.User{}, errors.New("user not found")
	}
	return user, nil
}

// AuthResolverMiddleware will add reference of auth info
// It will allow unauthenticated access as well
// Handler should verify requests
func AuthResolverMiddleware(dbClient *gorm.DB) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.Compare(c.Request().URL.Path, "/") == 0 ||
				strings.HasPrefix(c.Request().URL.Path, "/healthcheck") ||
				strings.HasPrefix(c.Request().URL.Path, "/.well-known") ||
				strings.HasPrefix(c.Request().URL.Path, "/webhook") ||
				strings.HasPrefix(c.Request().URL.Path, "/dashboard") ||
				strings.HasPrefix(c.Request().URL.Path, "/playground") {
				return next(c)
			}

			ctx := c.Request().Context()

			// Authenticate request
			sessionId, err := c.Cookie("session_id")
			if err != nil {
				c.Set("auth", AuthInfo{
					authorized: false,
					userID:     0,
					context:    ctx,
					db:         dbClient,
				})
			} else {
				userId, err := core.FetchUserIDBySessionID(ctx, *dbClient, sessionId.Value)
				if err != nil {
					c.Set("auth", AuthInfo{
						authorized: false,
						userID:     0,
						context:    ctx,
						db:         dbClient,
					})
				} else {
					c.Set("auth", AuthInfo{
						authorized: true,
						userID:     userId,
						context:    ctx,
						db:         dbClient,
					})
				}
			}
			return next(c)
		}
	}
}
