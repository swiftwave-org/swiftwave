package ssh_toolkit

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/oklog/ulid"
)

func CopyFolderToRemoteServer(
	localPath string,
	remotePath string,
	host string, port int, user string, privateKey string,
) error {
	if localPath == "" || remotePath == "" || host == "" || port == 0 || user == "" || privateKey == "" {
		return fmt.Errorf("invalid parameters")
	}
	tmpFile, err := storePrivateKeyInTmp(privateKey)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(tmpFile)
		if err != nil {
			fmt.Println("Error removing temporary file:", err)
		}
	}()
	// ensure local path has trailing slash
	if localPath[len(localPath)-1] != '/' {
		localPath = localPath + "/"
	}
	cmd := exec.Command("rsync", "-az", "--delete", "-e", "ssh -q -o StrictHostKeyChecking=no -p "+fmt.Sprintf("%d", port)+" -i "+tmpFile, localPath, user+"@"+host+":"+remotePath)
	cmdErr := cmd.Run()
	if cmdErr != nil {
		return cmdErr
	}
	return nil
}

func CopyFolderFromRemoteServer(
	remotePath string,
	localPath string,
	host string, port int, user string, privateKey string,
) error {
	if localPath == "" || remotePath == "" || host == "" || port == 0 || user == "" || privateKey == "" {
		return fmt.Errorf("invalid parameters")
	}
	tmpFile, err := storePrivateKeyInTmp(privateKey)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(tmpFile)
		if err != nil {
			fmt.Println("Error removing temporary file:", err)
		}
	}()
	// ensure remote path has trailing slash
	if remotePath[len(remotePath)-1] != '/' {
		remotePath = remotePath + "/"
	}
	cmd := exec.Command("rsync", "-az", "--delete", "-e", "ssh -q -o StrictHostKeyChecking=no -p "+fmt.Sprintf("%d", port)+" -i "+tmpFile, user+"@"+host+":"+remotePath, localPath)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func CopyFileToRemoteServer(
	localPath string,
	remotePath string,
	host string, port int, user string, privateKey string,
) error {
	if localPath == "" || remotePath == "" || host == "" || port == 0 || user == "" || privateKey == "" {
		return fmt.Errorf("invalid parameters")
	}
	tmpFile, err := storePrivateKeyInTmp(privateKey)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(tmpFile)
		if err != nil {
			fmt.Println("Error removing temporary file:", err)
		}
	}()
	cmd := exec.Command("rsync", "-z", "-e", "ssh -q -o StrictHostKeyChecking=no -p "+fmt.Sprintf("%d", port)+" -i "+tmpFile, localPath, user+"@"+host+":"+remotePath)
	cmdErr := cmd.Run()
	if cmdErr != nil {
		return cmdErr
	}
	return nil
}

func CopyFileFromRemoteServer(
	remotePath string,
	localPath string,
	host string, port int, user string, privateKey string,
) error {
	if localPath == "" || remotePath == "" || host == "" || port == 0 || user == "" || privateKey == "" {
		return fmt.Errorf("invalid parameters")
	}
	tmpFile, err := storePrivateKeyInTmp(privateKey)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(tmpFile)
		if err != nil {
			fmt.Println("Error removing temporary file:", err)
		}
	}()
	cmd := exec.Command("rsync", "-z", "-e", "ssh -q -o StrictHostKeyChecking=no -p "+fmt.Sprintf("%d", port)+" -i "+tmpFile, user+"@"+host+":"+remotePath, localPath)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// private functions
func storePrivateKeyInTmp(privateKey string) (string, error) {
	privateKey = privateKey + "\n"
	filename := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0))
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.pem", filename.String()))
	err := os.WriteFile(tmpFile, []byte(privateKey), 0600)
	if err != nil {
		return "", err
	}
	return tmpFile, nil
}
