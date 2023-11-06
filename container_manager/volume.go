package containermanger

import (
	"errors"
	"github.com/docker/docker/api/types/volume"
)

// CreateVolume : Create a new volume, return id of the volume
func (m Manager) CreateVolume(name string) error {
	_, err := m.client.VolumeCreate(m.ctx, volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return errors.New("error creating volume " + err.Error())
	}
	return nil
}

// RemoveVolume : Remove a volume by id
func (m Manager) RemoveVolume(id string) error {
	err := m.client.VolumeRemove(m.ctx, id, true)
	if err != nil {
		return errors.New("error removing volume " + err.Error())
	}
	return nil
}

// ExistsVolume : Check if volume exists
func (m Manager) ExistsVolume(id string) bool {
	_, err := m.client.VolumeInspect(m.ctx, id)
	return err == nil
}

// FetchVolumes Fetch all volumes
func (m Manager) FetchVolumes() ([]string, error) {
	volumes, err := m.client.VolumeList(m.ctx, volume.ListOptions{})
	if err != nil {
		return nil, errors.New("error fetching volumes " + err.Error())
	}
	var volumeNames []string = make([]string, len(volumes.Volumes))
	for i, v := range volumes.Volumes {
		volumeNames[i] = v.Name
	}
	return volumeNames, nil
}
