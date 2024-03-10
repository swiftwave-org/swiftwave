package worker

import (
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"time"
)

func (m Manager) SetupServer(request SetupServerRequest, ctx context.Context, _ context.CancelFunc) error {
	// fetch server
	server, err := core.FetchServerByID(&m.ServiceManager.DbClient, request.ServerId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// fetch server log
	serverLog, err := core.FetchServerLogByID(&m.ServiceManager.DbClient, request.LogId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// log
	logText := "Preparing server for deployment\n"
	// spawn a goroutine to update server log each 5 seconds
	go func() {
		lastSent := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if time.Since(lastSent) > 5*time.Second {
					serverLog.Content = logText
					_ = serverLog.Update(&m.ServiceManager.DbClient)
					lastSent = time.Now()
				}
			}
		}
	}()
	// defer to push final log
	defer func() {
		serverLog.Content = logText
		_ = serverLog.Update(&m.ServiceManager.DbClient)
	}()

	// Proceed request logic (reject in any other case)
	_ = server
	// TODO
	// Create all the required directories, use ssh exec with sudo
	// Add node to swarm cluster
	// set status

	return nil
}
