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
	"os"
	"os/user"
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

func AppendPublicSSHKeyLocally(pubKey string) error {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	// add \n to the end of the public key
	pubKey = pubKey + "\n"

	// Construct the path to the .ssh directory
	sshDirPath := filepath.Join(currentUser.HomeDir, ".ssh")

	// Create the .ssh directory if it doesn't exist
	err = os.MkdirAll(sshDirPath, 0700)
	if err != nil {
		return fmt.Errorf("failed to create .ssh directory: %v", err)
	}

	// Construct the path to the authorized_keys file
	authorizedKeysPath := filepath.Join(sshDirPath, "authorized_keys")

	// Open the authorized_keys file for appending
	f, err := os.OpenFile(authorizedKeysPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys file: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("failed to close authorized_keys file: %v", err)
		}
	}(f)

	// Append the public key to the file
	_, err = fmt.Fprintln(f, pubKey)
	if err != nil {
		return fmt.Errorf("failed to append public key to authorized_keys: %v", err)
	}

	return nil
}
