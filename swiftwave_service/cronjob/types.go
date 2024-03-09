package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"sync"
)

type CronJob interface {
	Start(nowait bool)
	Wait()
}

type Manager struct {
	SystemConfig   *local_config.Config
	ServiceManager *core.ServiceManager
	wg             *sync.WaitGroup
}
