package dockermanager

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/volume"
)

// Create a new volume, return id of the volume
func (m Manager) CreateVolume() (string, error) {
	name, err := generateLongRandomString(128)
	if err != nil {
		return "", errors.New("error generating uuid")
	}
	createdVolume, err := m.client.VolumeCreate(m.ctx, volume.CreateOptions{
		Name: name,
	});
	if err != nil {
		return "", errors.New("error creating volume "+ err.Error())
	}
	return createdVolume.Name, nil
}

// Remove a volume by id
func (m Manager) RemoveVolume(id string) error {
	err := m.client.VolumeRemove(m.ctx, id, true)
	if err != nil {
		return errors.New("error removing volume "+ err.Error())
	}
	return nil
}

// Check if volume exists
func (m Manager) VolumeExists(id string) bool {
	d, err := m.client.VolumeInspect(m.ctx, id)
	fmt.Println(d)
	return err == nil
}

// Get volume storage usage
// - return size in bytes
func (m Manager) VolumeUsage(id string) (int64, error) {
	// TODO: update later
	return 1000 , nil
	// foundVolume, err := m.client.VolumeInspect(m.ctx, id)
	// if err != nil {
	// 	return 0, errors.New("error inspecting volume "+ err.Error())
	// }
	// return foundVolume.ClusterVolume.CreatedAt.Unix(), nil
}