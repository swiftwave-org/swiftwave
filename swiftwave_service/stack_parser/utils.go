package stack_parser

import (
	"context"
	"errors"
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseStackYaml(yamlStr string, currentSwiftwaveVersion string) (Stack, error) {
	stack := Stack{}
	err := yaml.Unmarshal([]byte(yamlStr), &stack)
	if err != nil {
		return Stack{}, err
	}
	// convert the version to integer
	stackMinSWVersion, err := versionToInt(stack.MinimumSwiftwaveVersion)
	if err != nil {
		return Stack{}, err
	}
	if stackMinSWVersion > 0 {
		SWVersion, err := versionToInt(currentSwiftwaveVersion)
		if err != nil {
			return Stack{}, err
		}
		if SWVersion >= stackMinSWVersion {
			return Stack{}, fmt.Errorf(`Required Swiftwave %s. Current Version %s. Please upgrade to latest.`, stack.MinimumSwiftwaveVersion, currentSwiftwaveVersion)
		}

	}
	// Pre-fill default values
	for serviceName, service := range stack.Services {
		if service.Deploy.Mode == DeploymentModeNone {
			service.Deploy.Mode = DeploymentModeReplicated
		}
		if service.Deploy.Mode == DeploymentModeReplicated {
			if service.Deploy.Replicas == 0 {
				service.Deploy.Replicas = 1
			}
		} else if service.Deploy.Mode == DeploymentModeGlobal {
			if service.Deploy.Replicas != 0 {
				service.Deploy.Replicas = 0
			}
		} else {
			return Stack{}, errors.New("invalid deploy mode")
		}
		service.DockerProxyConfig.Permission.Ping = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Ping)
		service.DockerProxyConfig.Permission.Version = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Version)
		service.DockerProxyConfig.Permission.Info = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Info)
		service.DockerProxyConfig.Permission.Events = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Events)
		service.DockerProxyConfig.Permission.Auth = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Auth)
		service.DockerProxyConfig.Permission.Secrets = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Secrets)
		service.DockerProxyConfig.Permission.Build = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Build)
		service.DockerProxyConfig.Permission.Commit = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Commit)
		service.DockerProxyConfig.Permission.Configs = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Configs)
		service.DockerProxyConfig.Permission.Containers = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Containers)
		service.DockerProxyConfig.Permission.Distribution = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Distribution)
		service.DockerProxyConfig.Permission.Exec = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Exec)
		service.DockerProxyConfig.Permission.Grpc = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Grpc)
		service.DockerProxyConfig.Permission.Images = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Images)
		service.DockerProxyConfig.Permission.Networks = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Networks)
		service.DockerProxyConfig.Permission.Nodes = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Nodes)
		service.DockerProxyConfig.Permission.Plugins = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Plugins)
		service.DockerProxyConfig.Permission.Services = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Services)
		service.DockerProxyConfig.Permission.Session = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Session)
		service.DockerProxyConfig.Permission.Swarm = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Swarm)
		service.DockerProxyConfig.Permission.System = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.System)
		service.DockerProxyConfig.Permission.Tasks = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Tasks)
		service.DockerProxyConfig.Permission.Volumes = fillDefaultDockerProxyPermissionIfNotPresent(service.DockerProxyConfig.Permission.Volumes)
		stack.Services[serviceName] = service
	}
	// Append Stack Name Variable {{STACK_NAME}} to the services
	for serviceName, service := range stack.Services {
		if !strings.Contains(serviceName, "{{STACK_NAME}}") {
			newServiceName := "{{STACK_NAME}}_" + serviceName
			stack.Services[newServiceName] = service
			delete(stack.Services, serviceName)
		}
	}
	// delete the variables with `markdown` type
	if stack.Docs != nil {
		for variableKey, variable := range stack.Docs.Variables {
			if variable.Type == DocsVariableTypeMarkdown {
				delete(stack.Docs.Variables, variableKey)
			}
		}
	}
	return stack, nil
}

