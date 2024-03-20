package core

import (
	"errors"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"gorm.io/gorm"
	"log"
)

func MigrateDatabase(dbClient *gorm.DB) error {
	// Migrate the schema
	err := dbClient.AutoMigrate(
		&system_config.SystemConfig{},
		&SystemLog{},
		&Server{},
		&ServerLog{},
		&User{},
		&Domain{},
		&RedirectRule{},
		&PersistentVolume{},
		&Application{},
		&GitCredential{},
		&ImageRegistryCredential{},
		&IngressRule{},
		&EnvironmentVariable{},
		&PersistentVolumeBinding{},
		&Deployment{},
		&BuildArg{},
		&DeploymentLog{},
		&SSL.KeyAuthorizationToken{},
		&PersistentVolumeBackup{},
		&PersistentVolumeRestore{},
		&ConsoleToken{},
		&AnalyticsServiceToken{},
		&ServerResourceStat{},
		&ApplicationServiceResourceStat{},
	)
	if err != nil {
		log.Println(err)
		return errors.New("failed to migrate database \n" + err.Error())
	}
	return nil
}
