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
	"time"
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
		filePath := filepath.Join(server.Config.LocalConfig.ServiceConfig.PVBackupDirectoryPath, persistentVolumeBackup.File)
		// file name
		fileName := persistentVolumeBackup.File
		c.Request().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		return c.Attachment(filePath, fileName)
	} else if persistentVolumeBackup.Type == core.S3Backup {
		s3config := server.Config.SystemConfig.PersistentVolumeBackupConfig.S3BackupConfig
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

// GET /persistent-volume/backup/:id/filename
func (server *Server) getPersistentVolumeBackupFileName(c echo.Context) error {
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
	return c.String(200, persistentVolumeBackup.File)
}

// POST /persistent-volume/:id/restore
func (server *Server) uploadPersistentVolumeRestoreFile(c echo.Context) error {
	dbTx := server.ServiceManager.DbClient.Begin()
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	persistentVolumeIdStr := c.Param("id")
	// convert id to uint
	persistentVolumeId, err := strconv.Atoi(persistentVolumeIdStr)
	if err != nil {
		return c.JSON(400, map[string]string{
			"message": "Invalid id",
		})
	}
	// fetch persistent volume
	var persistentVolume core.PersistentVolume
	err = persistentVolume.FindById(c.Request().Context(), *dbTx, uint(persistentVolumeId))
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "Internal server error",
		})
	}
	// create a new persistent volume restore
	persistentVolumeRestore := core.PersistentVolumeRestore{
		Type:               core.LocalRestore,
		Status:             core.RestorePending,
		PersistentVolumeID: persistentVolume.ID,
		File:               "",
		CreatedAt:          time.Now(),
		CompletedAt:        time.Now(),
	}
	err = persistentVolumeRestore.Create(c.Request().Context(), *dbTx)
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "Internal server error",
		})
	}
	// open file
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
	filePath := filepath.Join(server.Config.LocalConfig.ServiceConfig.PVRestoreDirectoryPath, fileName)
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
	err = persistentVolumeRestore.Update(c.Request().Context(), *dbTx, server.Config.LocalConfig.ServiceConfig.PVBackupDirectoryPath)
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "failed to update restore",
		})
	}
	// commit
	err = dbTx.Commit().Error
	if err != nil {
		return c.JSON(500, map[string]string{
			"message": "failed to create restore",
		})
	}
	err = server.WorkerManager.EnqueuePersistentVolumeRestoreRequest(persistentVolumeRestore.ID)
	if err != nil {
		// mark restore as failed
		persistentVolumeRestore.Status = core.RestoreFailed
		err = persistentVolumeRestore.Update(c.Request().Context(), server.ServiceManager.DbClient, server.Config.LocalConfig.ServiceConfig.PVBackupDirectoryPath)
		if err != nil {
			log.Println(err)
		}
		return c.JSON(500, map[string]string{
			"message": "failed to enqueue restore job",
		})
	}
	return c.JSON(200, map[string]string{
		"message": "Restore job has been enqueued. You can check the status of the restore job in restore panel",
	})
}
