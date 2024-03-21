package containermanger

import (
	"bufio"
	"context"
	"errors"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
)

/*
CreateImageWithContext builds a Docker image from a Dockerfile and returns a scanner to read the build logs.
It takes the Dockerfile content as a string, a map of build arguments, the path to the code directory, and the name of the image to be built.
It returns a scanner to read the build logs and an error if any.
It takes a context.Context as an additional argument.
*/
func (m Manager) CreateImageWithContext(ctx context.Context, dockerfile string, buildargs map[string]string, sourceCodeDirectory string, codePath string, imagename string) (*bufio.Scanner, error) {
	// add path
	codePath = strings.TrimSpace(codePath)
	if codePath != "" && codePath != "/" {
		sourceCodeDirectory = sourceCodeDirectory + "/" + codePath
		sourceCodeDirectory = strings.ReplaceAll(sourceCodeDirectory, "\\", "/")
		sourceCodeDirectory = strings.ReplaceAll(sourceCodeDirectory, "//", "/")
		sourceCodeDirectory = strings.ReplaceAll(sourceCodeDirectory, "../", "")
		sourceCodeDirectory = strings.ReplaceAll(sourceCodeDirectory, "./", "")
	}
	// Move the dockerfile to the sourceCodeDirectory
	err := os.WriteFile(sourceCodeDirectory+"/Dockerfile", []byte(dockerfile), 0777)
	if err != nil {
		return nil, errors.New("failed to write the dockerfile")
	}
	// Buildargs map
	final_buildargs := map[string]*string{}
	// convert buildargs map to final_buildargs map
	for key, value := range buildargs {
		valueBytes := []byte(value)
		ptrValue := new(string)
		*ptrValue = string(valueBytes)
		final_buildargs[key] = ptrValue
	}
	// tar the sourceCodeDirectory
	tar, err := archive.TarWithOptions(sourceCodeDirectory, &archive.TarOptions{})
	if err != nil {
		return nil, errors.New("failed to tar the sourceCodeDirectory")
	}
	// Build the image
	response, err := m.client.ImageBuild(ctx, tar, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Remove:     true,
		NoCache:    true,
		Tags:       []string{imagename},
		BuildArgs:  final_buildargs,
	})
	if err != nil {
		return nil, errors.New("failed to build the image")
	}
	// Return scanner to read the build logs
	scanner := bufio.NewScanner(response.Body)
	return scanner, nil
}

// PushImage pushes a Docker image to a remote registry and returns a scanner to read the push logs.
func (m Manager) PushImage(ctx context.Context, imageTag string, username string, password string) (*bufio.Scanner, error) {
	authHeader, err := generateAuthHeader(username, password)
	if err != nil {
		return nil, errors.New("failed to generate auth header")
	}
	// Push the image
	scanner, err := m.client.ImagePush(ctx, imageTag, image.PushOptions{
		RegistryAuth: authHeader,
	})
	if err != nil {
		return nil, errors.New("failed to push the image")
	}
	return bufio.NewScanner(scanner), nil
}

// PullImage pulls a Docker image from a remote registry and returns a scanner to read the pull logs.
func (m Manager) PullImage(imageTag string, username string, password string) (*bufio.Scanner, error) {
	authHeader, err := generateAuthHeader(username, password)
	if err != nil {
		return nil, errors.New("failed to generate auth header")
	}
	// Pull the image
	scanner, err := m.client.ImagePull(m.ctx, imageTag, image.PullOptions{
		RegistryAuth: authHeader,
	})
	if err != nil {
		return nil, errors.New("failed to pull the image")
	}
	return bufio.NewScanner(scanner), nil
}

// ExistsImage checks if a Docker image exists locally.
func (m Manager) ExistsImage(image string) bool {
	// Check if the image exists locally
	_, _, err := m.client.ImageInspectWithRaw(m.ctx, image)
	return err == nil
}

// RemoveImage removes a Docker image from the local registry.
func (m Manager) RemoveImage(imageTag string) error {
	// Remove the image
	_, err := m.client.ImageRemove(m.ctx, imageTag, image.RemoveOptions{})
	if err != nil {
		return errors.New("failed to remove the image")
	}
	return nil
}

// PruneImages removes all unused Docker images from the local registry.
func (m Manager) PruneImages() error {
	// Prune the images including dangling images
	pruneFilters := filters.NewArgs()
	pruneFilters.Add("dangling", strconv.FormatBool(false))
	_, err := m.client.ImagesPrune(m.ctx, pruneFilters)
	if err != nil {
		return errors.New("failed to prune the images")
	}
	return nil
}
