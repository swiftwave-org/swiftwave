package containermanger

import (
	"encoding/base64"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
)

// FetchConfig fetches the config with the given id
func (m Manager) FetchConfig(configId string) (string, error) {
	config, _, err := m.client.ConfigInspectWithRaw(m.ctx, configId)
	if err != nil {
		return "", err
	}
	decodeBytes, err := base64.StdEncoding.DecodeString(string(config.Spec.Data))
	if err != nil {
		return "", err
	}
	return string(decodeBytes), nil
}

// CreateConfig creates a new config and returns the config id
func (m Manager) CreateConfig(content string, applicationId string) (configId string, err error) {
	// encode base64 content
	b64encodedContent := []byte(base64.StdEncoding.EncodeToString([]byte(content)))
	response, err := m.client.ConfigCreate(m.ctx, swarm.ConfigSpec{
		Annotations: swarm.Annotations{
			Labels: map[string]string{
				"applicationId": applicationId,
			},
		},
		Data: b64encodedContent,
	})
	if err != nil {
		return "", err
	}
	configId = response.ID
	return
}

// RemoveConfig removes all the configs with the given applicationId
func (m Manager) RemoveConfig(applicationId string) (bool, error) {
	res, err := m.client.ConfigList(m.ctx, types.ConfigListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "applicationId="+applicationId),
		),
	})
	if err != nil {
		return false, err
	}
	err = nil
	for _, c := range res {
		err2 := m.client.ConfigRemove(m.ctx, c.ID)
		if err2 != nil {
			err = err2
		}
	}
	return err == nil, err
}

// PruneConfig removes all the configs with the given applicationId.
// It will remove all the possible configs
// It will not raise any error if failed to remove a config
func (m Manager) PruneConfig(applicationId string) {
	res, err := m.client.ConfigList(m.ctx, types.ConfigListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "applicationId="+applicationId),
		),
	})
	if err != nil {
		return
	}
	for _, c := range res {
		_ = m.client.ConfigRemove(m.ctx, c.ID)
	}
}
