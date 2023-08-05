package server

import (
	"log"
	"net/http"
	"strconv"
	GIT "keroku/m/git_manager"
	"github.com/labstack/echo/v4"
)

// Init functions

func (server *Server) InitGitRestAPI() {
	server.ECHO_SERVER.GET("/git/credentials", server.getGitCredentials)
	server.ECHO_SERVER.GET("/git/credentials/:id", server.getGitCredential)
	server.ECHO_SERVER.POST("/git/credentials", server.createGitCredential)
	server.ECHO_SERVER.PUT("/git/credentials/:id", server.updateGitCredential)
	server.ECHO_SERVER.DELETE("/git/credentials/:id", server.deleteGitCredential)
	server.ECHO_SERVER.GET("/git/credentials/:id/test", server.testGitCredential)
}

// GET /git/credentials
func (server *Server) getGitCredentials(c echo.Context) error {
	// Fetch all git credentials from database
	var gitCredentials []GitCredential
	tx := server.DB_CLIENT.Find(&gitCredentials)
	if tx.Error != nil {
		log.Println(tx.Error.Error())
		return c.JSON(500, map[string]interface{}{
			"message": "Failed to fetch git credentials from database",
		})
	}
	return c.JSON(200, gitCredentials)
}

// GET /git/credentials/:id
func (server *Server) getGitCredential(c echo.Context) error {
	if c.Param("id") == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
	}
	var gitCredential GitCredential
	tx := server.DB_CLIENT.First(&gitCredential, c.Param("id"))
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"message": "git credential not found",
		})
	}
	return c.JSON(200, gitCredential)
}

// POST /git/credentials
func (server *Server) createGitCredential(c echo.Context) error {
	// JSON decode request body
	var gitCredential GitCredential
	tx := c.Bind(&gitCredential)
	if tx != nil {
		return c.JSON(400, map[string]interface{}{
			"message": "Failed to decode request body",
		})
	}
	if gitCredential.Username == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "username is required",
		})
	}
	if gitCredential.Name == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "name is required",
		})
	}
	// Insert git credential into database
	tx2 := server.DB_CLIENT.Create(&gitCredential)
	if tx2.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"message": "Failed to insert git credential into database",
		})
	}
	return c.JSON(200, gitCredential)
}

// PUT /git/credentials/:id
func (server *Server) updateGitCredential(c echo.Context) error {
	// JSON decode request body
	var gitCredential GitCredential
	tx := c.Bind(&gitCredential)
	if tx != nil {
		return c.JSON(400, map[string]interface{}{
			"message": "Failed to decode request body",
		})
	}
	if gitCredential.Username == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "username is required",
		})
	}
	if gitCredential.Password == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "password is required",
		})
	}
	if gitCredential.Name == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "name is required",
		})
	}
	id , err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "id parameter is required",
		})
	}
	gitCredential.ID = uint(id)

	// Update git credential in database
	tx2 := server.DB_CLIENT.Save(&gitCredential)
	if tx2.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"message": "Failed to update git credential in database",
		})
	}
	return c.JSON(200, gitCredential)
}

// DELETE /git/credentials/:id
func (server *Server) deleteGitCredential(c echo.Context) error {
	if c.Param("id") == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
	}
	
	// check if git credential is used by any application
	var applicationSource ApplicationSource
	tx2 := server.DB_CLIENT.Where("git_credential_id = ?", c.Param("id")).First(&applicationSource)
	if tx2.Error == nil {
		return c.JSON(400, map[string]interface{}{
			"message": "git credential is used by an application",
		})
	}

	// Delete git credential from database
	tx := server.DB_CLIENT.Delete(&GitCredential{}, c.Param("id"))
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"message": "Failed to delete git credential from database",
		})
	}
	return c.JSON(200, map[string]interface{}{
		"message": "Git credential deleted successfully",
	})
}

// GET /git/credentials/:id/test
func (server *Server) testGitCredential(c echo.Context) error {
	if c.Param("id") == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
	}
	var gitCredential GitCredential
	tx := server.DB_CLIENT.First(&gitCredential, c.Param("id"))
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"message": "git credential not found",
		})
	}
	repositoryUrl := c.QueryParam("repository_url")
	if repositoryUrl == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "repository_url query parameter is required",
		})
	}
	branch := c.QueryParam("branch")
	if branch == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "branch query parameter is required",
		})
	}
	// Test git credential
	hash, err := GIT.FetchLatestCommitHash(repositoryUrl, branch, gitCredential.Username, gitCredential.Password)

	if err != nil {
		log.Println(err.Error())
		return c.JSON(500, map[string]interface{}{
			"message": "failed to access repository",
		})
	}

	return c.JSON(200, map[string]interface{}{
		"message": "Git credential is valid and repository is accessible",
		"hash": hash,
	})
}
