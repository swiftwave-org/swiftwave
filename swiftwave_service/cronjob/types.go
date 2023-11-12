package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/system_config"
	"sync"
)

type CronJob interface {
	Start(nowait bool)
	Wait()
}

type Manager struct {
	SystemConfig   *system_config.Config
	ServiceManager *core.ServiceManager
	wg             *sync.WaitGroup
}
