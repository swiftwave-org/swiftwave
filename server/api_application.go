package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	GIT_MANAGER "github.com/swiftwave-org/swiftwave/git_manager"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Init functions

func (server *Server) InitApplicationRestAPI() {
	server.ECHO_SERVER.POST("/applications/deploy/upload", server.uploadTarFile)
	server.ECHO_SERVER.POST("/applications/deploy/dockerconfig/generate/tarball", server.generateDockerConfigFromTarball)
	server.ECHO_SERVER.POST("/applications/deploy/dockerconfig/generate/git", server.generateDockerConfigFromGit)
	server.ECHO_SERVER.POST("/applications/deploy/dockerconfig/generate/custom", server.generateDockerConfigFromCustomDockerfile)
	server.ECHO_SERVER.POST("/applications/deploy", server.deployApplication)
	server.ECHO_SERVER.GET("/applications", server.getApplications)
	server.ECHO_SERVER.GET("/applications/summary", server.getApplicationsSummary)
	server.ECHO_SERVER.GET("/applications/:id", server.getApplication)
	server.ECHO_SERVER.GET("/applications/:id/status", server.getApplicationStatus)
	server.ECHO_SERVER.PUT("/applications/:id", server.updateApplication)
	server.ECHO_SERVER.POST("/applications/:id/redeploy", server.redeployApplication)
	server.ECHO_SERVER.DELETE("/applications/:id", server.deleteApplication)
	server.ECHO_SERVER.GET("/applications/:id/logs/build", server.getApplicationBuildLogs)
	server.ECHO_SERVER.GET("/ws/applications/:id/logs/build/:log_id", server.getApplicationBuildLog)
	server.ECHO_SERVER.GET("/ws/applications/:id/logs/runtime", server.getApplicationRuntimeLogs)
	server.ECHO_SERVER.GET("/applications/availiblity/service_name", server.checkApplicationServiceNameAvailability)
	server.ECHO_SERVER.GET("/applications/servicenames", server.getApplicationServiceNames)
}