func (s *Stack) FillAndVerifyVariables(variableMapping *map[string]string, serviceManager service_manager.ServiceManager) (*Stack, error) {
	if variableMapping == nil {
		return nil, errors.New("variableMapping is nil")
	}
	// check if STACK_NAME is present in variableMapping
	if _, ok := (*variableMapping)["STACK_NAME"]; !ok {
		return nil, errors.New("STACK_NAME is not provided")
	} else {
		if len(strings.TrimSpace((*variableMapping)["STACK_NAME"])) == 0 {
			return nil, errors.New("STACK_NAME is empty")
		}
	}

	stackCopy, err := s.deepCopy()
	if err != nil {
		return nil, errors.New("error in copying stack")
	}
	// fill default variable values in variableMapping
	if s.Docs != nil {
		for variableKey, variable := range s.Docs.Variables {
			if _, ok := (*variableMapping)[variableKey]; !ok {
				(*variableMapping)[variableKey] = variable.Default
			}
		}
	}

	// fill variables name in Service Name
	for serviceName, service := range stackCopy.Services {
		newServiceName := variableFillerHelper(serviceName, variableMapping)
		if newServiceName != serviceName {
			stackCopy.Services[newServiceName] = service
			delete(stackCopy.Services, serviceName)
		}
	}
	// iterate over all services
	for serviceName, service := range stackCopy.Services {
		service.Image = variableFillerHelper(service.Image, variableMapping)
		// [IGNORE] Deploy config shouldn't have any variables
		// iterate over volumes
		for i, volume := range service.Volumes {
			volume.Name = variableFillerHelper(volume.Name, variableMapping)
			volume.MountingPoint = variableFillerHelper(volume.MountingPoint, variableMapping)
			service.Volumes[i] = volume
		}
		// iterate over environment variables
		for key, value := range service.Environment {
			newKey := variableFillerHelper(key, variableMapping)
			newValue := variableFillerHelper(value, variableMapping)
			delete(service.Environment, key)
			service.Environment[newKey] = newValue
		}
		// iterate over configs
		for i, config := range service.Configs {
			config.Content = variableFillerHelper(config.Content, variableMapping)
			config.MountingPath = variableFillerHelper(config.MountingPath, variableMapping)
			service.Configs[i] = config
		}
		// [IGNORE] CapAdd shouldn't have any variables
		// [IGNORE] Sysctls shouldn't have any variables
		// iterate over command
		for i, command := range service.Command {
			newCommand := variableFillerHelper(command, variableMapping)
			service.Command[i] = newCommand
		}
		// iterate over preferred deployment server
		if service.PreferredServerHostnames != nil {
			servers := make([]string, 0)
			for _, server := range service.PreferredServerHostnames {
				servers = append(servers, variableFillerHelper(server, variableMapping))
			}
			service.PreferredServerHostnames = servers
		}
		// inject variable in healthcheck if required
		newHealthCheckTestCommand := variableFillerHelper(service.CustomHealthCheck.TestCommand, variableMapping)
		service.CustomHealthCheck.TestCommand = newHealthCheckTestCommand
		stackCopy.Services[serviceName] = service
	}
	// check if docs present
	if stackCopy.Docs != nil {
		// fetch a swarm manager server
		server, err := core.FetchSwarmManager(&serviceManager.DbClient)
		if err != nil {
			return nil, errors.New("error in fetching swarm manager")
		}
		// fetch docker manager
		dockerManager, err := manager.DockerClient(context.Background(), server)
		if err != nil {
			return nil, errors.New("error in connecting to docker manager")
		}
		// verify the type of variables
		for variableKey, variable := range stackCopy.Docs.Variables {
			// check if variableKey is present in variableMapping
			if _, ok := (*variableMapping)[variableKey]; ok {
				if variable.Type == DocsVariableTypeInteger {
					_, err := stringToInteger((*variableMapping)[variableKey])
					if err != nil {
						return nil, errors.New("variable " + variableKey + " should be integer")
					}
				} else if variable.Type == DocsVariableTypeFloat {
					_, err := strconv.ParseFloat((*variableMapping)[variableKey], 64)
					if err != nil {
						return nil, errors.New("variable " + variableKey + " should be float")
					}
				} else if variable.Type == DocsVariableTypeOptions {
					isValid := false
					for _, option := range variable.Options {
						if option.Value == (*variableMapping)[variableKey] {
							isValid = true
							break
						}
					}
					if !isValid {
						return nil, errors.New("variable " + variableKey + " should be one of the provided options")
					}
				} else if variable.Type == DocsVariableTypeVolume {
					val := (*variableMapping)[variableKey]
					isExist, err := core.IsExistPersistentVolume(context.Background(), serviceManager.DbClient, val, *dockerManager)
					if err != nil {
						return nil, errors.New("error in checking volume " + val)
					}
					if !isExist {
						return nil, errors.New("volume " + val + " doesn't exist. Create it or choose another volume")
					}
				} else if variable.Type == DocsVariableTypeText {
					// do nothing, just for the sake of completeness
				} else if variable.Type == DocsVariableTypeApplication {
					val := (*variableMapping)[variableKey]
					isExist, err := core.IsExistApplicationName(context.Background(), serviceManager.DbClient, *dockerManager, val)
					if err != nil {
						return nil, errors.New("error in checking application " + val)
					}
					if !isExist {
						return nil, errors.New("application " + val + " doesn't exist. Create it or choose another application")
					}
				} else if variable.Type == DocsVariableTypeServer {
					val := (*variableMapping)[variableKey]
					_, err := core.FetchServerIDByHostName(&serviceManager.DbClient, val)
					if err != nil {
						return nil, errors.New("invalid server " + val + " provided")
					}
				} else {
					return nil, errors.New("invalid variable type")
				}
			}
		}
	}
	return stackCopy, nil
}

