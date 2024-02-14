package containermanger

import (
	"github.com/docker/docker/api/types/registry"
	"io"
	"log"
	"os"
)

func generateAuthHeader(username string, password string) (string, error) {
	if username == "" && password == "" {
		return "", nil
	}
	authConfig := registry.AuthConfig{
		Username: username,
		Password: password,
	}
	return registry.EncodeAuthConfig(authConfig)
}

// copyFile : Copy a file from source to destination
func copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func(sourceFile *os.File) {
		err := sourceFile.Close()
		if err != nil {
			log.Println("failed to close source file " + err.Error())
		}
	}(sourceFile)

	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer func(destinationFile *os.File) {
		err := destinationFile.Close()
		if err != nil {
			log.Println("failed to close destination file " + err.Error())
		}
	}(destinationFile)

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
