package db

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"gorm.io/gorm"
)

func MigrateDatabase(client *gorm.DB) error {
	sqlDb, err := client.DB()
	if err != nil {
		logger.DatabaseLoggerError.Println("Failed to create migrate instance")
		logger.DatabaseLoggerError.Println(err)
		return errors.New("failed to migrate database")
	}
	driver, err := postgres.WithInstance(sqlDb, &postgres.Config{})
	if err != nil {
		logger.DatabaseLoggerError.Println("Failed to create migrate instance")
		logger.DatabaseLoggerError.Println(err)
		return errors.New("failed to migrate database")
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://swiftwave_service/db/migrations",
		"postgres", driver)
	if err != nil {
		logger.DatabaseLoggerError.Println("Failed to create migrate instance")
		logger.DatabaseLoggerError.Println(err)
		return errors.New("failed to migrate database")
	}
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.DatabaseLogger.Println("No change in database schema")
			return nil
		}
		logger.DatabaseLoggerError.Println("Failed to migrate database")
		logger.DatabaseLoggerError.Println(err)
		return errors.New("failed to migrate database")
	}
	return nil
}
