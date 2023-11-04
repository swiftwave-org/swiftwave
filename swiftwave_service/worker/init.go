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
	panicOnError(taskQueueClient.RegisterFunction("build_application", m.BuildApplication))
	panicOnError(taskQueueClient.RegisterFunction("deploy_application", m.DeployApplication))
	panicOnError(taskQueueClient.RegisterFunction("ingress_rule_apply", m.IngressRuleApply))
	panicOnError(taskQueueClient.RegisterFunction("redirect_rule_apply", m.RedirectRuleApply))
	panicOnError(taskQueueClient.RegisterFunction("ssl_generate", m.SSLGenerate))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
