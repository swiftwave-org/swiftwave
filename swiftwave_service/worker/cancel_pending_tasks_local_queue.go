package worker

import (
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"log"
)

func CancelPendingTasksLocalQueue(config config.Config, manager service_manager.ServiceManager) error {
	if config.SystemConfig.TaskQueueConfig.Mode != system_config.LocalTaskQueue {
		return nil
	}
	tx := manager.DbClient.Begin()
	// Mark all deployment > pending, deployPending  tasks as failed
	tx.Model(&core.Deployment{}).Where("status = ? OR status = ?", core.DeploymentStatusPending, core.DeploymentStatusDeployPending).Update("status", core.DeploymentStatusFailed)
	// Mark all domain SSL > pending tasks as failed
	tx.Model(&core.Domain{}).Where("ssl_status = ?", core.DomainSSLStatusPending).Update("ssl_status", core.DomainSSLStatusFailed)
	// Mark all IngressRule > pending, deleting tasks as failed
	tx.Model(&core.IngressRule{}).Where("status = ? OR status = ?", core.IngressRuleStatusPending, core.IngressRuleStatusDeleting).Update("status", core.IngressRuleStatusFailed)
	// Mark all RedirectRule > pending, deleting tasks as failed
	tx.Model(&core.RedirectRule{}).Where("status = ? OR status = ?", core.RedirectRuleStatusPending, core.RedirectRuleStatusDeleting).Update("status", core.RedirectRuleStatusFailed)
	// Mark all PersistentVolumeBackup > pending tasks as failed
	tx.Model(&core.PersistentVolumeBackup{}).Where("status = ?", core.BackupPending).Update("status", core.BackupFailed)
	// Mark all PersistentVolumeRestore > uploaded tasks as failed
	tx.Model(&core.PersistentVolumeRestore{}).Where("status = ?", core.RestorePending).Update("status", core.RestoreFailed)
	// commit
	err := tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return errors.New("error cancelling pending tasks > " + err.Error())
	} else {
		log.Println("Cancelled pending tasks.\nConfigure `amqp` service to avoid cancellation of tasks on service restart")
		return nil
	}
}
