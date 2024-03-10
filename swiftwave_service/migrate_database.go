package swiftwave

import (
	"errors"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"log"
)

func MigrateDatabase(dbClient *gorm.DB) error {
	// Migrate the schema
	err := dbClient.AutoMigrate(
		&system_config.SystemConfig{},
		&core.SystemLog{},
		&core.Server{},
		&core.ServerLog{},
		&core.User{},
		&core.Domain{},
		&core.RedirectRule{},
		&core.PersistentVolume{},
		&core.Application{},
		&core.GitCredential{},
		&core.ImageRegistryCredential{},
		&core.IngressRule{},
		&core.EnvironmentVariable{},
		&core.PersistentVolumeBinding{},
		&core.Deployment{},
		&core.BuildArg{},
		&core.DeploymentLog{},
		&SSL.KeyAuthorizationToken{},
		&core.PersistentVolumeBackup{},
		&core.PersistentVolumeRestore{},
	)
	if err != nil {
		log.Println(err)
		return errors.New("failed to migrate database \n" + err.Error())
	}
	return nil
}
