package core

import (
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
)

func createDbClient(dsn string) (*gorm.DB, error) {
	// Initiating database client
	dbDialect := postgres.Open(dsn)
	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		dbClient, err := gorm.Open(dbDialect, &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			if i == maxAttempts-1 {
				return nil, err
			}
			log.Println("Failed to connect to database. Retrying in 10 seconds...")
			time.Sleep(10 * time.Second)
			continue
		}
		return dbClient, nil
	}
	log.Println("Failed to connect to database. Retried 5 times. Exiting...")
	return nil, errors.New("failed to connect to database")
}
