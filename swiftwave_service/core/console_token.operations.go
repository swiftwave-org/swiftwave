package core

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/labstack/gommon/random"
	"gorm.io/gorm"
	"time"
)

func GenerateConsoleTokenForServer(db gorm.DB, serverId uint) (*ConsoleToken, error) {
	// find server
	server, err := FetchServerByID(&db, serverId)
	if err != nil {
		return nil, errors.New("failed to fetch server")
	}
	// generate token
	record := &ConsoleToken{
		ID:        uuid.NewString(),
		Target:    ConsoleTargetTypeServer,
		ServerID:  &server.ID,
		ExpiresAt: time.Now().Add(time.Minute * 1),
		Token:     random.String(64),
	}
	// save record
	tx := db.Create(record)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return record, nil
}

func GenerateConsoleTokenForApplication(db gorm.DB, applicationId string) (*ConsoleToken, error) {
	// find application
	application := &Application{}
	err := application.FindById(context.TODO(), db, applicationId)
	if err != nil {
		return nil, errors.New("failed to fetch application")
	}
	// generate token
	record := &ConsoleToken{
		ID:            uuid.NewString(),
		Target:        ConsoleTargetTypeApplication,
		ApplicationID: &application.ID,
		ExpiresAt:     time.Now().Add(time.Minute * 1),
		Token:         random.String(64),
	}
	// save record
	tx := db.Create(record)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return record, nil
}

func FindConsoleToken(db gorm.DB, id string, token string) (*ConsoleToken, error) {
	// read from DB
	record := &ConsoleToken{}
	tx := db.First(record, "id = ? AND token = ?", id, token)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// delete from DB (defer)
	defer func() {
		_ = db.Delete(record)
	}()
	// check if expired
	if record.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	return record, nil
}
