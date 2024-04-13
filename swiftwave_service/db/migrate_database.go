package db

import (
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"gorm.io/gorm"
)

//go:embed all:migrations
var migrationFilesFS embed.FS

func MigrateDatabase(client *gorm.DB) error {
	migrationFSDriver, err := iofs.New(migrationFilesFS, "migrations")
	if err != nil {
		logger.DatabaseLoggerError.Println("Failed to create migration fs driver")
		logger.DatabaseLoggerError.Println(err)
		return errors.New("unable to parse migration files")
	}
	// Create db connection
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
	m, err := migrate.NewWithInstance(
		"iofs",
		migrationFSDriver,
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
