package worker

import (
	"context"
	"sync"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
)

func NewManager(config *config.Config, manager *service_manager.ServiceManager) *Manager {
	if config == nil {
		panic("config cannot be nil")
	}
	if manager == nil {
		panic("manager cannot be nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	workerManager := &Manager{
		Config:         config,
		ServiceManager: manager,
		ctx:            ctx,
		cancel:         cancel,
		wg:             wg,
	}
	workerManager.registerWorkerFunctions()
	wg.Add(1)
	go bulkInsertDeploymentLogs(ctx, wg, manager.DbClient)
	return workerManager
}

// Shutdown cancels background goroutines (e.g. the deployment-log batcher)
// and blocks until they finish flushing. Safe to call multiple times.
func (m *Manager) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.wg != nil {
		m.wg.Wait()
	}
}

func (m Manager) StartConsumers(nowait bool) error {
	return m.ServiceManager.TaskQueueClient.StartConsumers(nowait)
}

func (m Manager) WaitForConsumers() {
	m.ServiceManager.TaskQueueClient.WaitForConsumers()
}

func (m Manager) registerWorkerFunctions() {
	taskQueueClient := m.ServiceManager.TaskQueueClient
	panicOnError(taskQueueClient.RegisterFunction(buildApplicationQueueName, m.BuildApplication))
	panicOnError(taskQueueClient.RegisterFunction(deployApplicationQueueName, m.DeployApplication))
	panicOnError(taskQueueClient.RegisterFunction(deleteApplicationQueueName, m.DeleteApplication))
	panicOnError(taskQueueClient.RegisterFunction(ingressRuleApplyQueueName, m.IngressRuleApply))
	panicOnError(taskQueueClient.RegisterFunction(ingressRuleDeleteQueueName, m.IngressRuleDelete))
	panicOnError(taskQueueClient.RegisterFunction(ingressRuleHttpsRedirectQueueName, m.IngressRuleHttpsRedirect))
	panicOnError(taskQueueClient.RegisterFunction(redirectRuleApplyQueueName, m.RedirectRuleApply))
	panicOnError(taskQueueClient.RegisterFunction(redirectRuleDeleteQueueName, m.RedirectRuleDelete))
	panicOnError(taskQueueClient.RegisterFunction(sslGenerateQueueName, m.SSLGenerate))
	panicOnError(taskQueueClient.RegisterFunction(sslProxyUpdateQueueName, m.SSLProxyUpdate))
	panicOnError(taskQueueClient.RegisterFunction(deletePersistentVolumeQueueName, m.PersistentVolumeDeletion))
	panicOnError(taskQueueClient.RegisterFunction(persistentVolumeBackupQueueName, m.PersistentVolumeBackup))
	panicOnError(taskQueueClient.RegisterFunction(persistentVolumeRestoreQueueName, m.PersistentVolumeRestore))
	panicOnError(taskQueueClient.RegisterFunction(installDependenciesOnServerQueueName, m.InstallDependenciesOnServer))
	panicOnError(taskQueueClient.RegisterFunction(setupServerQueueName, m.SetupServer))
	panicOnError(taskQueueClient.RegisterFunction(setupAndEnableProxyQueueName, m.SetupAndEnableProxy))
	panicOnError(taskQueueClient.RegisterFunction(updateApplicationOnServerScheduleDeploymentUpdateQueueName, m.UpdateApplicationOnServerScheduleDeploymentUpdate))
	// When adding a new function, add it to the list of Queues() as well
}

func Queues() []string {
	return []string{
		buildApplicationQueueName,
		deployApplicationQueueName,
		deleteApplicationQueueName,
		ingressRuleApplyQueueName,
		ingressRuleDeleteQueueName,
		ingressRuleHttpsRedirectQueueName,
		redirectRuleApplyQueueName,
		redirectRuleDeleteQueueName,
		sslGenerateQueueName,
		sslProxyUpdateQueueName,
		deletePersistentVolumeQueueName,
		persistentVolumeBackupQueueName,
		persistentVolumeRestoreQueueName,
		installDependenciesOnServerQueueName,
		setupServerQueueName,
		setupAndEnableProxyQueueName,
		updateApplicationOnServerScheduleDeploymentUpdateQueueName,
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
