package imagemanager

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
	for _, template := range m.Config.Templates {
		data, err := templateFolder.ReadFile("templates/"+template.Name)
		if err != nil {
			return errors.New("failed to read template : "+template.Name)
		}
		m.DockerTemplates[template.Name] = string(data)
	}
	return nil
}