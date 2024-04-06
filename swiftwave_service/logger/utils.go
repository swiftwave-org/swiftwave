package logger

import (
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"os"
	"time"
)

type FileInfo struct {
	Name    string
	ModTime time.Time
}

func FetchSystemLogRecords() ([]*FileInfo, error) {
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
	var logFiles []*FileInfo
	for _, file := range files {
		fileName := file.Name()
		fileModTime := time.Now()
		x, err := file.Info()
		if err == nil {
			fileModTime = x.ModTime()
		}
		fileInfo := FileInfo{
			Name:    fileName,
			ModTime: fileModTime,
		}
		logFiles = append(logFiles, &fileInfo)
	}
	return logFiles, nil
}
