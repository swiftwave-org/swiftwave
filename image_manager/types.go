package imagemanager

type Manager struct {
	Config      Config            `yaml:"config"`
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
	Name      string     `yaml:"name"`
	Variables map[string]Variable `yaml:"variables"`
}

type Variable struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
}

type Identifier struct {
	Selector []IdentifierSelector `yaml:"selector"`
}

type IdentifierSelector struct {
	File     string   `yaml:"file"`
	Keywords []string `yaml:"keywords"`
}
