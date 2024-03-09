package system_config

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
)

var config *core.SystemConfig
var configVersion uint = -1

func Fetch(db *gorm.DB) (*core.SystemConfig, error) {
	if config != nil {
		// Fetch the latest version of the config
		var record core.SystemConfig
		tx := db.First(&record).Select("config_version")
		if tx.Error != nil {
			return nil, tx.Error
		}
		// if the version is the same, return the cached config
		if record.ConfigVersion == configVersion {
			return config, nil
		}
	}
	// fetch first record
	var record core.SystemConfig
	tx := db.First(&record)
	if tx.Error != nil {
		return nil, tx.Error
	}
	config = &record
	configVersion = record.ConfigVersion
	return config, nil
}

func Update(db *gorm.DB, config *core.SystemConfig) error {
	tx := db.Save(config)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
