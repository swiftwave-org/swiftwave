package core

import (
	"gorm.io/gorm"
)

func MigrateDatabase(dbClient *gorm.DB) {
	// Migrate the schema
	err := dbClient.AutoMigrate(
		&GitCredential{},
		&ImageRegistryCredential{},
		&Domain{},
		&IngressRule{},
		&RedirectRule{},
		&Application{},
		&Deployment{},
		&DeploymentLog{},
	)
	if err != nil {
		panic("failed to migrate database")
	}
}
