package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/labstack/gommon/random"
	"github.com/tredoe/osutil/user/crypt"
	"github.com/tredoe/osutil/user/crypt/sha256_crypt"
	"golang.org/x/exp/rand"
)

func RunCommand(command string) (input string, output string, err error) {
	cmd := exec.Command("bash", "-c", command)
	stdoutBuffer := bytes.NewBuffer([]byte{})
	stderrBuffer := bytes.NewBuffer([]byte{})
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer
	err = cmd.Run()
	return strings.TrimSpace(stdoutBuffer.String()), strings.TrimSpace(stderrBuffer.String()), err
}

func RunCommandWithoutBuffer(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetCPUArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	case "386":
		return "i686"
	default:
		return "unknown"
	}
}

func GenerateBasicAuthPassword(password string) (string, error) {
	c := crypt.New(crypt.SHA256)
	s := sha256_crypt.GetSalt()
	randomSalt := random.String(5)
	saltString := fmt.Sprintf("%s%s", s.MagicPrefix, randomSalt)
	return c.Generate([]byte(password), []byte(saltString))
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func GetServiceStatus(serviceName string) bool {
	_, _, err := RunCommand("systemctl is-active " + serviceName)
	return err == nil
}
