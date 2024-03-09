package system_config

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
)

var config *core.SystemConfig

func Fetch(db *gorm.DB) (*core.SystemConfig, error) {
	if config != nil {
		return config, nil
	}
	// fetch first record
	var record core.SystemConfig
	tx := db.First(&record)
	if tx.Error != nil {
		return nil, tx.Error
	}
	config = &record
	return config, nil
}

func Update(db *gorm.DB, config *core.SystemConfig) error {
	tx := db.Save(config)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
