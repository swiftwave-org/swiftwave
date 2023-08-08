package server

import (
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

func (server *Server) InitPersistentVolumeAPI() {
	server.ECHO_SERVER.GET("/volumes", server.GetVolumes)
	server.ECHO_SERVER.POST("/volumes", server.CreateVolume)
	server.ECHO_SERVER.DELETE("/volumes", server.DeleteVolume)
}

// GET /volumes
func (server *Server) GetVolumes(c echo.Context) error {
	volumes , err := server.DOCKER_MANAGER.FetchVolumes()
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"message": "error fetching volumes. docker daemon not responding",
		})
	}
	return c.JSON(200, volumes)
}

// POST /volumes
func (server *Server) CreateVolume(c echo.Context) error {
	volume_name := c.FormValue("name")
	volume_name = sanitizeVolumeName(volume_name)
	if volume_name == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "name query parameter is required",
		})
	}
	isExists := server.DOCKER_MANAGER.ExistsVolume(volume_name)
	if isExists {
		return c.JSON(500, map[string]interface{}{
			"message": "volume already exists",
		})
	}
	err := server.DOCKER_MANAGER.CreateVolume(volume_name)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"message": err.Error(),
		})
	}
	return c.JSON(200, map[string]interface{}{
		"message": volume_name + " : volume created successfully",
	})
}

// DELETE /volumes/:name
func (server *Server) DeleteVolume(c echo.Context) error {
	volume_name := c.Param("name")
	if volume_name == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "volume name is required",
		})
	}
	isExists := server.DOCKER_MANAGER.ExistsVolume(volume_name)
	if !isExists {
		return c.JSON(500, map[string]interface{}{
			"message": "volume does not exist",
		})
	}
	err := server.DOCKER_MANAGER.RemoveVolume(volume_name)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"message": err.Error(),
		})
	}
	return c.JSON(200, map[string]interface{}{
		"message": volume_name + " : volume deleted successfully",
	})
}

// Private functions
func sanitizeVolumeName(input string) string {
	// Define a regular expression pattern to match alphanumeric characters, underscore, and hash '#'
	pattern := "[a-zA-Z0-9_]+"
	re := regexp.MustCompile(pattern)

	// Find all substrings that match the pattern
	matches := re.FindAllString(input, -1)

	// Join the matched substrings to form the filtered string
	filtered := strings.Join(matches, "")

	// Convert the filtered string to lowercase
	filtered = strings.ToLower(filtered)

	return filtered
}