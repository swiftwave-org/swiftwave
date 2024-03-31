package containermanger

import (
	"github.com/docker/docker/api/types/registry"
	"strings"
)

func generateAuthHeader(username string, password string) (string, error) {
	if strings.Compare(username, "") == 0 && strings.Compare(password, "") == 0 {
		return "", nil
	}
	authConfig := registry.AuthConfig{
		Username: username,
		Password: password,
	}
	return registry.EncodeAuthConfig(authConfig)
}
