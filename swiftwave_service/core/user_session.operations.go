package core

import (
	"context"
	"errors"
	"github.com/labstack/gommon/random"
	"gorm.io/gorm"
	"time"
)

// This file contains the operations for the User Sessions model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

// CreateSession : create session for user
func CreateSession(ctx context.Context, db gorm.DB, user User) (string, error) {
	sessionRecord := &UserSession{
		UserID:    user.ID,
		SessionID: random.String(128),
		ExpiresAt: time.Now().Add(time.Hour * 720),
	}
	// Create record
	err := db.Create(sessionRecord).Error
	if err != nil {
		return "", err
	}
	return sessionRecord.SessionID, nil
}

// GetUserIDBySessionID : get user by session id
func GetUserIDBySessionID(ctx context.Context, db gorm.DB, sessionID string) (uint, error) {
	var session UserSession
	err := db.Where("session_id = ?", sessionID).Select("user_id").First(&session).Error
	if err != nil {
		return 0, errors.New("invalid session")
	}
	return session.UserID, nil
}
