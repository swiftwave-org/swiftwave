package stack_parser

import (
	"errors"
	"github.com/hashicorp/go-set"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// KeyValuePair Generic key-value pair
// Support both map and slice of strings (key=value) formats
type KeyValuePair map[string]string

func (p *KeyValuePair) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*p = make(map[string]string, 0)

	switch reflect.TypeOf(raw).Kind() {
	case reflect.Map:
		for key, value := range raw.(map[string]interface{}) {
			if value != nil && reflect.TypeOf(value).Kind() == reflect.String {
				if val, ok := value.(string); ok {
					(*p)[key] = strings.TrimSpace(val)
				}
			} else {
				(*p)[key] = ""
			}
		}
	case reflect.Slice:
		for _, record := range raw.([]interface{}) {
			if record != nil && reflect.TypeOf(record).Kind() == reflect.String {
				recordString := record.(string)
				recordSplit := strings.SplitN(recordString, "=", 2)
				if len(recordSplit) > 0 {
					key := strings.TrimSpace(recordSplit[0])
					value := ""
					if len(recordSplit) == 2 {
						value = strings.TrimSpace(recordSplit[1])
					}
					(*p)[key] = value
				}
			}
		}
	default:
		return errors.New("invalid key-value pair")
	}
	return nil
}

// VolumeList List of volumes
type VolumeList []Volume

type Volume struct {
	Name          string `yaml:"name"`
	MountingPoint string `yaml:"mounting_point"`
}

func (v Volume) isNamedVolume() bool {
	return !strings.Contains(v.Name, "/")
}

func (v Volume) MarshalYAML() (interface{}, error) {
	return v.Name + ":" + v.MountingPoint, nil
}

func (l *VolumeList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*l = make([]Volume, 0)

	typeOf := reflect.TypeOf(raw).Kind()
	if typeOf != reflect.Slice {
		return nil
	}

	// iterate over the elements of the list
	for _, record := range raw.([]interface{}) {
		if record != nil && reflect.TypeOf(record).Kind() == reflect.String {
			volume := Volume{}
			recordString, ok := record.(string)
			if ok {
				recordSplit := strings.Split(recordString, ":")
				if len(recordSplit) >= 2 {
					volume.Name = strings.TrimSpace(recordSplit[0])
					volume.MountingPoint = strings.TrimSpace(recordSplit[1])
					*l = append(*l, volume)
				}
			}
		} else {
			return errors.New("invalid volume definition")
		}
	}

	// check if any volume is not named volume
	for _, volume := range *l {
		if !volume.isNamedVolume() {
			return errors.New("only named volumes are supported")
		}
	}

	return nil
}

// Command Command definition
type Command []string

func (c *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*c = make([]string, 0)

	switch reflect.TypeOf(raw).Kind() {
	case reflect.Slice:
		for _, record := range raw.([]interface{}) {
			if record != nil && reflect.TypeOf(record).Kind() == reflect.String {
				recordString := record.(string)
				*c = append(*c, recordString)
			}
		}
	case reflect.String:
		*c = append(*c, raw.(string))
	default:
		return errors.New("invalid command definition")
	}
	return nil
}

func (c *Command) String() string {
	return strings.Join(*c, " ")
}

// Stack Stack definition
type Stack struct {
	Services map[string]Service `yaml:"services"`
	Docs     *Docs              `yaml:"docs"`
}

type Service struct {
	Image                    string            `yaml:"image"`
	Deploy                   Deploy            `yaml:"deploy"`
	Volumes                  VolumeList        `yaml:"volumes"`
	Environment              KeyValuePair      `yaml:"environment"`
	CapAdd                   []string          `yaml:"cap_add"`
	Sysctls                  KeyValuePair      `yaml:"sysctls"`
	Command                  Command           `yaml:"command"`
	Configs                  []Config          `yaml:"configs"`
	PreferredServerHostnames []string          `yaml:"preferred_server_hostnames"`
	DockerProxyConfig        DockerProxyConfig `yaml:"docker_proxy_config"`
}

// DeploymentMode mode of deployment of application (replicated or global)
type DeploymentMode string

const (
	DeploymentModeReplicated DeploymentMode = "replicated"
	DeploymentModeGlobal     DeploymentMode = "global"
	DeploymentModeNone       DeploymentMode = ""
)

type Deploy struct {
	Mode      DeploymentMode `yaml:"mode"`
	Replicas  uint           `yaml:"replicas"`
	Resources Resources      `yaml:"resources"`
}

// Resources for the service
type Resources struct {
	Limits       ResourcesLimits       `yaml:"limits"`
	Reservations ResourcesReservations `yaml:"reservations"`
}

type ResourcesLimits struct {
	MemoryMB int `yaml:"memory"`
}

type ResourcesReservations struct {
	MemoryMB int `yaml:"memory"`
}

// Config for the service
type Config struct {
	Content      string `yaml:"content"`
	Uid          uint   `yaml:"uid"`
	Gid          uint   `yaml:"gid"`
	MountingPath string `yaml:"mounting_path"`
}

