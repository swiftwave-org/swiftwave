package dockerconfiggenerator

import (
	"embed"
	"errors"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var fileByte []byte

//go:embed templates/*
var templateFolder embed.FS

func (m *Manager) Init() error {
	err := yaml.Unmarshal([]byte(fileByte), &m.Config)
	if err != nil {
		return errors.New("failed to unmarshal config.yaml")
	}
	// Load templates
	m.DockerTemplates = map[string]string{}
	for service, template := range m.Config.Templates {
		data, err := templateFolder.ReadFile("templates/"+template.Name)
		if err != nil {
			return errors.New("failed to read template : "+template.Name)
		}
		m.DockerTemplates[service] = string(data)
	}
	return nil
}