func (s *Stack) String(ignoreDocs bool) (string, error) {
	// copy the stack
	stackCopy, err := s.deepCopy()
	if err != nil {
		return "", errors.New("error in copying stack")
	}
	// set nil docs
	if ignoreDocs {
		stackCopy.Docs = nil
	}
	// marshal the stack to yaml
	yamlBytes, err := yaml.Marshal(stackCopy)
	if err != nil {
		return "", err
	}
	output := string(yamlBytes)
	if ignoreDocs {
		// remove `docs: null` from the yaml
		output = strings.Replace(output, "docs: null", "", -1)
	}
	output = strings.TrimSpace(output)
	return output, nil
}

func variableFillerHelper(text string, variableMapping *map[string]string) string {
	// fill random number
	for {
		output, isUpdated := randomNumberFiller(text)
		text = output
		if !isUpdated {
			break
		}
	}
	// fill random character
	for {
		output, isUpdated := randomCharacterFiller(text)
		text = output
		if !isUpdated {
			break
		}
	}
	//  fill variables
	for {
		output, isUpdated := variableFiller(text, variableMapping)
		text = output
		if !isUpdated {
			break
		}
	}
	return text
}

func variableFiller(text string, variableMapping *map[string]string) (string, bool) {
	regexStr := `{{[^{}]*}}`
	regex := regexp.MustCompile(regexStr)
	isUpdated := false
	if regex.MatchString(text) {
		indexRange := regex.FindStringIndex(text)
		variableName := text[indexRange[0]+2 : indexRange[1]-2]
		if val, ok := (*variableMapping)[variableName]; ok {
			text = sliceOfString(text, 0, indexRange[0]) + val + sliceOfString(text, indexRange[1], len(text))
			isUpdated = true
		} else {
			// check if started with RANDOM_
			if len(variableName) > 7 && strings.Compare(variableName[:7], "RANDOM_") == 0 {
				randomValue := generateRandomString(16)
				(*variableMapping)[variableName] = randomValue
				text = sliceOfString(text, 0, indexRange[0]) + randomValue + sliceOfString(text, indexRange[1], len(text))
				isUpdated = true
			}
		}
	}
	return text, isUpdated
}

func randomNumberFiller(text string) (string, bool) {
	isUpdated := false
	// check if {{$+}} is present
	randomCharacterRegexStr := `{{\$+}}`
	randomCharacterRegex := regexp.MustCompile(randomCharacterRegexStr)
	if randomCharacterRegex.MatchString(text) {
		indexRange := randomCharacterRegex.FindStringIndex(text)
		length := indexRange[1] - indexRange[0] - 4
		if length > 0 {
			randomString := generateRandomNumber(length)
			text = sliceOfString(text, 0, indexRange[0]) + randomString + sliceOfString(text, indexRange[1], len(text))
			isUpdated = true
		}
	}
	return text, isUpdated
}

func randomCharacterFiller(text string) (string, bool) {
	isUpdated := false
	// check if {{#+}} is present
	randomCharacterRegexStr := `{{#+}}`
	randomCharacterRegex := regexp.MustCompile(randomCharacterRegexStr)
	if randomCharacterRegex.MatchString(text) {
		indexRange := randomCharacterRegex.FindStringIndex(text)
		length := indexRange[1] - indexRange[0] - 4
		if length > 0 {
			randomString := generateRandomString(length)
			text = sliceOfString(text, 0, indexRange[0]) + randomString + sliceOfString(text, indexRange[1], len(text))
			isUpdated = true
		}
	}
	return text, isUpdated
}

func sliceOfString(text string, start int, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	if start >= end {
		return ""
	}
	return text[start:end]
}

func generateRandomString(length int) string {
	if length <= 0 {
		return ""
	}
	const charset = "abcdefghijklmnopqrstuvwxyz"
	randomBytes := make([]byte, length)
	for i := range randomBytes {
		randomBytes[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomBytes)
}

func generateRandomNumber(length int) string {
	if length <= 0 {
		return ""
	}
	const charset = "0123456789"
	randomBytes := make([]byte, length)
	for i := range randomBytes {
		randomBytes[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomBytes)
}

func stringToInteger(str string) (int, error) {
	return strconv.Atoi(str)
}

func fillDefaultDockerProxyPermissionIfNotPresent(val DockerProxyPermissionType) DockerProxyPermissionType {
	if val == DockerProxyNoPermission || val == DockerProxyReadPermission || val == DockerProxyReadWritePermission {
		return val
	}
	return DockerProxyNoPermission
}

func versionToInt(version string) (int, error) {
	if len(version) == 0 {
		// fallback to the lowest version for backward compatibility
		return 0, nil
	}
	if strings.Compare(version, "develop") == 0 {
		return math.MaxInt32, nil
	}
	// Remove the 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Split the version string into parts
	parts := strings.Split(version, ".")

	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch := strings.Split(parts[2], "-")[0] // Remove any pre-release suffix
	patchNum, err := strconv.Atoi(patch)
	if err != nil {
		return 0, fmt.Errorf("invalid patch version: %s", patch)
	}

	// Combine the parts into a single integer
	return major*100 + minor*10 + patchNum, nil
}
