package server

import (
	"fmt"
	"strings"
	"time"
)

func (src ApplicationSource) RepositoryURL() string {
	if src.GitProvider == GitProviderGithub {
		return "https://github.com/"+src.RepositoryUsername+"/"+src.RepositoryName+".git"
	}
	if src.GitProvider == GitProviderGitlab {
		return "https://gitlab.com/"+src.RepositoryUsername+"/"+src.RepositoryName+".git"
	}
	return ""
}


func (s *Server) AddLogToApplicationDeployLog(log_id string, message string, loglevel string){
	var logRecord ApplicationDeployLog
	tx := s.DB_CLIENT.Where("id = ?", log_id).First(&logRecord)
	if tx.Error != nil {
		return
	}
	logRecord.Logs += fmt.Sprintf("\n[%s]-[%s] %s", time.Now(), strings.ToUpper(loglevel), message)
	s.DB_CLIENT.Save(&logRecord)
}