// Upload tar file and return the file name
// POST /application/deploy/upload
func (server *Server) uploadTarFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		log.Println(err)
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	src, err := file.Open()
	if err != nil {
		log.Println(err)
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	defer src.Close()

	// Check if file is tar
	if file.Header.Get("Content-Type") != "application/x-tar" {
		return c.JSON(400, map[string]string{
			"message": "file is not a tar file",
		})
	}

	// Destination
	destFilename := uuid.New().String() + ".tar"
	destFile := filepath.Join(server.CODE_TARBALL_DIR, destFilename)
	dst, err := os.Create(destFile)
	if err != nil {
		log.Println(err)
		return c.JSON(500, map[string]string{
			"message": "failed to create file",
		})
	}
	defer dst.Close()

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

// Dockerconfig generate from git repo
// POST /application/deploy/dockerconfig/generate/git
func (server *Server) generateDockerConfigFromGit(c echo.Context) error {
	repository_url := c.FormValue("repository_url")
	branch := c.FormValue("branch")
	git_credential_id := c.FormValue("git_credential_id")
	if repository_url == "" || branch == "" || git_credential_id == "" {
		return c.JSON(400, map[string]string{
			"message": "missing parameters",
		})
	}
	// Fetch git credential
	var gitCredential GitCredential
	if err := server.DB_CLIENT.Where("id = ?", git_credential_id).First(&gitCredential).Error; err != nil {
		log.Println(err)
		return c.JSON(400, map[string]string{
			"message": "git credential not found",
		})
	}
	// Generate config
	config, err := server.DOCKER_CONFIG_GENERATOR.GenerateConfigFromGitRepository(repository_url, branch, gitCredential.Username, gitCredential.Password)
	if err != nil {
		log.Println(err)
		return c.JSON(500, map[string]string{
			"message": "failed to generate docker config",
		})
	}
	// Return config
	return c.JSON(200, config)
}

// Dockerconfig generate from source code
// POST /application/deploy/dockerconfig/generate/tarball
func (server *Server) generateDockerConfigFromTarball(c echo.Context) error {
	fileName := c.FormValue("file")
	if fileName == "" {
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	fileName = SanitizeFileName(fileName)
	filePath := filepath.Join(server.CODE_TARBALL_DIR, fileName)
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.JSON(400, map[string]string{
			"message": "file not found",
		})
	}
	// Anaylze tarball
	config, err := server.DOCKER_CONFIG_GENERATOR.GenerateConfigFromSourceCodeTar(filePath)
	if err != nil {
		log.Println(err)
		return c.JSON(500, map[string]string{
			"message": "failed to generate docker config",
		})
	}
	// Return config
	return c.JSON(200, config)
}

// Dockerconfig generate from Custom Dockerfile
// POST /application/deploy/dockerconfig/generate/custom
func (server *Server) generateDockerConfigFromCustomDockerfile(c echo.Context) error {
	dockerfile := c.FormValue("dockerfile")
	dockerfile = strings.ReplaceAll(dockerfile, "\\n", "\n")
	if dockerfile == "" {
		return c.JSON(400, map[string]string{
			"message": "dockerfile not found",
		})
	}

	// Generate config
	config := server.DOCKER_CONFIG_GENERATOR.GenerateConfigFromCustomDocker(dockerfile)
	// Send
	return c.JSON(200, config)
}

// Deploy application
// POST /application/deploy
func (server *Server) deployApplication(c echo.Context) error {
	// Get data
	var deployRequest ApplicationDeployRequest
	err := c.Bind(&deployRequest)
	if err != nil {
		log.Println(err)
		return c.JSON(400, map[string]string{
			"message": "missing parameters",
		})
	}
	if (deployRequest.ApplicationSourceType == ApplicationSourceTypeGit && (deployRequest.RepositoryURL == "" ||
		deployRequest.Branch == "" ||
		deployRequest.GitCredentialID == 0)) ||
		(deployRequest.ApplicationSourceType == ApplicationSourceTypeTarball && deployRequest.TarballFile == "") {
		return c.JSON(400, map[string]string{
			"message": "missing parameters",
		})
	}
	var gitCredential GitCredential = GitCredential{}
	// Check if git credential exists
	if deployRequest.ApplicationSourceType == ApplicationSourceTypeGit {
		tx := server.DB_CLIENT.Where("id = ?", deployRequest.GitCredentialID).First(&gitCredential)
		if tx.Error != nil {
			log.Println(err)
			return c.JSON(400, map[string]string{
				"message": "git credential not found",
			})
		}
	}
	// Check if tarball exists
	if deployRequest.ApplicationSourceType == ApplicationSourceTypeTarball {
		filePath := filepath.Join(server.CODE_TARBALL_DIR, SanitizeFileName(deployRequest.TarballFile))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.JSON(400, map[string]string{
				"message": "tarball not found",
			})
		}
	}
	// Check if image exists
	if deployRequest.ApplicationSourceType == ApplicationSourceTypeImage {
		if deployRequest.DockerImage == "" {
			return c.JSON(400, map[string]string{
				"message": "image not found",
			})
		}
	}

	// Check if volume exists
	if deployRequest.Volumes == nil {
		deployRequest.Volumes = make(map[string]string, 0)
	}
	for volume_name := range deployRequest.Volumes {
		if !server.DOCKER_MANAGER.ExistsVolume(volume_name) {
			return c.JSON(400, map[string]string{
				"message": volume_name + " : volume not exists",
			})
		}
	}

	var application Application
	var applicationSource ApplicationSource

	err = server.DB_CLIENT.Transaction(func(tx *gorm.DB) error {
		// Create application source
		applicationSource = ApplicationSource{
			ID:                 0,
			Type:               deployRequest.ApplicationSourceType,
			GitCredential:      gitCredential,
			GitCredentialID:    gitCredential.ID,
			GitProvider:        FetchGitProviderFromURL(deployRequest.RepositoryURL),
			RepositoryUsername: FetchRepositoryUsernameFromURL(deployRequest.RepositoryURL),
			RepositoryName:     FetchRepositoryNameFromURL(deployRequest.RepositoryURL),
			Branch:             deployRequest.Branch,
			LastCommit:         "",
			TarballFile:        deployRequest.TarballFile,
			DockerImage:        deployRequest.DockerImage,
		}
		if err := tx.Create(&applicationSource).Error; err != nil {
			return err
		}
		if deployRequest.EnvironmentVariables == nil {
			deployRequest.EnvironmentVariables = make(map[string]string, 0)
		}
		envVariables, err := json.Marshal(deployRequest.EnvironmentVariables)
		if err != nil {
			return err
		}
		if deployRequest.BuildArgs == nil {
			deployRequest.BuildArgs = make(map[string]string, 0)
		}
		buildArgs, err := json.Marshal(deployRequest.BuildArgs)
		if err != nil {
			return err
		}
		volumes, err := json.Marshal(deployRequest.Volumes)
		if err != nil {
			return err
		}
		// Create application
		application = Application{
			ID:                   0,
			ServiceName:          deployRequest.ServiceName,
			Source:               applicationSource,
			SourceID:             applicationSource.ID,
			EnvironmentVariables: string(envVariables),
			BuildArgs:            string(buildArgs),
			Volumes:              string(volumes),
			Dockerfile:           deployRequest.Dockerfile,
			Image:                "",
			Status:               ApplicationStatusPending,
			Replicas:             deployRequest.Replicas,
		}
		if err := tx.Create(&application).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println(err)
		return c.JSON(500, map[string]string{
			"message": "failed to create application",
		})
	} else {
		// Send message to queue
		// server.QUEUE_CLIENT.Publish("application", "deploy", application.ID)
		return c.JSON(200, application)
	}
}

// GET /applications
func (server *Server) getApplications(c echo.Context) error {
	var applications []Application
	tx := server.DB_CLIENT.Preload("Source.GitCredential").Preload(clause.Associations).Find(&applications)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(500, map[string]string{
			"message": "failed to get applications",
		})
	}
	return c.JSON(200, applications)
}

