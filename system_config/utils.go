package system_config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func (p PostgresqlConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=%s", p.Host, p.Port, p.User, p.Password, p.Database, p.TimeZone, p.SSLMode)
}

func (a AMQPConfig) URI() string {
	return fmt.Sprintf("%s://%s:%s@%s", a.Protocol, a.User, a.Password, a.Host)
}

func (config *Config) String() (string, error) {
	// marshal to yaml
	out, err := yaml.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (config *Config) WriteToFile(path string) error {
	// marshal to yaml
	out, err := config.String()
	if err != nil {
		return err
	}
	// write to file
	err = os.WriteFile(path, []byte(out), 0644)
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
