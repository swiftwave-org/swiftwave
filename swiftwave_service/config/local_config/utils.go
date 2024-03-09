package local_config

import (
	"errors"
	"fmt"
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
	err = fillDefaults(&config)
	return &config, nil
}

func fillDefaults(config *Config) error {
	if config.ServiceConfig.BindAddress == "" {
		config.ServiceConfig.BindAddress = defaultBindAddress
	}
	if config.ServiceConfig.BindPort == 0 {
		config.ServiceConfig.BindPort = defaultBindPort
	}
	if config.ServiceConfig.ManagementNodeAddress == "" {
		return errors.New("management_node_address is required in config")
	}
	config.ServiceConfig.SocketPathDirectory = defaultSocketPathDirectory
	config.ServiceConfig.DataDirectory = defaultDataDirectory
	config.ServiceConfig.NetworkName = defaultNetworkName
	config.ServiceConfig.HAProxyServiceName = defaultHAProxyServiceName
	config.ServiceConfig.HAProxyUnixSocketPath = defaultHAProxyUnixSocketPath
	config.ServiceConfig.HAProxyDataDirectoryPath = defaultHAProxyDataDirectoryPath
	config.ServiceConfig.UDPProxyServiceName = defaultUDPProxyServiceName
	config.ServiceConfig.UDPProxyDataDirectoryPath = defaultUDPProxyDataDirectoryPath
	config.ServiceConfig.SSLCertDirectoryPath = defaultSSLCertDirectoryPath
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
