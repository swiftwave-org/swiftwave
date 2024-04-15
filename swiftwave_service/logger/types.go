package logger

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"log"
	"os"
	"strconv"
)

var DatabaseLogger = log.New(os.Stdout, "[DATABASE] ", log.Ldate|log.Ltime|log.LUTC)
var DatabaseLoggerError = log.New(os.Stdout, "[DATABASE] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var CronJobLogger = log.New(os.Stdout, "[CRONJOB] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var CronJobLoggerError = log.New(os.Stdout, "[CRONJOB] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var WorkerLogger = log.New(os.Stdout, "[WORKER] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var WorkerLoggerError = log.New(os.Stdout, "[WORKER] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var GraphQLLogger = log.New(os.Stdout, "[GRAPHQL] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var GraphQLLoggerError = log.New(os.Stdout, "[GRAPHQL] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var HTTPLogger = log.New(os.Stdout, "[HTTP] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var HTTPLoggerError = log.New(os.Stdout, "[HTTP] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var InternalLogger = log.New(os.Stdout, "[INTERNAL] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var InternalLoggerError = log.New(os.Stdout, "[INTERNAL] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var PubSubLogger = log.New(os.Stdout, "[PUBSUB] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var PubSubLoggerError = log.New(os.Stdout, "[PUBSUB] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var TaskQueueLogger = log.New(os.Stdout, "[TASKQUEUE] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var TaskQueueLoggerError = log.New(os.Stdout, "[TASKQUEUE] ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

func init() {
	// try to fetch local config
	isDevelopmentMode := false
	config, err := local_config.Fetch()
	if err == nil {
		isDevelopmentMode = config.IsDevelopmentMode
	}
	// log file path
	infoLogFilePath := local_config.InfoLogFilePath
	errorLogFilePath := local_config.ErrorLogFilePath
	infoLogFile, err := openLogFile(infoLogFilePath)
	if isDevelopmentMode {
		log.Println("Using stdout for info logs in development mode")
		showInfoLogsInStdout()
	} else if err != nil {
		log.Println("Failed to open info log file. Using stdout")
		showInfoLogsInStdout()
	} else {
		DatabaseLogger.SetOutput(infoLogFile)
		WorkerLogger.SetOutput(infoLogFile)
		GraphQLLogger.SetOutput(infoLogFile)
		HTTPLogger.SetOutput(infoLogFile)
		InternalLogger.SetOutput(infoLogFile)
		PubSubLogger.SetOutput(infoLogFile)
		TaskQueueLogger.SetOutput(infoLogFile)
		CronJobLogger.SetOutput(infoLogFile)
	}
	errorLogFile, err := openLogFile(errorLogFilePath)
	if isDevelopmentMode {
		log.Println("Using stdout for error logs in development mode")
		showInfoLogsInStdout()
	} else if err != nil {
		log.Println("Failed to open error log file. Using stdout")
		showErrorLogsInStdout()

	} else {
		DatabaseLoggerError.SetOutput(errorLogFile)
		WorkerLoggerError.SetOutput(errorLogFile)
		GraphQLLoggerError.SetOutput(errorLogFile)
		HTTPLoggerError.SetOutput(errorLogFile)
		InternalLoggerError.SetOutput(errorLogFile)
		PubSubLoggerError.SetOutput(errorLogFile)
		TaskQueueLoggerError.SetOutput(errorLogFile)
		CronJobLoggerError.SetOutput(errorLogFile)
	}
}

func ShowLogsInStdout() {
	showInfoLogsInStdout()
	showErrorLogsInStdout()
}

// Private functions

func openLogFile(path string) (*os.File, error) {
	// Check if the file exists
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		// Rename the file to increment the version > .0, .1
		count := 0
		for {
			newPath := path + "." + strconv.Itoa(count)
			count++
			_, err := os.Stat(newPath)
			if os.IsNotExist(err) {
				err = os.Rename(path, newPath)
				if err != nil {
					return nil, err
				}
				break
			}
		}
	}
	// Create the file
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func showInfoLogsInStdout() {
	DatabaseLogger.SetOutput(os.Stdout)
	WorkerLogger.SetOutput(os.Stdout)
	GraphQLLogger.SetOutput(os.Stdout)
	HTTPLogger.SetOutput(os.Stdout)
	InternalLogger.SetOutput(os.Stdout)
	PubSubLogger.SetOutput(os.Stdout)
	TaskQueueLogger.SetOutput(os.Stdout)
	CronJobLogger.SetOutput(os.Stdout)
}

func showErrorLogsInStdout() {
	DatabaseLoggerError.SetOutput(os.Stdout)
	WorkerLoggerError.SetOutput(os.Stdout)
	GraphQLLoggerError.SetOutput(os.Stdout)
	HTTPLoggerError.SetOutput(os.Stdout)
	InternalLoggerError.SetOutput(os.Stdout)
	PubSubLoggerError.SetOutput(os.Stdout)
	TaskQueueLoggerError.SetOutput(os.Stdout)
	CronJobLoggerError.SetOutput(os.Stdout)
}
