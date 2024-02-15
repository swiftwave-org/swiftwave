package rest

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/uploader"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	if persistentVolumeBackup.Type == core.LocalBackup {
		// send file
		filePath := filepath.Join(server.SystemConfig.ServiceConfig.DataDir, persistentVolumeBackup.File)
		// file name
		fileName := persistentVolumeBackup.File
		c.Request().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		return c.Attachment(filePath, fileName)
	} else if persistentVolumeBackup.Type == core.S3Backup {
		s3config := server.SystemConfig.PersistentVolumeBackupConfig.S3Config
		if !s3config.Enabled {
			return c.String(400, "S3 backup is not enabled")
		}
		// download file from s3
		s3Client, err := uploader.GenerateS3Client(s3config)
		if err != nil {
			return c.String(500, "Internal server error")
		}
		// download file
		resp, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(s3config.Bucket),
			Key:    aws.String(persistentVolumeBackup.File),
		})
		if err != nil {
			return c.String(500, "Internal server error")
		}
		defer func(resp *s3.GetObjectOutput) {
			err := resp.Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(resp)
		// send file
		if resp.ContentLength != nil {
			contentLength, err := strconv.ParseInt(strconv.FormatInt(*resp.ContentLength, 10), 10, 64)
			if err != nil {
				return c.String(500, "Internal server error")
			}
			c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
		}
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", persistentVolumeBackup.File))
		return c.Stream(200, "application/octet-stream", resp.Body)
	} else {
		return c.String(500, "Internal server error")
	}
}

// POST /persistent-volume/restore/:id/upload
func (server *Server) uploadPersistentVolumeRestoreFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	// fetch persistent volume restore
	idStr := c.Param("id")
	// convert id to uint
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(400, map[string]string{
			"message": "Invalid id",
		})
	}
	var persistentVolumeRestore core.PersistentVolumeRestore
	err = persistentVolumeRestore.FindById(c.Request().Context(), server.ServiceManager.DbClient, uint(id))
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "Internal server error",
		})
	}
	if persistentVolumeRestore.Status != core.RestorePending {
		return c.JSON(400, map[string]string{
			"message": "Sorry, you can't upload file for this restore anymore",
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
	// Check if filename ends with .tar.gz
	if !strings.HasSuffix(file.Filename, ".tar.gz") {
		return c.JSON(400, map[string]string{
			"message": "file is not a tar.gz file",
		})
	}
	// Destination
	fileName := fmt.Sprintf("restore-%s-%d.tar.gz", uuid.NewString(), persistentVolumeRestore.ID)
	filePath := filepath.Join(server.SystemConfig.ServiceConfig.DataDir, fileName)
	// Write file
	dst, err := os.Create(filePath)
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
	// update persistent volume restore
	persistentVolumeRestore.File = fileName
	persistentVolumeRestore.Status = core.RestoreUploaded
	err = persistentVolumeRestore.Update(c.Request().Context(), server.ServiceManager.DbClient, server.SystemConfig.ServiceConfig.DataDir)
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "failed to update restore",
		})
	}
	return c.JSON(200, map[string]string{
		"message": "file uploaded successfully, you can now start the restore process",
	})
}