// GET /applications/summary
func (server *Server) getApplicationsSummary(c echo.Context) error {
	var applications []Application
	tx := server.DB_CLIENT.Preload("Source.GitCredential").Preload(clause.Associations).Find(&applications)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(500, map[string]string{
			"message": "failed to get applications",
		})
	}
	var applicationSummaries []ApplicationSummary = make([]ApplicationSummary, 0)
	for _, application := range applications {
		applicationSummaries = append(applicationSummaries, ApplicationSummary{
			ID:          application.ID,
			ServiceName: application.ServiceName,
			Source:      application.Source.GetSourceSummary(),
			Replicas:    application.Replicas,
			Status:      application.Status,
		})
	}
	return c.JSON(200, applicationSummaries)
}

// GET /application/:id
func (server *Server) getApplication(c echo.Context) error {
	applicationID := c.Param("id")
	var application Application
	tx := server.DB_CLIENT.Preload("Source.GitCredential").Preload(clause.Associations).Where("id = ?", applicationID).First(&application)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(404, map[string]string{
			"message": "failed to get application",
		})
	}
	return c.JSON(200, application)
}

// DELETE /application/:id
func (server *Server) deleteApplication(c echo.Context) error {
	applicationID := c.Param("id")
	var application Application
	tx := server.DB_CLIENT.Where("id = ?", applicationID).First(&application)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(404, map[string]string{
			"message": "failed to get application",
		})
	}
	// Verify if there is no ingress rule
	var ingressRule IngressRule
	tx = server.DB_CLIENT.Where("service_name = ?", application.ServiceName).First(&ingressRule)
	if tx.Error == nil {
		return c.JSON(500, map[string]string{
			"message": "failed ! ingress rule exists. delete ingress rule first",
		})
	}
	// Delete application
	tx = server.DB_CLIENT.Delete(&application)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(500, map[string]string{
			"message": "failed to delete application",
		})
	}
	// Remove service
	err := server.DOCKER_MANAGER.RemoveService(application.ServiceName)
	if err != nil {
		log.Println(err)
		return c.JSON(500, map[string]string{
			"message": "failed to delete application",
		})
	}
	// Delete all logs
	server.DB_CLIENT.Where("application_id = ?", applicationID).Delete(&ApplicationBuildLog{})
	return c.JSON(200, map[string]string{
		"message": "application deleted",
	})
}

