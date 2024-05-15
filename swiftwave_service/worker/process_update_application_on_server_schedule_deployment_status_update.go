package worker

import (
	"context"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
)

func (m Manager) UpdateApplicationOnServerScheduleDeploymentUpdate(_ PersistentVolumeDeletionRequest, ctx context.Context, _ context.CancelFunc) error {
	deploymentInfo, err := core.FindApplicationsForForceUpdate(ctx, m.ServiceManager.DbClient)
	if err != nil {
		logger.WorkerLoggerError.Println("Failed to fetch application deployment info", err.Error())
		return err
	}
	for _, deploymentInfo := range deploymentInfo {
		err = m.EnqueueDeployApplicationRequestWithNoProxyUpdate(deploymentInfo.ApplicationID, deploymentInfo.DeploymentID)
		if err != nil {
			logger.WorkerLoggerError.Println("Failed to enqueue deploy application request", err.Error())
			return err
		}
	}
	return nil
}
