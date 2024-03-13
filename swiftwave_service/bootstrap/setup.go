package bootstrap

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"net/http"
	"os"
	"os/exec"
)

var localConfig *local_config.Config

func StartBootstrapServer() error {
	// Fetch local configuration
	lc, err := local_config.Fetch()
	if err != nil {
		return err
	}
	localConfig = lc
	// Pre-check if system setup is required
	setupRequired, err := IsSystemSetupRequired()
	if err != nil {
		return err
	}
	if !setupRequired {
		return errors.New("system setup already completed")
	}
	// Create echo instance
	e := echo.New()
	// Setup routes
	e.GET("/setup", SystemSetupHandler)
	// Start server
	return e.Start(fmt.Sprintf("%s:%d", localConfig.ServiceConfig.BindAddress, localConfig.ServiceConfig.BindPort))
}

// SystemSetupHandler : System setup handler
// POST /setup
func SystemSetupHandler(c echo.Context) error {
	// Rerun the setup check to ensure that the setup is still required
	setupRequired, err := IsSystemSetupRequired()
	if err != nil {
		return err
	}
	if !setupRequired {
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"message": "System setup already completed",
		})
	}
	// Create DB client
	dbClient, err := db.GetClient(localConfig, 1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to connect to database",
		})
	}
	// Create system configuration
	systemConfigReq := new(SystemConfigurationPayload)
	if err := c.Bind(systemConfigReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request payload",
		})
	}
	// Convert to DB record
	systemConfig := payloadToDBRecord(*systemConfigReq)
	// Save to DB
	if err := dbClient.Create(&systemConfig).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to save system configuration",
		})
	}
	// Restart swiftwave service
	defer func() {
		_ = exec.Command("systemctl", "restart", "swiftwave.service").Run()
		os.Exit(1)
	}()
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "System setup completed successfully",
	})
}
