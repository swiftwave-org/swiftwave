package config

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
)

// Config : hold all configuration
type Config struct {
	LocalConfig  *local_config.Config
	SystemConfig *system_config.SystemConfig
}

// Fetch : fetch all configuration
func Fetch() (*Config, error) {
	// fetch local config first
	localConfig, err := local_config.Fetch()
	if err != nil {
		return nil, err
	}
	// fetch db client
	dbClient, err := db.GetClient(localConfig, 0)
	if err != nil {
		return nil, err
	}
	// fetch system config
	systemConfig, err := system_config.Fetch(dbClient)
	if err != nil {
		return nil, err
	}
	return &Config{
		LocalConfig:  localConfig,
		SystemConfig: systemConfig,
	}, nil
}

func ImageRegistryEndpoint(config *Config) (string, error) {
	if config.SystemConfig.ImageRegistryConfig.IsConfigured() {
		return config.SystemConfig.ImageRegistryConfig.Endpoint, nil
	}
	return config.LocalConfig.GetRegistryURL(), nil
}

func ImageRegistryUsername(config *Config) (string, error) {
	if config.SystemConfig.ImageRegistryConfig.IsConfigured() {
		return config.SystemConfig.ImageRegistryConfig.Username, nil
	}
	return config.LocalConfig.LocalImageRegistryConfig.Username, nil
}

func ImageRegistryPassword(config *Config) (string, error) {
	if config.SystemConfig.ImageRegistryConfig.IsConfigured() {
		return config.SystemConfig.ImageRegistryConfig.Password, nil
	}
	return config.LocalConfig.LocalImageRegistryConfig.Password, nil
}
