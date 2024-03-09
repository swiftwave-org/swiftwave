package db

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

var dbClient *gorm.DB

func GetClient(config *local_config.Config) *gorm.DB {
	if dbClient != nil {
		return dbClient
	}
	dbDialect := postgres.Open(config.PostgresqlConfig.DSN())
	logLevel := gormlogger.Error
	if config.IsDevelopmentMode {
		logLevel = gormlogger.Info
	}
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: gormlogger.New(logger.DatabaseLogger, gormlogger.Config{
			SlowThreshold: 500 * time.Millisecond,
			Colorful:      false,
			LogLevel:      logLevel,
		}),
	}
	var db *gorm.DB
	var err error
	for {

		db, err = gorm.Open(dbDialect, gormConfig)
		if err != nil {
			logger.DatabaseLogger.Println("Failed to connect to database. Retrying in 2 seconds...")
			time.Sleep(2 * time.Second)
		} else {
			dbClient = db
			break
		}
	}
	return dbClient
}