// GET /application/:id/logs/build
// Return record without `Logs` field
func (server *Server) getApplicationBuildLogs(c echo.Context) error {
	var applicationBuildLogs []ApplicationBuildLog
	applicationID := c.Param("id")
	tx := server.DB_CLIENT.Model(&ApplicationBuildLog{}).Select("id", "application_id", "time").Where("application_id = ?", applicationID).Find(&applicationBuildLogs)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(500, map[string]string{
			"message": "failed to get application logs",
		})
	}
	return c.JSON(200, applicationBuildLogs)
}

// GET /ws/application/:id/logs/build/:log_id
func (server *Server) getApplicationBuildLog(c echo.Context) error {
	var applicationBuildLog ApplicationBuildLog
	logID := c.Param("log_id")
	applicationID := c.Param("id")
	tx := server.DB_CLIENT.Model(&ApplicationBuildLog{}).Select("completed", "logs").Where(map[string]interface{}{
		"id":             logID,
		"application_id": applicationID,
	}).Find(&applicationBuildLog)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.NoContent(http.StatusNotFound)
	}
	// Upgrade connection to websocket
	ws, err := server.WEBSOCKET_UPGRADER.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// Close connection
	defer ws.Close()
	closed := false

	// Listen to redis subscriber
	ctx := context.Background()
	channel := "log_update/" + logID
	pubsub := server.REDIS_CLIENT.Subscribe(ctx, channel)
	defer pubsub.Close()

	if !applicationBuildLog.Completed {
		// listen for close message from client
		go func() {
			for {
				messageType, _, err := ws.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						c.Logger().Error(err)
					}
					closed = true
					return
				}
				if messageType == websocket.CloseMessage {
					// Peer has sent a close message, indicating they want to disconnect
					log.Println("Peer has sent a close message, indicating they want to disconnect")
					return
				}
			}
		}()
	}

	// send logs to client
	go func() {
		// Write initial logs
		err := ws.WriteMessage(websocket.TextMessage, []byte(applicationBuildLog.Logs))
		if err != nil {
			log.Println("failed to write initial logs to websocket")
		}
		if !applicationBuildLog.Completed {
			// Listen to redis channel and send logs to client
			for {
				if closed {
					return
				}
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					log.Println(err)
					return
				} else {
					if strings.Compare(msg.Payload, "SWIFTWAVE_EOF_LOG") == 0 {
						return
					}
					err := ws.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
					if err != nil {
						log.Println("failed to write logs to websocket")
					}
				}
			}
		}
	}()
	select {}
}

