package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	SSL "github.com/swiftwave-org/swiftwave/pkg/ssl_manager"
	"github.com/swiftwave-org/swiftwave/pkg/task_queue"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(&system_config.SystemConfig{},
		&core.Server{},
		&core.ServerLog{},
		&core.User{},
		&core.UserSession{},
		&core.Domain{},
		&core.RedirectRule{},
		&core.PersistentVolume{},
		&core.ConfigMount{},
		&core.ApplicationGroup{},
		&core.Application{},
		&core.GitCredential{},
		&core.ImageRegistryCredential{},
		&core.IngressRule{},
		&core.EnvironmentVariable{},
		&core.PersistentVolumeBinding{},
		&core.Deployment{},
		&core.BuildArg{},
		&core.DeploymentLog{},
		&SSL.KeyAuthorizationToken{},
		&core.PersistentVolumeBackup{},
		&core.PersistentVolumeRestore{},
		&core.ConsoleToken{},
		&core.AnalyticsServiceToken{},
		&core.ServerResourceStat{},
		&core.ApplicationServiceResourceStat{},
		&core.AppBasicAuthAccessControlList{},
		&core.AppBasicAuthAccessControlUser{},
		&task_queue.EnqueuedTask{},
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	_, _ = io.WriteString(os.Stdout, stmts)
}
