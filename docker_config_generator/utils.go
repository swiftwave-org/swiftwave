package dockerconfiggenerator

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	GIT "github.com/swiftwave-org/swiftwave/git_manager"

	"github.com/google/uuid"
)

// Generate DockerConfig from git repository.
func (m Manager) GenerateConfigFromGitRepository(git_url string, branch string, username string, password string) (DockerFileConfig, error) {
	tmpFolder := "/tmp/" + uuid.New().String()
	if os.Mkdir(tmpFolder, 0777) != nil {
		return DockerFileConfig{}, errors.New("failed to create tmp folder")
	}
	defer deleteDirectory(tmpFolder)
	// Clone repository
	err := GIT.CloneRepository(git_url, branch, username, password, tmpFolder)
	if err != nil {
		return DockerFileConfig{}, errors.New("failed to clone repository")
	}
	// Generate config from source code directory
	return m.generateConfigFromSourceCodeDirectory(tmpFolder)
}

// Generate DockerConfig from source code .tar file.
func (m Manager) GenerateConfigFromSourceCodeTar(tarFile string) (DockerFileConfig, error) {
	// Extract tar file to a temporary folder
	tmpFolder := "/tmp/" + uuid.New().String()
	defer deleteDirectory(tmpFolder)
	err := ExtractTar(tarFile, tmpFolder)
	if err != nil {
		log.Println(err)
		return DockerFileConfig{}, errors.New("failed to extract tar file")
	}
	// Generate config from source code directory
	return m.generateConfigFromSourceCodeDirectory(tmpFolder)
}

// Generate DockerConfig from source code directory.
func (m Manager) generateConfigFromSourceCodeDirectory(directory string) (DockerFileConfig, error) {
	// Try to find docker file
	file, err := os.ReadFile(directory + "/Dockerfile")
	if err != nil {
		file, err = os.ReadFile(directory + "/dockerfile")
		if err != nil {
			file, err = os.ReadFile(directory + "/DockerFile")
		}
	}

	if err == nil {
		// Dockerfile found
		dockerConfig := DockerFileConfig{}
		dockerConfig.DetectedService = "Dockerfile from source code"
		dockerConfig.DockerFile = string(file)
		dockerConfig.Variables = ParseBuildArgsFromDockerfile(string(file))
		if dockerConfig.Variables == nil {
			dockerConfig.Variables = map[string]Variable{}
		}
		return dockerConfig, nil
	}

	// In case Dockerfile is not found, try to detect service
	// Look for other files and generate docker file
	var lookupFiles map[string]string = map[string]string{}
	for _, lookupFile := range m.Config.LookupFiles {
		if existsInFolder(directory, lookupFile) {
			f, err := os.Open(directory + "/" + lookupFile)
			if err != nil {
				return DockerFileConfig{}, errors.New("failed to open file " + lookupFile + "")
			}
			file, err := io.ReadAll(f)
			if err != nil {
				return DockerFileConfig{}, errors.New("failed to fetch file content for " + lookupFile + "")
			}
			lookupFiles[lookupFile] = string(file)
		} else {
			lookupFiles[lookupFile] = ""
		}
	}

	// detect service
	for _, serviceName := range m.Config.ServiceOrder {
		// Fetch service selectors
		identifiers := m.Config.Identifiers[serviceName]
		for _, identifier := range identifiers {
			// Fetch file content for each selector
			isIdentifierMatched := false
			// check keywords for each selector
			for _, selector := range identifier.Selectors {
				// check if file exists
				if lookupFiles[selector.File] == "" {
					break
				}
				isMatched := true
				// Check if file content contains keywords
				for _, keyword := range selector.Keywords {
					isMatched = isMatched && strings.Contains(lookupFiles[selector.File], keyword)
				}
				isIdentifierMatched = isIdentifierMatched || isMatched
			}
			// if identifiers is not matched, continue to check extension files if specified
			isFileExtensionMatched := false
			if !isIdentifierMatched {
				for _, extension := range identifier.Extensions {
					if hasFileWithExtension(directory, extension) {
						isFileExtensionMatched = true
						break
					}
				}
			}
			if isIdentifierMatched || isFileExtensionMatched {
				// Fetch docker file
				dockerConfig := DockerFileConfig{}
				dockerConfig.DetectedService = serviceName
				dockerConfig.DockerFile = m.DockerTemplates[serviceName]
				dockerConfig.Variables = m.Config.Templates[serviceName].Variables
				if dockerConfig.Variables == nil {
					dockerConfig.Variables = map[string]Variable{}
				}
				return dockerConfig, nil
			}
		}
	}

	return DockerFileConfig{}, errors.New("failed to detect service")
}

// Generate DockerConfig from custom dockerfile. If GenerateConfigFromGitRepository fails to detect service, this function will be used.
func (m Manager) GenerateConfigFromCustomDocker(dockerfile string) DockerFileConfig {
	dockerConfig := DockerFileConfig{}
	dockerConfig.DetectedService = "Custom Dockerfile"
	dockerConfig.DockerFile = dockerfile
	dockerConfig.Variables = ParseBuildArgsFromDockerfile(dockerfile)
	if dockerConfig.Variables == nil {
		dockerConfig.Variables = map[string]Variable{}
	}
	return dockerConfig
}

// DefaultArgs returns default arguments for a service.
func (m Manager) DefaultArgsFromService(serviceName string) map[string]string {
	args := map[string]string{}
	if _, ok := m.Config.Templates[serviceName]; !ok {
		return args
	}
	for key, variable := range m.Config.Templates[serviceName].Variables {
		args[key] = variable.Default
	}
	return args
}
