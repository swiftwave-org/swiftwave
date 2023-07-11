package containermanger

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/volume"
)

// Create a new volume, return id of the volume
func (m Manager) CreateVolume(name string) (string, error) {
	createdVolume, err := m.client.VolumeCreate(m.ctx, volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return "", errors.New("error creating volume " + err.Error())
	}
	return createdVolume.Name, nil
}

// Remove a volume by id
func (m Manager) RemoveVolume(id string) error {
	err := m.client.VolumeRemove(m.ctx, id, true)
	if err != nil {
		return errors.New("error removing volume " + err.Error())
	}
	return nil
}

// Check if volume exists
func (m Manager) ExistVolume(id string) bool {
	d, err := m.client.VolumeInspect(m.ctx, id)
	fmt.Println(d)
	return err == nil
}