package bootstrap

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/dashboard"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
)

var localConfig *local_config.Config

func loadConfig() error {
	if localConfig != nil {
		return nil
	}
	// Fetch local configuration
	lc, err := local_config.Fetch()
	if err != nil {
		return err
	}
	localConfig = lc
	return nil
}

func StartBootstrapServer() error {
	if err := loadConfig(); err != nil {
		return err
	}

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
	e.HideBanner = true
	e.Use(middleware.CORS())
	// Setup routes
	e.POST("/setup", SystemSetupHandler)
	// Register dashboard
	dashboard.RegisterHandlers(e, true)
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
			"message": err.Error(),
		})
	}
	// Convert to DB record
	systemConfig, err := payloadToDBRecord(*systemConfigReq)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": err.Error(),
		})
	}
	// Save to DB
	if err := dbClient.Create(&systemConfig).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to save system configuration",
		})
	}
	// Restart swiftwave service
	go func() {
		// wait for 2 seconds
		<-time.After(2 * time.Second)
		color.Green("Restarting swiftwave service")
		color.Yellow("Swiftwave service will be restarted in 5 seconds")
		color.Yellow("If you are running without enabling service, run `swiftwave start` to start the service")
		_ = exec.Command("systemctl", "restart", "swiftwave.service").Run()
		os.Exit(1)
	}()
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "System setup completed successfully",
	})
}

// FetchSystemConfigHandler : Fetch system configuration handler
// GET /config/system
func FetchSystemConfigHandler(c echo.Context) error {
	if err := loadConfig(); err != nil {
		return err
	}
	// Fetch system configuration
	dbClient, err := db.GetClient(localConfig, 1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to connect to database",
		})
	}
	sysConfig, err := system_config.Fetch(dbClient)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to fetch system configuration",
		})
	}
	// Hide sensitive fields
	payload := dbRecordToPayload(sysConfig)
	return c.JSON(http.StatusOK, payload)
}

// UpdateSystemConfigHandler : Update system configuration handler
// PUT /config/system
func UpdateSystemConfigHandler(c echo.Context) error {
	if err := loadConfig(); err != nil {
		return err
	}
	dbClient, err := db.GetClient(localConfig, 1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to connect to database",
		})
	}
	sysConfig, err := system_config.Fetch(dbClient)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to fetch system configuration",
		})
	}
	// Update system configuration
	systemConfigReq := new(SystemConfigurationPayload)
	if err := c.Bind(systemConfigReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request payload",
		})
	}
	// Inject some fields
	systemConfigReq.JWTSecretKey = sysConfig.JWTSecretKey
	systemConfigReq.HAProxyConfig.Username = sysConfig.HAProxyConfig.Username
	systemConfigReq.HAProxyConfig.Password = sysConfig.HAProxyConfig.Password
	systemConfigReq.SSHPrivateKey = sysConfig.SshPrivateKey
	systemConfigReq.LetsEncrypt.PrivateKey = sysConfig.LetsEncryptConfig.PrivateKey
	// Convert to DB record
	systemConfig, err := payloadToDBRecord(*systemConfigReq)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request payload",
		})
	}
	// Update DB record
	if err := systemConfig.Update(dbClient); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to update system configuration",
		})
	}
	// Restart swiftwave service
	go func() {
		// wait for 2 seconds
		<-time.After(2 * time.Second)
		color.Green("Restarting swiftwave service")
		color.Yellow("Swiftwave service will be restarted in 5 seconds")
		color.Yellow("If you are running without enabling service, run `swiftwave start` to start the service")
		_ = exec.Command("systemctl", "restart", "swiftwave.service").Run()
		os.Exit(1)
	}()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "System configuration updated successfully",
	})
}