// GET /ws/application/:id/logs/runtime
func (server *Server) getApplicationRuntimeLogs(c echo.Context) error {
	applicationID := c.Param("id")
	var application Application
	tx := server.DB_CLIENT.Where("id = ?", applicationID).First(&application)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.NoContent(http.StatusNotFound)
	}
	// Get logs
	logsReader, err := server.DOCKER_MANAGER.LogsService(application.ServiceName)
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	scanner := bufio.NewScanner(logsReader)
	ws, err := server.WEBSOCKET_UPGRADER.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	// Close connection
	defer ws.Close()
	// Close logs reader
	defer logsReader.Close()

	closed := false

	// listen for close message from client
	go func() {
		for {
			messageType, _, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.Logger().Error(err)
				}
				closed = true
				return
			}
			if messageType == websocket.CloseMessage {
				// Peer has sent a close message, indicating they want to disconnect
				log.Println("Peer has sent a close message, indicating they want to disconnect")
				return
			}
		}
	}()

	// send logs to client
	go func() {
		// Write
		for scanner.Scan() {
			if closed {
				return
			}
			// Specific format for raw-stream logs
			// docs : https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerAttach
			log_text := scanner.Bytes()
			if len(log_text) > 8 {
				log_text = log_text[8:]
			}
			// add new line
			log_text = append(log_text, []byte("\n")...)
			err = ws.WriteMessage(websocket.TextMessage, log_text)
			if err != nil {
				log.Println("failed to write logs to websocket")
			}
		}
	}()
	select {}
}

// PUT /application/:id
func (server *Server) updateApplication(c echo.Context) error {
	var updateRequest ApplicationDeployUpdateRequest
	if err := c.Bind(&updateRequest); err != nil {
		log.Println(err)
		return c.JSON(400, map[string]string{
			"message": "invalid request body",
		})
	}
	// Fetch application record
	applicationID := c.Param("id")
	var application Application
	tx := server.DB_CLIENT.Preload("Source.GitCredential").Preload(clause.Associations).Where("id = ?", applicationID).First(&application)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(404, map[string]string{
			"message": "failed to get application",
		})
	}
	var source ApplicationSource = application.Source
	// Update application
	if updateRequest.Source.Type != application.Source.Type {
		log.Println("Source type cannot be changed")
		return c.JSON(400, map[string]string{
			"message": "source type cannot be changed",
		})
	}
	// Track whether image configuration is changed
	var imageChanged bool = false
	// Verify condition for Type `git`
	if updateRequest.Source.Type == ApplicationSourceTypeGit {
		if updateRequest.Source.GitCredentialID == 0 {
			log.Println("Git credential is required")
			return c.JSON(400, map[string]string{
				"message": "git credential is required",
			})
		}
		source.GitCredentialID = updateRequest.Source.GitCredentialID
		if updateRequest.Source.RepositoryName == "" {
			log.Println("Repository name is required")
			return c.JSON(400, map[string]string{
				"message": "repository name is required",
			})
		}
		source.RepositoryName = updateRequest.Source.RepositoryName
		if updateRequest.Source.Branch == "" {
			log.Println("Branch is required")
			return c.JSON(400, map[string]string{
				"message": "branch is required",
			})
		}
		source.Branch = updateRequest.Source.Branch
		commithash, err := GIT_MANAGER.FetchLatestCommitHash(source.RepositoryURL(), source.Branch, source.GitCredential.Username, source.GitCredential.Password)
		if err == nil {
			imageChanged = imageChanged || (commithash != application.Source.LastCommit)
		} else {
			imageChanged = true
		}
	}
	// Verify condition for Type `tarball`
	if updateRequest.Source.Type == ApplicationSourceTypeTarball {
		if updateRequest.Source.TarballFile == "" {
			log.Println("Tarball URL is required")
			return c.JSON(400, map[string]string{
				"message": "tarball URL is required",
			})
		}
		imageChanged = imageChanged || (updateRequest.Source.TarballFile != source.TarballFile)
		source.TarballFile = updateRequest.Source.TarballFile
	}

	// Verify condition for Type `image`
	if updateRequest.Source.Type == ApplicationSourceTypeImage {
		if updateRequest.Source.DockerImage == "" {
			log.Println("Image is required")
			return c.JSON(400, map[string]string{
				"message": "image is required",
			})
		}
		imageChanged = imageChanged || (updateRequest.Source.DockerImage != source.DockerImage)
		source.DockerImage = updateRequest.Source.DockerImage
	}

	err := server.DB_CLIENT.Transaction(func(tx *gorm.DB) error {
		tx2 := tx.Save(&source)
		if tx2.Error != nil {
			log.Println(tx2.Error)
			return tx2.Error
		}
		application.SourceID = source.ID
		application.Replicas = updateRequest.Replicas
		application.Dockerfile = updateRequest.Dockerfile
		application.Replicas = updateRequest.Replicas
		environment_variables, err := json.Marshal(updateRequest.EnvironmentVariables)
		if err != nil {
			log.Println(err)
			return err
		}
		application.EnvironmentVariables = string(environment_variables)
		build_args, err := json.Marshal(updateRequest.BuildArgs)
		if err != nil {
			log.Println(err)
			return err
		}
		if string(build_args) != application.BuildArgs {
			imageChanged = true
		}
		application.BuildArgs = string(build_args)
		// Reset status
		if imageChanged {
			application.Status = ApplicationStatusRedeployPending
		} else {
			application.Status = ApplicationStatusDeployingPending
		}
		// Update application
		tx3 := tx.Save(&application)
		if tx3.Error != nil {
			log.Println(tx3.Error)
			return tx3.Error
		}
		return nil
	})

	if err != nil {
		log.Println(err)
		return c.JSON(500, map[string]string{
			"message": "failed to update application",
		})
	}
	return c.JSON(200, application)
}

