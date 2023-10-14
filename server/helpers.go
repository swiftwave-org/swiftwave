package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// Migrate database table
func (server *Server) MigrateDatabaseTables() {
	err := server.DB_CLIENT.AutoMigrate(&Domain{})
	if err != nil {
		panic(err)
	}
	err = server.DB_CLIENT.AutoMigrate(&GitCredential{})
	if err != nil {
		panic(err)
	}
	err = server.DB_CLIENT.AutoMigrate(&ApplicationSource{})
	if err != nil {
		panic(err)
	}
	err = server.DB_CLIENT.AutoMigrate(&Application{})
	if err != nil {
		panic(err)
	}
	err = server.DB_CLIENT.AutoMigrate(&ApplicationBuildLog{})
	if err != nil {
		panic(err)
	}
	err = server.DB_CLIENT.AutoMigrate(&IngressRule{})
	if err != nil {
		panic(err)
	}
	err = server.DB_CLIENT.AutoMigrate(&RedirectRule{})
	if err != nil {
		panic(err)
	}
}

// Generate Git Repository URL from ApplicationSource
func (src ApplicationSource) RepositoryURL() string {
	if src.GitProvider == GitProviderGithub {
		return "https://github.com/" + src.RepositoryUsername + "/" + src.RepositoryName + ".git"
	}
	if src.GitProvider == GitProviderGitlab {
		return "https://gitlab.com/" + src.RepositoryUsername + "/" + src.RepositoryName + ".git"
	}
	return ""
}

// Add a log to the application build log
func (s *Server) AddLogToApplicationBuildLog(log_id string, message string, loglevel string, add_newline bool) {
	var logRecord ApplicationBuildLog
	tx := s.DB_CLIENT.Where("id = ?", log_id).First(&logRecord)
	if tx.Error != nil {
		return
	}
	if add_newline {
		message += "\n"
	}

	// info, warning, error, success
	loglevel = strings.ToLower(loglevel)
	if loglevel == "error" {
		message = fmt.Sprintf("\x1b[1;31m\x1b[0m%s", message)
	} else if loglevel == "warning" {
		message = fmt.Sprintf("\x1b[1;33m\x1b[0m%s", message)
	} else if loglevel == "success" {
		message = fmt.Sprintf("\x1b[1;32m\x1b[0m%s", message)
	}
	// don't need special modification for -> loglevel == "info" || loglevel == ""

	// Appends the log in database
	logRecord.Logs = logRecord.Logs + message
	s.DB_CLIENT.Save(&logRecord)

	// push the logs to redis topic for realtime log streaming
	// topic -> log_update/<log_id>
	s.REDIS_CLIENT.Publish(context.Background(), "log_update/"+log_id, message)
}

// Mark the application build log as completed
// Marking the log as completed will stop the log streaming
func (s *Server) MarkBuildLogAsCompleted(log_id string) {
	var logRecord ApplicationBuildLog
	tx := s.DB_CLIENT.Where("id = ?", log_id).First(&logRecord)
	if tx.Error != nil {
		return
	}
	s.REDIS_CLIENT.Publish(context.Background(), "log_update/"+log_id, "SWIFTWAVE_EOF_LOG")
	logRecord.Completed = true
	s.DB_CLIENT.Save(&logRecord)
}

// Create a default git user if not exists
// Username -> default & Password -> ""
func (s *Server) CreateDefaultGitUser() {
	var git_credential GitCredential
	tx := s.DB_CLIENT.Where("name = ?", "default").First(&git_credential)
	if tx.Error != nil {
		log.Println("`default` git user not found, creating...")
		git_credential.Name = "default"
		git_credential.Username = ""
		git_credential.Password = ""
		tx2 := s.DB_CLIENT.Create(&git_credential)
		if tx2.Error != nil {
			log.Println("Failed to create `default` git user")
			panic(tx2.Error)
		}
	}
}

/*
Generate summary text of Application Code Source for displaying in the UI
For Git -> Show Repository URL & Branch
For Tarball -> Show `Source Code`
For Image -> Show Docker Image Name
*/
func (s ApplicationSource) GetSourceSummary() string {
	if s.Type == ApplicationSourceTypeGit {
		return fmt.Sprintf("%s Branch: %s", s.RepositoryURL(), s.Branch)
	}
	if s.Type == ApplicationSourceTypeTarball {
		return "Source Code"
	}
	if s.Type == ApplicationSourceTypeImage {
		return fmt.Sprintf("Image: %s", s.DockerImage)
	}
	return "Unknown"
}

// Function to check if the server is running in production environment
func (s *Server) isProductionEnvironment() bool {
	return strings.Compare(s.ENVIRONMENT, "production") == 0
}

/*
Function to set application status to `building_image_failed` and update in database
Used when the image build fails
*/
func failImageBuildUpdateStatus(application *Application, db_client gorm.DB) {
	application.Status = ApplicationStatusBuildingImageFailed
	tx := db_client.Save(&application)
	if tx.Error != nil {
		log.Println("Failed to update application status in database")
	}
}

/*
Function to set application status to `deploying_failed` and update in database
Used when the application deploy fails
*/
func failApplicationDeployUpdateStatus(application *Application, db_client gorm.DB) {
	application.Status = ApplicationStatusDeployingFailed
	tx := db_client.Save(&application)
	if tx.Error != nil {
		log.Println("Failed to update application status in database")
	}
}
