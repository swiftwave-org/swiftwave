package ssh_toolkit

import (
	"fmt"
	"github.com/oklog/ulid"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func CopyFileToRemoteServer(
	localPath string,
	remotePath string,
	host string, port int, user string, privateKey string,
) error {
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
	cmdStr := fmt.Sprintf("rsync -az -e 'ssh -q -p %d -i %s' %s %s@%s:%s", port, tmpFile, localPath, user, host, remotePath)
	cmd := exec.Command(cmdStr)
	return cmd.Run()
}

func CopyFileFromRemoteServer(
	remotePath string,
	localPath string,
	host string, port int, user string, privateKey string,
) error {
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
	cmdStr := fmt.Sprintf("rsync -az -e 'ssh -q -p %d -i %s' %s@%s:%s %s", port, tmpFile, user, host, remotePath, localPath)
	cmd := exec.Command(cmdStr)
	return cmd.Run()
}

// private functions
func storePrivateKeyInTmp(privateKey string) (string, error) {
	filename := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0))
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.pem", filename.String()))
	err := os.WriteFile(tmpFile, []byte(privateKey), 0600)
	if err != nil {
		return "", err
	}
	return tmpFile, nil
}