// GET /application/:id/redeploy
func (server *Server) redeployApplication(c echo.Context) error {
	applicationID := c.Param("id")
	var application Application
	tx := server.DB_CLIENT.Preload("Source.GitCredential").Preload(clause.Associations).Where("id = ?", applicationID).First(&application)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(404, map[string]string{
			"message": "failed to get application",
		})
	}
	// Update application
	application.Status = ApplicationStatusRedeployPending
	tx2 := server.DB_CLIENT.Save(&application)
	if tx2.Error != nil {
		log.Println(tx2.Error)
		return c.JSON(500, map[string]string{
			"message": "failed to update application",
		})
	}
	return c.JSON(200, application)
}

// GET /application/:id/status
func (server *Server) getApplicationStatus(c echo.Context) error {
	applicationID := c.Param("id")
	var application Application
	tx := server.DB_CLIENT.Where("id = ?", applicationID).First(&application)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(404, map[string]string{
			"message": "failed to get application",
		})
	}
	return c.JSON(200, application.Status)
}

// GET /application/availiblity/service_name/?name=xxxx
func (server *Server) checkApplicationServiceNameAvailability(c echo.Context) error {
	name := c.QueryParam("name")
	isAvailable := true
	var application Application
	if name != "" {
		// Check from database
		tx := server.DB_CLIENT.Where("service_name = ?", name).First(&application)
		if tx.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				isAvailable = isAvailable && true
			} else {
				isAvailable = isAvailable && false
			}
		} else {
			isAvailable = isAvailable && false
		}
		// Check from docker
		_, err := server.DOCKER_MANAGER.StatusService(name)
		isAvailable = isAvailable && (err != nil)
	} else {
		isAvailable = false
	}
	return c.JSON(200, map[string]bool{
		"available": isAvailable,
	})
}

// GET /applications/servicenames
func (server *Server) getApplicationServiceNames(c echo.Context) error {
	var applications []Application
	tx := server.DB_CLIENT.Select("service_name").Find(&applications)
	if tx.Error != nil {
		log.Println(tx.Error)
		return c.JSON(500, map[string]string{
			"message": "failed to get application",
		})
	}
	var serviceNames []string = []string{}
	for _, application := range applications {
		serviceNames = append(serviceNames, application.ServiceName)
	}
	return c.JSON(200, serviceNames)
}
