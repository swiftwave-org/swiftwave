package bootstrap

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
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
	systemConfig, err := payloadToDBRecord(*systemConfigReq)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request payload",
		})
	}
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

// FetchSystemConfigHandler : Fetch system configuration handler
// GET /setup
func FetchSystemConfigHandler(c echo.Context) error {
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
// PUT /setup
func UpdateSystemConfigHandler(c echo.Context) error {
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
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "System configuration updated successfully",
	})
}
