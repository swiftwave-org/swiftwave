package worker

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

func NewManager(config *core.ServiceConfig, manager *core.ServiceManager) *Manager {
	if config == nil {
		panic("config cannot be nil")
	}
	if manager == nil {
		panic("manager cannot be nil")
	}
	workerManager := Manager{
		ServiceConfig:  config,
		ServiceManager: manager,
	}
	workerManager.registerWorkerFunctions()
	return &workerManager
}

// private functions
// registerWorkerFunctions : register all the functions to the task queue client
func (m Manager) registerWorkerFunctions() {
	taskQueueClient := m.ServiceManager.TaskQueueClient
	panicOnError(taskQueueClient.RegisterFunction(buildApplicationQueueName, m.BuildApplication))
	panicOnError(taskQueueClient.RegisterFunction(deployApplicationQueueName, m.DeployApplication))
	panicOnError(taskQueueClient.RegisterFunction(ingressRuleApplyQueueName, m.IngressRuleApply))
	panicOnError(taskQueueClient.RegisterFunction(ingressRuleDeleteQueueName, m.IngressRuleDelete))
	panicOnError(taskQueueClient.RegisterFunction(redirectRuleApplyQueueName, m.RedirectRuleApply))
	panicOnError(taskQueueClient.RegisterFunction(redirectRuleDeleteQueueName, m.RedirectRuleDelete))
	panicOnError(taskQueueClient.RegisterFunction(sslGenerateQueueName, m.SSLGenerate))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
