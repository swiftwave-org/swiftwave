package server

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func (src ApplicationSource) RepositoryURL() string {
	if src.GitProvider == GitProviderGithub {
		return "https://github.com/" + src.RepositoryUsername + "/" + src.RepositoryName + ".git"
	}
	if src.GitProvider == GitProviderGitlab {
		return "https://gitlab.com/" + src.RepositoryUsername + "/" + src.RepositoryName + ".git"
	}
	return ""
}

func (s *Server) AddLogToApplicationBuildLog(log_id string, message string, loglevel string) {
	var logRecord ApplicationBuildLog
	tx := s.DB_CLIENT.Where("id = ?", log_id).First(&logRecord)
	if tx.Error != nil {
		return
	}
	logRecord.Logs += fmt.Sprintf("\n[%s]-[%s] %s", time.Now(), strings.ToUpper(loglevel), message)
	s.DB_CLIENT.Save(&logRecord)
}

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


func (s *Server) isProductionEnvironment() bool {
	return strings.Compare(s.ENVIRONMENT, "production") == 0;
}