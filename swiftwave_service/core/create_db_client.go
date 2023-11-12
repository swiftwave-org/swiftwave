package core

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createDbClient(dsn string) (*gorm.DB, error) {
	// Initiating database client
	dbDialect := postgres.Open(dsn)
	client, err := gorm.Open(dbDialect, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return client, nil
}
