package dockerconfiggenerator

type Manager struct {
	Config          Config            `yaml:"config"`
	DockerTemplates map[string]string `yaml:"docker_files"`
}

type Config struct {
	ServiceOrder []string                `yaml:"service_order"`
	LookupFiles  []string                `yaml:"lookup_files"`
	Services     map[string]Service      `yaml:"services"`
	Templates    map[string]Template     `yaml:"templates"`
	Identifiers  map[string][]Identifier `yaml:"identifiers"`
}

type Service struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type Template struct {
	Name      string              `yaml:"name"`
	Variables map[string]Variable `yaml:"variables"`
}

type Variable struct {
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description" json:"description"`
	Default     string `yaml:"default" json:"default"`
}

type Identifier struct {
	Selectors  []IdentifierSelector `yaml:"selectors"`
	Extensions []string             `yaml:"extensions"`
}

type IdentifierSelector struct {
	File     string   `yaml:"file"`
	Keywords []string `yaml:"keywords"`
}

// DockerFile Config
type DockerFileConfig struct {
	DetectedService string              `json:"detected_service"`
	DockerFile      string              `json:"docker_file"`
	Variables       map[string]Variable `json:"variables"`
}
