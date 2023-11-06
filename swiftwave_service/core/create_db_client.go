package core

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func createDbClient() (*gorm.DB, error) {
	// Initiating database client
	dbDialect := postgres.Open(os.Getenv("POSTGRESQL_DSN"))
	client, err := gorm.Open(dbDialect, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return client, nil
}
