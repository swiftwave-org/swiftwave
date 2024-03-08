package system_config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func ReadFromFile(path string) (*Config, error) {
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
		config.ServiceConfig.BindAddress = DefaultBindAddress
	}
	if config.ServiceConfig.BindPort == 0 {
		config.ServiceConfig.BindPort = DefaultBindPort
	}
	if config.ServiceConfig.ManagementNodeAddress == "" {
		return errors.New("management_node_address is required in config")
	}
	config.ServiceConfig.SocketPathDirectory = DefaultSocketPathDirectory
	config.ServiceConfig.DataDirectory = DefaultDataDirectory
	config.ServiceConfig.NetworkName = DefaultNetworkName
	config.ServiceConfig.HAProxyServiceName = DefaultHAProxyServiceName
	config.ServiceConfig.HAProxyUnixSocketPath = DefaultHAProxyUnixSocketPath
	config.ServiceConfig.HAProxyDataDirectoryPath = DefaultHAProxyDataDirectoryPath
	config.ServiceConfig.UDPProxyServiceName = DefaultUDPProxyServiceName
	config.ServiceConfig.UDPProxyDataDirectoryPath = DefaultUDPProxyDataDirectoryPath
	return nil
}
