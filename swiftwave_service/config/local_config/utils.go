package local_config

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

var config *Config

func Fetch() (*Config, error) {
	if config != nil {
		return config, nil
	}
	c, e := readConfigFile(LocalConfigPath)
	if e != nil {
		return nil, e
	}
	config = c
	config.Version = softwareVersion
	return config, nil
}

func Update(config *Config) error {
	// marshal to yaml
	out, err := config.String()
	if err != nil {
		return err
	}
	// write to file
	err = os.WriteFile(LocalConfigPath, []byte(out), 0600)
	if err != nil {
		return err
	}
	return nil
}

func (config *Config) DeepCopy() *Config {
	// marshal to yaml
	out, _ := config.String()
	// unmarshal to new object
	newConfig := &Config{}
	err := yaml.Unmarshal([]byte(out), newConfig)
	if err != nil {
		return nil
	}
	return newConfig
}

func readConfigFile(path string) (*Config, error) {
	// check if exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("config file does not exist at > " + path)
	}
	// create a reader
	reader, err := os.Open(path)
	if err != nil {
		return nil, errors.New("failed to open config file")
	}
	// read file
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.New("failed to read config file")
	}
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			panic(err)
		}
	}(reader)
	// create config
	config := Config{}
	// parse yaml
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("failed to parse config file")
	}
	// validate config
	_ = FillDefaults(&config)
	return &config, nil
}

func FillDefaults(config *Config) error {
	if config.ServiceConfig.BindAddress == "" {
		config.ServiceConfig.BindAddress = defaultBindAddress
	}
	if config.ServiceConfig.BindPort == 0 {
		config.ServiceConfig.BindPort = defaultBindPort
	}
	if config.ServiceConfig.ManagementNodeAddress == "" {
		return errors.New("management_node_address is required in config")
	}
	if config.LocalImageRegistryConfig.Port == 0 {
		config.LocalImageRegistryConfig.Port = defaultImageRegistryPort
	}
	config.ServiceConfig.SocketPathDirectory = defaultSocketPathDirectory
	config.ServiceConfig.DataDirectory = defaultDataDirectory
	config.ServiceConfig.LocalPostgresDataDirectory = defaultLocalPostgresDataDirectory
	config.ServiceConfig.TarballDirectoryPath = defaultTarballDirectoryPath
	config.ServiceConfig.NetworkName = defaultNetworkName
	config.ServiceConfig.HAProxyServiceName = defaultHAProxyServiceName
	config.ServiceConfig.HAProxyUnixSocketDirectory = defaultHAProxyUnixSocketDirectory
	config.ServiceConfig.HAProxyUnixSocketPath = defaultHAProxyUnixSocketPath
	config.ServiceConfig.HAProxyDataDirectoryPath = defaultHAProxyDataDirectoryPath
	config.ServiceConfig.UDPProxyServiceName = defaultUDPProxyServiceName
	config.ServiceConfig.UDPProxyUnixSocketDirectory = defaultUDPProxyUnixSocketDirectory
	config.ServiceConfig.UDPProxyUnixSocketPath = defaultUDPProxyUnixSocketPath
	config.ServiceConfig.UDPProxyDataDirectoryPath = defaultUDPProxyDataDirectoryPath
	config.ServiceConfig.SSLCertDirectoryPath = defaultSSLCertDirectoryPath
	config.ServiceConfig.LocalImageRegistryDirectoryPath = defaultLocalImageRegistryDirectoryPath
	config.ServiceConfig.LogDirectoryPath = LogDirectoryPath
	config.ServiceConfig.InfoLogFilePath = InfoLogFilePath
	config.ServiceConfig.ErrorLogFilePath = ErrorLogFilePath
	config.ServiceConfig.PVBackupDirectoryPath = defaultPVBackupDirectoryPath
	config.ServiceConfig.PVRestoreDirectoryPath = defaultPVRestoreDirectoryPath
	config.LocalImageRegistryConfig.CertPath = defaultLocalImageRegistryCertDirectoryPath
	config.LocalImageRegistryConfig.AuthPath = defaultLocalImageRegistryAuthDirectoryPath
	config.LocalImageRegistryConfig.DataPath = defaultLocalImageRegistryDataDirectoryPath
	return nil
}

func (p PostgresqlConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=%s", p.Host, p.Port, p.User, p.Password, p.Database, p.TimeZone, p.SSLMode)
}

func (config *Config) String() (string, error) {
	// marshal to yaml
	out, err := yaml.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (config *Config) GetRegistryURL() string {
	return fmt.Sprintf("%s:%d", config.ServiceConfig.ManagementNodeAddress, config.LocalImageRegistryConfig.Port)
}

func (l *LocalImageRegistryConfig) Htpasswd() (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(l.Password), 20)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", l.Username, string(hashedPassword)), nil
}
