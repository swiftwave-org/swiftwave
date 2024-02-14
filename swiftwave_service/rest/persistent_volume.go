package rest

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"path/filepath"
	"strconv"
)

// GET /persistent-volume/backup/:id/download
func (server *Server) downloadPersistentVolumeBackup(c echo.Context) error {
	idStr := c.Param("id")
	// convert id to uint
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.String(400, "Invalid id")
	}
	// fetch persistent volume backup
	var persistentVolumeBackup core.PersistentVolumeBackup
	err = persistentVolumeBackup.FindById(c.Request().Context(), server.ServiceManager.DbClient, uint(id))
	if err != nil {
		return c.String(500, "Internal server error")
	}
	// check status should be success
	if persistentVolumeBackup.Status != core.BackupSuccess {
		return c.String(400, "Sorry, backup is not available for download")
	}
	// send file
	filePath := filepath.Join(server.SystemConfig.ServiceConfig.DataDir, persistentVolumeBackup.File)
	// file name
	fileName := fmt.Sprintf("backup-%s-%d.tar.gz", persistentVolumeBackup.File, persistentVolumeBackup.ID)
	return c.Attachment(filePath, fileName)
}
