package worker

import (
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/system_config"
	"log"
)

func CancelPendingTasksLocalQueue(config system_config.Config, manager core.ServiceManager) error {
	if config.TaskQueueConfig.Mode != system_config.LocalTaskQueue {
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
