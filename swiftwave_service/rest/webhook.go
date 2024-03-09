package rest

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"gorm.io/gorm"
	"io"
	"net/url"
	"strings"
)

// ANY /webhook/redeploy-app/:app-id/:webhook-token
func (server *Server) redeployApp(c echo.Context) error {
	appId := c.Param("app-id")
	webhookToken := c.Param("webhook-token")
	if appId == "" || webhookToken == "" {
		return c.String(400, "Invalid request")
	}
	ctx := context.Background()
	// Fetch App
	application := core.Application{
		ID: appId,
	}
	err := application.FindById(ctx, server.ServiceManager.DbClient, application.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.String(404, "App not found")
		}
		return c.String(500, "Error fetching app")
	}
	// Check if webhook token matches
	if application.WebhookToken != webhookToken {
		return c.String(401, "Unauthorized")
	}
	// Fetch latest deployment
	deployment, err := core.FindCurrentLiveDeploymentByApplicationId(ctx, server.ServiceManager.DbClient, application.ID)
	if err != nil {
		_, err = core.FindLatestDeploymentByApplicationId(ctx, server.ServiceManager.DbClient, application.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.String(404, "No deployment found")
		}
		return c.String(500, "Error fetching deployment")
	}

	// Get body from request
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(500, "Error reading request body")
	}
	bodyString := string(body)
	// url decode body
	bodyString, err = url.QueryUnescape(bodyString)
	if err != nil {
		logger.HTTPLoggerError.Println(err)
	}

	triggeredRebuild := false
	// Check if latest deployment is git
	if deployment.UpstreamType == core.UpstreamTypeGit {
		searchText := "refs/heads/" + deployment.RepositoryBranch
		repoName := deployment.RepositoryOwner + "/" + deployment.RepositoryName
		if strings.Contains(bodyString, searchText) && strings.Contains(bodyString, repoName) {
			triggeredRebuild = true
		} else {
			return c.String(200, "OK - No rebuild")
		}
	}

	// Check if latest deployment is image
	if deployment.UpstreamType == core.UpstreamTypeImage {
		// remove the tag from the image name
		splits := strings.Split(deployment.DockerImage, ":")
		if len(splits) == 0 {
			return c.String(500, "Error parsing docker image name")
		}
		imageName := splits[0]
		// remove the registry from the image name
		splits = strings.Split(imageName, "/")
		if len(splits) == 0 {
			return c.String(500, "Error parsing docker image name")
		}
		if len(splits) == 1 {
			imageName = splits[0]
		} else {
			imageName = splits[len(splits)-2] + "/" + splits[len(splits)-1]
		}
		if strings.Contains(bodyString, imageName) {
			triggeredRebuild = true
		} else {
			return c.String(200, "OK - No rebuild")
		}
	}

	if triggeredRebuild {
		// fetch record
		var record = &core.Application{
			ID: application.ID,
		}
		tx := server.ServiceManager.DbClient.Begin()
		deploymentId, err := record.RebuildApplication(ctx, *tx)
		if err != nil {
			tx.Rollback()
			return errors.New("failed to create new deployment")
		}
		// commit transaction
		err = tx.Commit().Error
		if err != nil {
			tx.Rollback()
			return errors.New("failed to create new deployment due to database error")
		}
		// enqueue build request
		err = server.WorkerManager.EnqueueBuildApplicationRequest(record.ID, deploymentId)
		if err != nil {
			return errors.New("failed to queue build request")
		}
		return c.String(200, "OK - Rebuild triggered")
	}

	return c.String(200, "OK - No rebuild")
}
