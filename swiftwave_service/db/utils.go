package db

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"time"
)

var dbClient *gorm.DB

func GetClient(config *local_config.Config) *gorm.DB {
	if dbClient != nil {
		return dbClient
	}
	dbDialect := postgres.Open(config.PostgresqlConfig.DSN())
	var db *gorm.DB
	var err error
	for {
		if config.IsDevelopmentMode {
			db, err = gorm.Open(dbDialect, &gorm.Config{
				SkipDefaultTransaction: true,
			})
		} else {
			db, err = gorm.Open(dbDialect, &gorm.Config{
				SkipDefaultTransaction: true,
				Logger:                 logger.Default.LogMode(logger.Silent),
			})
		}
		if err != nil {
			log.Println("Failed to connect to database. Retrying in 2 seconds...")
			time.Sleep(2 * time.Second)
		} else {
			dbClient = db
			break
		}
	}
	return dbClient
}
