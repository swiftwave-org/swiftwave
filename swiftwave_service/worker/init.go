package worker

import (
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
	workerManager := Manager{
		Config:         config,
		ServiceManager: manager,
	}
	workerManager.registerWorkerFunctions()
	return &workerManager
}

func (m Manager) StartConsumers(nowait bool) error {
	return m.ServiceManager.TaskQueueClient.StartConsumers(nowait)
}

func (m Manager) WaitForConsumers() {
	m.ServiceManager.TaskQueueClient.WaitForConsumers()
}

// private functions
// registerWorkerFunctions : register all the functions to the task queue client
func (m Manager) registerWorkerFunctions() {
	taskQueueClient := m.ServiceManager.TaskQueueClient
	panicOnError(taskQueueClient.RegisterFunction(buildApplicationQueueName, m.BuildApplication))
	panicOnError(taskQueueClient.RegisterFunction(deployApplicationQueueName, m.DeployApplication))
	panicOnError(taskQueueClient.RegisterFunction(deleteApplicationQueueName, m.DeleteApplication))
	panicOnError(taskQueueClient.RegisterFunction(ingressRuleApplyQueueName, m.IngressRuleApply))
	panicOnError(taskQueueClient.RegisterFunction(ingressRuleDeleteQueueName, m.IngressRuleDelete))
	panicOnError(taskQueueClient.RegisterFunction(redirectRuleApplyQueueName, m.RedirectRuleApply))
	panicOnError(taskQueueClient.RegisterFunction(redirectRuleDeleteQueueName, m.RedirectRuleDelete))
	panicOnError(taskQueueClient.RegisterFunction(sslGenerateQueueName, m.SSLGenerate))
	panicOnError(taskQueueClient.RegisterFunction(persistentVolumeBackupQueueName, m.PersistentVolumeBackup))
	panicOnError(taskQueueClient.RegisterFunction(persistentVolumeRestoreQueueName, m.PersistentVolumeRestore))
	panicOnError(taskQueueClient.RegisterFunction(installDependenciesOnServerQueueName, m.InstallDependenciesOnServer))
	panicOnError(taskQueueClient.RegisterFunction(setupServerQueueName, m.SetupServer))

}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
