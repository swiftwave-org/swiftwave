package core

import (
	"gorm.io/gorm"
)

func MigrateDatabase(dbClient *gorm.DB) {
	// Migrate the schema
	err := dbClient.AutoMigrate(
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
	)
	if err != nil {
		panic("failed to migrate database")
	}
}
