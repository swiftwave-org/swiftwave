package dockermanager

import "fmt"

func (m Manager) VolumeExists(name string) bool {
	volume, err := m.client.VolumeInspect(m.ctx, name);
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(volume)
	if volume.Name == name {
		return true
	}
	return false
}