package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"sync"
)

type CronJob interface {
	Start(nowait bool)
	Wait()
}

type Manager struct {
	ServiceConfig  *core.ServiceConfig
	ServiceManager *core.ServiceManager
	wg             *sync.WaitGroup
}
