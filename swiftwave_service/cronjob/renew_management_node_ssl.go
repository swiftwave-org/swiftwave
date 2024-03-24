package cronjob

import (
	"bytes"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"os"
	"os/exec"
	"time"
)

func (m Manager) RenewManagementNodeSSL() {
	logger.CronJobLogger.Println("Starting renew management node SSL [cronjob]")
	for {
		m.renewManagementNodeSSL()
		time.Sleep(24 * time.Hour)
	}
}

func (m Manager) renewManagementNodeSSL() {
	// get current executable
	executablePath, err := os.Executable()
	if err != nil {
		logger.CronJobLoggerError.Println("Error while fetching executable path \n", err)
		return
	}
	// Run `tls renew` command
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd := exec.Command(executablePath, "tls", "renew")
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	err = cmd.Run()
	if err != nil {
		logger.CronJobLoggerError.Println("Error while renewing management node SSL \n", err, "\n", stderrBuf.String())
	} else {
		logger.CronJobLogger.Println("Management node SSL renewed successfully")
	}
}
