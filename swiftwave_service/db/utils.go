package db

import (
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"sync"
	"time"
)

var dbClient *gorm.DB
var mutex = &sync.Mutex{}

func GetClient(config *local_config.Config, maxRetry uint) (*gorm.DB, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if dbClient != nil {
		return dbClient, nil
	}
	dbDialect := postgres.Open(config.PostgresqlConfig.DSN())
	logLevel := gormlogger.Error
	if config.IsDevelopmentMode {
		logLevel = gormlogger.Info
	}

	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: newCustomLogger(logger.DatabaseLogger, gormlogger.Config{
			SlowThreshold:             500 * time.Millisecond,
			Colorful:                  false,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		}),
	}
	currentRetry := uint(0)
	var db *gorm.DB
	var err error
	for maxRetry == 0 || currentRetry < maxRetry {
		db, err = gorm.Open(dbDialect, gormConfig)
		if err != nil {
			logger.DatabaseLoggerError.Println(err.Error())
			logger.DatabaseLogger.Println("Failed to connect to database. Retrying in 2 seconds...")
			time.Sleep(2 * time.Second)
		} else {
			dbClient = db
			break
		}
		currentRetry++
	}
	if err != nil {
		return nil, errors.New("failed to connect to database")
	}
	return dbClient, nil
}
