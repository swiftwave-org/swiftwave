package server

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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
	server.ECHO_SERVER.GET("/applications/:id", server.getApplication)
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
	// Check if git credential exists
	var gitCredential GitCredential
	tx := server.DB_CLIENT.Where("id = ?", deployRequest.GitCredentialID).First(&gitCredential)
	if tx.Error != nil {
		log.Println(err)
		return c.JSON(400, map[string]string{
			"message": "git credential not found",
		})
	}
	// Check if tarball exists
	if deployRequest.ApplicationSourceType == ApplicationSourceTypeTarball {
		filePath := filepath.Join(server.CODE_TARBALL_DIR, deployRequest.TarballFile)
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
		envVariables, err := json.Marshal(deployRequest.EnvironmentVariables)
		if err != nil {
			return err
		}
		buildArgs, err := json.Marshal(deployRequest.BuildArgs)
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
			Dockerfile:           deployRequest.Dockerfile,
			Image:                "",
			Status:               ApplicationStatusPending,
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

// GET /application/:id/logs
// GET /application/:id/resources
// PUT /application/:id
