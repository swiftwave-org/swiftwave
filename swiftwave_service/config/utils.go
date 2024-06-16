package config

import (
	"github.com/lib/pq"
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

	// add swiftwave, registry and db port in restricted ports if not configured
	if systemConfig != nil && localConfig != nil {
		if localConfig.ServiceConfig.BindPort != 80 && localConfig.ServiceConfig.BindPort != 443 {
			if !isPortAdded(localConfig.ServiceConfig.BindPort, systemConfig.RestrictedPorts) {
				systemConfig.RestrictedPorts = append(systemConfig.RestrictedPorts, int64(localConfig.ServiceConfig.BindPort))
			}
		}
		if localConfig.PostgresqlConfig.RunLocalPostgres {
			if !isPortAdded(localConfig.PostgresqlConfig.Port, systemConfig.RestrictedPorts) {
				systemConfig.RestrictedPorts = append(systemConfig.RestrictedPorts, int64(localConfig.PostgresqlConfig.Port))
			}
		}
		if !systemConfig.ImageRegistryConfig.IsConfigured() {
			if !isPortAdded(localConfig.LocalImageRegistryConfig.Port, systemConfig.RestrictedPorts) {
				systemConfig.RestrictedPorts = append(systemConfig.RestrictedPorts, int64(localConfig.LocalImageRegistryConfig.Port))
			}
		}
	}

	return &Config{
		LocalConfig:  localConfig,
		SystemConfig: systemConfig,
	}, nil
}

func (config *Config) ImageRegistryURI() string {
	if config.SystemConfig.ImageRegistryConfig.IsConfigured() {
		return config.SystemConfig.ImageRegistryConfig.URI()
	}
	return config.LocalConfig.GetRegistryURL()
}

func (config *Config) ImageRegistryUsername() string {
	if config.SystemConfig.ImageRegistryConfig.IsConfigured() {
		return config.SystemConfig.ImageRegistryConfig.Username
	}
	return config.LocalConfig.LocalImageRegistryConfig.Username
}

func (config *Config) ImageRegistryPassword() string {
	if config.SystemConfig.ImageRegistryConfig.IsConfigured() {
		return config.SystemConfig.ImageRegistryConfig.Password
	}
	return config.LocalConfig.LocalImageRegistryConfig.Password
}

// private functions
func isPortAdded(port int, ports pq.Int64Array) bool {
	for _, p := range ports {
		if int(p) == port {
			return true
		}
	}
	return false
}
