package logger

import (
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"os"
)

func FetchSystemLogRecords() ([]*string, error) {
	logDirPath := local_config.LogDirectoryPath
	// Check if log directory exists
	_, err := os.Stat(logDirPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("log directory does not exist")
	}
	// Fetch all the file names in the log directory
	files, err := os.ReadDir(logDirPath)
	if err != nil {
		return nil, err
	}
	var logFiles []*string
	for _, file := range files {
		fileName := file.Name()
		logFiles = append(logFiles, &fileName)
	}
	return logFiles, nil
}
