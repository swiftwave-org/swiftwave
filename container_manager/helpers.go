package containermanger

import (
	"github.com/docker/docker/api/types/registry"
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
