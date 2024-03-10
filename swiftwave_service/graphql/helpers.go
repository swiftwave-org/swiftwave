package graphql

import (
	"context"
	"errors"
	"fmt"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	dockerconfiggenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
	"log"
	"path/filepath"
	"strings"
)

func generateGitUrl(provider model.GitProvider, owner string, repo string) string {
	if provider == model.GitProviderGithub {
		return fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	} else if provider == model.GitProviderGitlab {
		return fmt.Sprintf("https://gitlab.com/%s/%s", owner, repo)
	} else if provider == model.GitProviderNone {
		return ""
	} else {
		return ""
	}
}

func convertMapToDockerConfigBuildArgs(input map[string]dockerconfiggenerator.Variable) []*model.DockerConfigBuildArg {
	var output = make([]*model.DockerConfigBuildArg, 0)
	for key, value := range input {
		output = append(output, &model.DockerConfigBuildArg{
			Key:          key,
			Type:         value.Type,
			Description:  value.Description,
			DefaultValue: value.Default,
		})
	}
	return output
}

/*
SanitizeFileName Sanitize the fileName to remove potentially dangerous characters
It's meant to be used for filename
Should not use this to sanitize file path
*/
func sanitizeFileName(fileName string) string {
	// Remove any path components and keep only the file name
	fileName = filepath.Base(fileName)

	// Remove potentially dangerous characters like ".."
	fileName = strings.ReplaceAll(fileName, "..", "")

	// Remove potentially dangerous characters like "/"
	fileName = strings.ReplaceAll(fileName, "/", "")

	// You can add more sanitization rules as needed

	return fileName
}

func FetchDockerManager(ctx context.Context, db *gorm.DB) (*containermanger.Manager, error) {
	// Fetch a random swarm manager
	swarmManagerServer, err := core.FetchSwarmManager(db)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to fetch swarm manager")
	}
	// Fetch docker manager
	dockerManager, err := manager.DockerClient(ctx, swarmManagerServer)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to fetch docker manager")
	}
	return dockerManager, nil
}
