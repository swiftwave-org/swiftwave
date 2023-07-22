package imagemanager

import (
	_ "embed"
	"errors"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var fileByte []byte

func (m *Manager) Init() error {
	err := yaml.Unmarshal([]byte(fileByte), &m.Config)
	if err != nil {
		return errors.New("failed to unmarshal config.yaml")
	}
	return nil
}