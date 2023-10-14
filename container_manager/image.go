package containermanger

import (
	"bufio"
	"errors"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
)

/*
`CreateImage` builds a Docker image from a Dockerfile and returns a scanner to read the build logs.
It takes the Dockerfile content as a string, a map of build arguments, the path to the code directory, and the name of the image to be built.
It returns a scanner to read the build logs and an error if any.
*/
func (m Manager) CreateImage(dockerfile string, buildargs map[string]string, codepath string, imagename string) (*bufio.Scanner, error) {
	// Move the dockerfile to the codepath
	err := os.WriteFile(codepath+"/Dockerfile", []byte(dockerfile), 0777)
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
	// tar the codepath
	tar, err := archive.TarWithOptions(codepath, &archive.TarOptions{})
	if err != nil {
		return nil, errors.New("failed to tar the codepath")
	}
	// Build the image
	response, err := m.client.ImageBuild(m.ctx, tar, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Remove:     true,
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
