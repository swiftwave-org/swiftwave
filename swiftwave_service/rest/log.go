package rest

import (
	"github.com/labstack/echo/v4"
	"os"
	"path/filepath"
)

// GET /log/<log_file_name>
func (server *Server) fetchLog(c echo.Context) error {
	logFileName := c.Param("log_file_name")
	// clean the log file name
	logFileName = filepath.Clean(logFileName)
	path := filepath.Join(server.Config.LocalConfig.ServiceConfig.LogDirectoryPath, logFileName)
	// Check if the log file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return c.String(404, "log file does not exist")
	}
	// Read and pipe the log file
	c.Response().Header().Set("Content-Type", "text/plain")
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+logFileName)
	return c.File(path)
}
