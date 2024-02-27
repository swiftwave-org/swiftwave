package stack_parser

import (
	"errors"
	"math/rand"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseStackYaml(yamlStr string) (Stack, error) {
	stack := Stack{}
	err := yaml.Unmarshal([]byte(yamlStr), &stack)
	if err != nil {
		return Stack{}, err
	}
	// Pre-fill default values
	for serviceName, service := range stack.Services {
		if service.Deploy.Mode == "" {
			service.Deploy.Mode = "replicated"
		} else if service.Deploy.Mode == "replicated" && service.Deploy.Replicas == 0 {
			service.Deploy.Replicas = 1
		} else if service.Deploy.Mode == "global" && service.Deploy.Replicas != 0 {
			service.Deploy.Replicas = 0
		} else {
			return Stack{}, errors.New("invalid deploy mode")
		}
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
	return stack, nil
}

func (s *Stack) FillVariable(variableMapping *map[string]string) (*Stack, error) {
	if variableMapping == nil {
		return nil, errors.New("variableMapping is nil")
	}
	// check if STACK_NAME is present in variableMapping
	if _, ok := (*variableMapping)["STACK_NAME"]; !ok {
		return nil, errors.New("STACK_NAME is not provided")
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
		// [IGNORE] CapAdd shouldn't have any variables
		// [IGNORE] Sysctls shouldn't have any variables
		// iterate over command
		for i, command := range service.Command {
			newCommand := variableFillerHelper(command, variableMapping)
			service.Command[i] = newCommand
		}
		stackCopy.Services[serviceName] = service
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
