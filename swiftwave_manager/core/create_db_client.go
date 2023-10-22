package core

import (
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

func createDbClient() (*gorm.DB, error) {
	// Initiating database client
	dbType := os.Getenv("DATABASE_TYPE")
	var dbDialect gorm.Dialector
	if dbType == "postgres" {
		dbDialect = postgres.Open(os.Getenv("POSTGRESQL_URI"))
	} else if dbType == "sqlite" {
		dbDialect = sqlite.Open(os.Getenv("SQLITE_DATABASE"))
	} else {
		return nil, errors.New("unknown database type")
	}
	client, err := gorm.Open(dbDialect, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return client, nil
}
