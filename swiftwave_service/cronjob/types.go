package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"sync"
)

type CronJob interface {
	Start(nowait bool)
	Wait()
}

type Manager struct {
	Config         *config.Config
	ServiceManager *service_manager.ServiceManager
	wg             *sync.WaitGroup
}
