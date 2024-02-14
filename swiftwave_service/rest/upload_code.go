package rest

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

// Upload tar file and return the file name
// POST /upload/code
func (server *Server) uploadTarFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	src, err := file.Open()
	if err != nil {
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Println(err)
		}
	}(src)

	// Check if file is tar
	if file.Header.Get("Content-Type") != "application/x-tar" {
		return c.JSON(400, map[string]string{
			"message": "file is not a tar file",
		})
	}

	// Destination
	destFilename := uuid.New().String() + ".tar"
	destFile := filepath.Join(server.SystemConfig.ServiceConfig.DataDir, destFilename)
	dst, err := os.Create(destFile)
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "failed to create file",
		})
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			log.Println(err)
		}
	}(dst)

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		log.Println(err)
		return c.JSON(500, map[string]string{
			"message": "failed to copy file",
		})
	}

	// Return file name
	return c.JSON(200, map[string]string{
		"file":    destFilename,
		"message": "file uploaded successfully",
	})
}