// Docs for the stack
type Docs struct {
	LogoURL           string                  `yaml:"logo_url"`
	Name              string                  `yaml:"name"`
	Description       string                  `yaml:"description"`
	ReadmeDescription string                  `yaml:"readme_description"`
	IframeVideoEmbed  string                  `yaml:"iframe_video_embed"`
	Variables         map[string]DocsVariable `yaml:"variables"`
}

type DocsVariable struct {
	Title       string                   `yaml:"title"`
	Description string                   `yaml:"description"`
	Default     string                   `yaml:"default"`
	Type        DocsVariableType         `yaml:"type"`
	Options     []DocsVariableOptionType `yaml:"options"`
}

type DocsVariableOptionType struct {
	Title string `yaml:"title"`
	Value string `yaml:"value"`
}

type DocsVariableType string

const (
	DocsVariableTypeText        DocsVariableType = "text"
	DocsVariableTypeInteger     DocsVariableType = "integer"
	DocsVariableTypeFloat       DocsVariableType = "float"
	DocsVariableTypeOptions     DocsVariableType = "options"
	DocsVariableTypeVolume      DocsVariableType = "volume"
	DocsVariableTypeApplication DocsVariableType = "application"
	DocsVariableTypeServer      DocsVariableType = "server"
)

// DockerProxyConfig is the configuration for the docker proxy
type DockerProxyConfig struct {
	Enabled    bool                  `yaml:"enabled"`
	Permission DockerProxyPermission `yaml:"permissions"`
}

// DockerProxyPermissionType is the type of permission to be granted to the proxy
type DockerProxyPermissionType string

const (
	// DockerProxyNoPermission no request will be allowed
	DockerProxyNoPermission DockerProxyPermissionType = "none"
	// DockerProxyReadPermission only [GET, HEAD] requests will be allowed
	DockerProxyReadPermission DockerProxyPermissionType = "read"
	// DockerProxyReadWritePermission all requests will be allowed [GET, HEAD, POST, PUT, DELETE, OPTIONS]
	DockerProxyReadWritePermission DockerProxyPermissionType = "read_write"
)

type DockerProxyPermission struct {
	Ping         DockerProxyPermissionType `yaml:"ping"`
	Version      DockerProxyPermissionType `yaml:"version"`
	Info         DockerProxyPermissionType `yaml:"info"`
	Events       DockerProxyPermissionType `yaml:"events"`
	Auth         DockerProxyPermissionType `yaml:"auth"`
	Secrets      DockerProxyPermissionType `yaml:"secrets"`
	Build        DockerProxyPermissionType `yaml:"build"`
	Commit       DockerProxyPermissionType `yaml:"commit"`
	Configs      DockerProxyPermissionType `yaml:"configs"`
	Containers   DockerProxyPermissionType `yaml:"containers"`
	Distribution DockerProxyPermissionType `yaml:"distribution"`
	Exec         DockerProxyPermissionType `yaml:"exec"`
	Grpc         DockerProxyPermissionType `yaml:"grpc"`
	Images       DockerProxyPermissionType `yaml:"images"`
	Networks     DockerProxyPermissionType `yaml:"networks"`
	Nodes        DockerProxyPermissionType `yaml:"nodes"`
	Plugins      DockerProxyPermissionType `yaml:"plugins"`
	Services     DockerProxyPermissionType `yaml:"services"`
	Session      DockerProxyPermissionType `yaml:"session"`
	Swarm        DockerProxyPermissionType `yaml:"swarm"`
	System       DockerProxyPermissionType `yaml:"system"`
	Tasks        DockerProxyPermissionType `yaml:"tasks"`
	Volumes      DockerProxyPermissionType `yaml:"volumes"`
}

func (s *Stack) deepCopy() (*Stack, error) {
	yamlBytes, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}
	newStack := &Stack{}
	err = yaml.Unmarshal(yamlBytes, newStack)
	if err != nil {
		return nil, err
	}
	return newStack, nil
}

func (s *Stack) VolumeNames() []string {
	volumeNames := make([]string, 0)
	for _, service := range s.Services {
		for _, volume := range service.Volumes {
			volumeNames = append(volumeNames, volume.Name)
		}
	}
	return volumeNames
}

func (s *Stack) ServiceNames() []string {
	serviceNames := make([]string, 0)
	for serviceName := range s.Services {
		serviceNames = append(serviceNames, serviceName)
	}
	return serviceNames
}

func (s *Stack) PreferredServerHostnames() []string {
	preferredServers := set.New[string](0)
	for _, service := range s.Services {
		for _, preferredServerHostname := range service.PreferredServerHostnames {
			preferredServers.Insert(preferredServerHostname)
		}
	}
	return preferredServers.Slice()
}

func (s *Service) ValidateDockerProxyConfig() error {
	if s.DockerProxyConfig.Enabled && len(s.PreferredServerHostnames) != 1 {
		return errors.New("you must provide exactly one preferred server for getting access to docker socket proxy")
	}
	return nil
}
