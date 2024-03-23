package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/gommon/random"
	"gorm.io/gorm"
)

func ValidateAnalyticsServiceToken(ctx context.Context, db gorm.DB, id string, token string) (verified bool, serverHostName string, err error) {
	// fetch token from database
	var tokenData AnalyticsServiceToken
	tx := db.Where("id = ? AND token = ?", id, token).First(&tokenData)
	if tx.Error != nil {
		return false, "", tx.Error
	}
	// fetch hostname from database
	var server Server
	tx = db.Select("host_name").Where("id = ?", tokenData.ServerID).First(&server)
	if tx.Error != nil {
		return false, "", tx.Error
	}
	return true, server.HostName, nil
}

func FetchAnalyticsServiceToken(ctx context.Context, db gorm.DB, serverId uint) (*AnalyticsServiceToken, error) {
	// check if token exists
	var tokenData AnalyticsServiceToken
	tx := db.Where("server_id = ?", serverId).First(&tokenData)
	if tx.Error == nil {
		return &tokenData, nil
	}
	// create a new token
	tokenData = AnalyticsServiceToken{
		ID:       uuid.NewString(),
		Token:    random.String(32),
		ServerID: serverId,
	}
	tx = db.Create(&tokenData)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &tokenData, nil
}

func (token *AnalyticsServiceToken) IDToken() (string, error) {
	if token == nil {
		return "", errors.New("token is nil")
	}
	return fmt.Sprintf("%s:%s", token.ID, token.Token), nil
}

func DeleteAnalyticsServiceToken(ctx context.Context, db gorm.DB, serverId uint) error {
	tx := db.Where("server_id = ?", serverId).Delete(&AnalyticsServiceToken{})
	if tx.Error != nil {
		// don't return error if token does not exist
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return tx.Error
		}
	}
	return nil
}

// RotateAnalyticsServiceToken : delete existing token and create a new token. [Recommended to use transaction]
func RotateAnalyticsServiceToken(ctx context.Context, db gorm.DB, serverId uint) (*AnalyticsServiceToken, error) {
	// delete existing token
	err := DeleteAnalyticsServiceToken(ctx, db, serverId)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	// create a new token
	return FetchAnalyticsServiceToken(ctx, db, serverId)
}
