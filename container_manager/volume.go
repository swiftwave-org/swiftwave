package containermanger

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var volumeToolkitImage = "ghcr.io/swiftwave-org/volume-toolkit:latest"

// CreateLocalVolume : Create a new volume, return id of the volume
func (m Manager) CreateLocalVolume(name string) error {
	_, err := m.client.VolumeCreate(m.ctx, volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return errors.New("error creating volume " + err.Error())
	}
	return nil
}

// CreateNFSVolume : Create a new NFS volume, return id of the volume
func (m Manager) CreateNFSVolume(name string, nfsServer string, nfsPath string, version int) error {
	_, err := m.client.VolumeCreate(m.ctx, volume.CreateOptions{
		Name:   name,
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "nfs",
			"o":      "addr=" + nfsServer + ",rw,nfsvers=" + fmt.Sprint(version),
			"device": ":" + nfsPath,
		},
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
	var volumeNames = make([]string, len(volumes.Volumes))
	for i, v := range volumes.Volumes {
		volumeNames[i] = v.Name
	}
	return volumeNames, nil
}

// SizeVolume : Return a file which will contain the size of the volume in bytes
func (m Manager) SizeVolume(volumeName string, host string, port int, user string, privateKey string) (sizeMB float64, err error) {
	path, err := m.volumeToolkitRunner(volumeName, "size", nil, false)
	if err != nil {
		return 0, err
	}
	defer func(path string) {
		_ = ssh_toolkit.ExecCommandOverSSH(fmt.Sprintf("rm -rf %s", path), nil, nil, 10, host, port, user, privateKey, 20)
	}(path)
	resultPath := filepath.Join(path, "size.txt")
	// fetch the size
	stdoutBuf := &bytes.Buffer{}
	err = ssh_toolkit.ExecCommandOverSSH(fmt.Sprintf("cat %s", resultPath), stdoutBuf, nil, 10, host, port, user, privateKey, 20)
	if err != nil {
		return 0, errors.New("failed to read size file " + err.Error())
	}
	sizeBytes := 0
	_, err = fmt.Sscanf(stdoutBuf.String(), "%d", &sizeBytes)
	if err != nil {
		return 0, errors.New("failed to parse size " + err.Error())
	}
	sizeMB = float64(sizeBytes) / 1024 / 1024
	return sizeMB, nil
}

// BackupVolume : Backup a volume to a file
func (m Manager) BackupVolume(volumeName string, backupFilePath string, host string, port int, user string, privateKey string) error {
	if !strings.HasSuffix(backupFilePath, ".tar.gz") {
		return errors.New("backupFilePath should have .tar.gz extension")
	}
	path, err := m.volumeToolkitRunner(volumeName, "export", nil, false)
	if err != nil {
		return err
	}
	defer func(path string) {
		_ = ssh_toolkit.ExecCommandOverSSH(fmt.Sprintf("rm -rf %s", path), nil, nil, 10, host, port, user, privateKey, 20)
	}(path)
	dumpFilePath := filepath.Join(path, "dump.tar.gz")
	if err != nil {
		return errors.New("failed to change permission of dump file " + err.Error())
	}
	// copy the backup file to the backupFilePath
	err = ssh_toolkit.CopyFileFromRemoteServer(dumpFilePath, backupFilePath, host, port, user, privateKey)
	if err != nil {
		return errors.New("failed to copy backup file " + err.Error())
	}
	// make the backup file read only
	err = os.Chmod(backupFilePath, 0660)
	if err != nil {
		// delete the backup file
		_ = os.Remove(backupFilePath)
		return errors.New("failed to change permission of dump file " + err.Error())
	}
	return nil
}

// RestoreVolume : Restore a volume from a backup file
func (m Manager) RestoreVolume(volumeName string, backupFilePath string, host string, port int, user string, privateKey string) error {
	if !strings.HasSuffix(backupFilePath, ".tar.gz") {
		return errors.New("backupFilePath should have .tar.gz extension")
	}
	// fetch file name
	_, fileName := filepath.Split(backupFilePath)
	// remove extension
	fileName = strings.TrimSuffix(fileName, ".tar.gz")
	// prepare paths
	outputPath := filepath.Join("/tmp", fileName)
	// create temp directory
	err := ssh_toolkit.ExecCommandOverSSH(fmt.Sprintf("mkdir -p %s", outputPath), nil, nil, 10, host, port, user, privateKey, 20)
	if err != nil {
		return errors.New("failed to create directory for moving backup file to server  " + err.Error())
	}
	// remove temp directory
	defer func(outputPath string) {
		_ = ssh_toolkit.ExecCommandOverSSH(fmt.Sprintf("rm -rf %s", outputPath), nil, nil, 10, host, port, user, privateKey, 20)
	}(outputPath)
	// file path
	dumpFilePath := filepath.Join(outputPath, "dump.tar.gz")
	// copy the backup file to the server
	err = ssh_toolkit.CopyFileToRemoteServer(backupFilePath, dumpFilePath, host, port, user, privateKey)
	if err != nil {
		return errors.New("failed to copy backup file to server " + err.Error())
	}
	// run the volume toolkit
	_, err = m.volumeToolkitRunner(volumeName, "import", &outputPath, true)
	return err
}

// private  function

// volumeToolkitRunner : Run the volume toolkit container
func (m Manager) volumeToolkitRunner(volumeName string, command string, predefinedOutputPath *string, dataDirectoryRW bool) (string, error) {
	// check if volume exists
	if !m.ExistsVolume(volumeName) {
		return "", errors.New("volume does not exist")
	}

	// pull image if not exists
	if !m.ExistsImage(volumeToolkitImage) {
		resReader, err := m.client.ImagePull(m.ctx, volumeToolkitImage, types.ImagePullOptions{})
		if err != nil {
			return "", errors.New("failed to pull image " + err.Error())
		}
		// read the response but ignore it
		_, err = io.Copy(io.Discard, resReader)
		if err != nil {
			return "", errors.New("failed to pull image > response " + err.Error())
		}
	}
	// create temp directory
	if predefinedOutputPath == nil {
		outputPath, err := os.MkdirTemp("", "swiftwave-volume-toolkit-*")
		if err != nil {
			return "", errors.New("failed to create temp directory " + err.Error())
		}
		predefinedOutputPath = &outputPath
	}
	var binds []string
	if dataDirectoryRW {
		binds = []string{fmt.Sprintf("%s:/data:rw", volumeName), fmt.Sprintf("%s:/app:rw", *predefinedOutputPath)}
	} else {
		binds = []string{fmt.Sprintf("%s:/data:ro", volumeName), fmt.Sprintf("%s:/app:rw", *predefinedOutputPath)}
	}

	createRes, err := m.client.ContainerCreate(m.ctx, &container.Config{
		Image:           volumeToolkitImage,
		AttachStdin:     false,
		AttachStdout:    true,
		AttachStderr:    true,
		Tty:             false,
		Cmd:             []string{command},
		NetworkDisabled: true,
	}, &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: network.NetworkNone,
		Privileged:  false,
		Binds:       binds,
	}, nil, nil, "")
	if err != nil {
		return "", errors.New("failed to create container " + err.Error())
	}
	// start the container
	err = m.client.ContainerStart(m.ctx, createRes.ID, container.StartOptions{})
	if err != nil {
		return "", errors.New("failed to start container " + err.Error())
	}

	// wait for the container to finish
	waitRes, waitErr := m.client.ContainerWait(m.ctx, createRes.ID, container.WaitConditionRemoved)
	for {
		select {
		case err := <-waitErr:
			return "", errors.New("failed to wait for container " + err.Error())
		case res := <-waitRes:
			if res.Error != nil {
				return "", errors.New("container error " + res.Error.Message)
			}
			return *predefinedOutputPath, nil
		}
	}
}
