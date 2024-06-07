package main

import (
	"ariga.io/atlas-provider-gorm/gormschema"
	"fmt"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/task_queue"
	"io"
	"os"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(&system_config.SystemConfig{},
		&core.Server{},
		&core.ServerLog{},
		&core.User{},
		&core.Domain{},
		&core.RedirectRule{},
		&core.PersistentVolume{},
		&core.ConfigMount{},
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
		&task_queue.EnqueuedTask{})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	_, _ = io.WriteString(os.Stdout, stmts)
}
