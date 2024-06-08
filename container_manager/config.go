package containermanger

import (
	"encoding/base64"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/oklog/ulid"
	"math/rand"
	"time"
)

// configId is the actual id of the config in swiftwave system
// it's different from the docker config id, it's the same as docker config name

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

// FetchDockerConfigId fetches the docker config id of a config
func (m Manager) FetchDockerConfigId(configId string) (string, error) {
	config, _, err := m.client.ConfigInspectWithRaw(m.ctx, configId)
	if err != nil {
		return "", err
	}
	return config.ID, nil
}

// CreateConfig creates a new config and returns the config id
func (m Manager) CreateConfig(content string, applicationId string) (string, error) {
	// generate a random config id
	configId := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
	_, err := m.client.ConfigCreate(m.ctx, swarm.ConfigSpec{
		Annotations: swarm.Annotations{
			Name: configId,
			Labels: map[string]string{
				"applicationId": applicationId,
			},
		},
		Data: []byte(content),
	})
	if err != nil {
		return "", err
	}
	return configId, nil
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
