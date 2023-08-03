package server

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Init functions

func (server *Server) InitApplicationRestAPI() {
	server.ECHO_SERVER.POST("/application/deploy/upload", server.UploadTarFile)
	server.ECHO_SERVER.POST("/application/deploy/dockerconfig/generate/tarball", server.GenerateDockerConfigFromTarball)
	server.ECHO_SERVER.POST("/application/deploy/dockerconfig/generate/git", server.GenerateDockerConfigFromGit)
}


// Upload tar file and return the file name
// POST /application/deploy/upload
 func (server *Server) UploadTarFile(c echo.Context) error {
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
	destFilename := uuid.New().String()+".tar"
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
		"file": destFilename,
		"message": "file uploaded successfully",
	})
 }

// Dockerconfig generate from git repo
// POST /application/deploy/dockerconfig/generate/git
func (server *Server) GenerateDockerConfigFromGit(c echo.Context) error {
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
func (server *Server) GenerateDockerConfigFromTarball(c echo.Context) error {
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
// Deploy application
// POST /application/deploy
// Data :
// - git / tarball
// - dockerconfig
// - env variables
// - build args
// - image

// GET /applications
// GET /application/:id
// GET /application/:id/logs
// GET /application/:id/resources
// PUT /application/:id