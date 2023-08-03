package dockerconfiggenerator

import (
	"errors"
	GIT "keroku/m/git_manager"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

// Generate DockerConfig from git repository.
func (m Manager) GenerateConfigFromGitRepository(manager GIT.Manager, repo GIT.Repository) (DockerFileConfig, error) {
	folderStructure, err := manager.FetchFolderStructure(repo)
	if err != nil {
		return DockerFileConfig{}, errors.New("failed to fetch folder structure")
	}
	// Try to find docker file
	file, err := manager.FetchFileContent(repo, "Dockerfile")
	if err != nil {
		file, err = manager.FetchFileContent(repo, "dockerfile")
		if err != nil {
			file, err = manager.FetchFileContent(repo, "DockerFile")
		}
	}
	if err == nil {
		// Dockerfile found
		dockerConfig := DockerFileConfig{}
		dockerConfig.DetectedService = "Dockerfile from repository"
		dockerConfig.DockerFile = file
		dockerConfig.Variables = ParseBuildArgsFromDockerfile(file)
		return dockerConfig, nil
	}

	// In case Dockerfile is not found, try to detect service
	// Look for other files and generate docker file
	var lookupFiles map[string]string = map[string]string{}
	for _, lookupFile := range m.Config.LookupFiles {
		if existsInArray(folderStructure, lookupFile) {
			file, err := manager.FetchFileContent(repo, lookupFile)
			if err != nil {
				return DockerFileConfig{}, errors.New("failed to fetch file content for " + lookupFile + "")
			}
			lookupFiles[lookupFile] = file
		} else {
			lookupFiles[lookupFile] = ""
		}
	}

	for _, serviceName := range m.Config.ServiceOrder {
		// Fetch service selectors
		identifiers := m.Config.Identifiers[serviceName]
		for _, identifier := range identifiers {
			// Fetch file content for each selector
			isIdentifierMatched := false
			for _, selector := range identifier.Selector {
				isMatched := true
				// Check if file content contains keywords
				for _, keyword := range selector.Keywords {
					isMatched = isMatched && strings.Contains(lookupFiles[selector.File], keyword)
				}
				isIdentifierMatched = isIdentifierMatched || isMatched
			}
			if isIdentifierMatched {
				// Fetch docker file
				dockerConfig := DockerFileConfig{}
				dockerConfig.DetectedService = serviceName
				dockerConfig.DockerFile = m.DockerTemplates[serviceName]
				dockerConfig.Variables = m.Config.Templates[serviceName].Variables
				return dockerConfig, nil
			}
		}
	}

	return DockerFileConfig{}, errors.New("failed to detect service")
}

// Generate DockerConfig from source code .tar file.
func (m Manager) GenerateConfigFromSourceCodeTar(tarFile string) (DockerFileConfig, error) {
	// Extract tar file to a temporary folder
	tmpFolder := "/tmp/" + uuid.New().String()
	defer deleteDirectory(tmpFolder)
	err := ExtractTar(tarFile, tmpFolder)
	if err != nil {
		log.Println(err)
		deleteDirectory(tmpFolder)
		return DockerFileConfig{}, errors.New("failed to extract tar file")
	}
	// Try to find docker file
	file, err := os.ReadFile(tmpFolder + "/Dockerfile")
	if err != nil {
		file, err = os.ReadFile(tmpFolder + "/dockerfile")
		if err != nil {
			file, err = os.ReadFile(tmpFolder + "/DockerFile")
		}
	}

	if err == nil {
		// Dockerfile found
		dockerConfig := DockerFileConfig{}
		dockerConfig.DetectedService = "Dockerfile from source code"
		dockerConfig.DockerFile = string(file)
		dockerConfig.Variables = ParseBuildArgsFromDockerfile(string(file))
		return dockerConfig, nil
	}

	// In case Dockerfile is not found, try to detect service
	// Look for other files and generate docker file
	var lookupFiles map[string]string = map[string]string{}
	for _, lookupFile := range m.Config.LookupFiles {
		if existsInFolder(tmpFolder, lookupFile) {
			file, err := os.ReadFile(tmpFolder + "/" + lookupFile)
			if err != nil {
				return DockerFileConfig{}, errors.New("failed to fetch file content for " + lookupFile + "")
			}
			lookupFiles[lookupFile] = string(file)
		} else {
			lookupFiles[lookupFile] = ""
		}
	}

	for _, serviceName := range m.Config.ServiceOrder {
		// Fetch service selectors
		identifiers := m.Config.Identifiers[serviceName]
		for _, identifier := range identifiers {
			// Fetch file content for each selector
			isIdentifierMatched := false
			for _, selector := range identifier.Selector {
				isMatched := true
				// Check if file content contains keywords
				for _, keyword := range selector.Keywords {
					isMatched = isMatched && strings.Contains(lookupFiles[selector.File], keyword)
				}
				isIdentifierMatched = isIdentifierMatched || isMatched
			}
			if isIdentifierMatched {
				// Fetch docker file
				dockerConfig := DockerFileConfig{}
				dockerConfig.DetectedService = serviceName
				dockerConfig.DockerFile = m.DockerTemplates[serviceName]
				dockerConfig.Variables = m.Config.Templates[serviceName].Variables
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
