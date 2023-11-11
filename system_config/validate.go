package system_config

import (
	"errors"
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
		return nil, errors.New("failed to parse config file")
	}
	// validate config
	err = config.Validate()
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (config *Config) Validate() error {
	// If mode is cluster, pubsub and task queue must be remote
	if config.Mode == Cluster {
		if config.PubSubConfig.Mode != RemotePubSub {
			return errors.New("in cluster mode, pubsub must be remote, configure a redis server in config file")
		}
		if config.TaskQueueConfig.Mode != RemoteTaskQueue {
			return errors.New("in cluster mode, task queue must be remote, configure a redis server in config file")
		}
	}

	return nil
}
