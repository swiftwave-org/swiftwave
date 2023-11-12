package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	dockerconfiggenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	gitmanager "github.com/swiftwave-org/swiftwave/git_manager"
	"github.com/swiftwave-org/swiftwave/pubsub"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

func (m Manager) BuildApplication(request BuildApplicationRequest) error {
	// database client to work without transaction
	dbWithoutTx := m.ServiceManager.DbClient
	// pubSub client
	pubSubClient := m.ServiceManager.PubSubClient
	// start a database transaction
	db := m.ServiceManager.DbClient.Begin()
	ctx := context.Background()
	// find out the deployment
	deployment := &core.Deployment{}
	err := deployment.FindById(ctx, *db, request.DeploymentId)
	if err != nil {
		// check if error due to record not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// just return nil as we don't want to requeue the job
			return nil
		}
		// update it as failed
		err := deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		if err != nil {
			return err
		}
		db.Rollback()
		// retuning nil as we don't want to requeue the job
		return nil
	}
	// ensure deployment is in pending state
	if deployment.Status != core.DeploymentStatusPending {
		db.Rollback()
		// retuning nil as we don't want to requeue the job
		return nil
	}
	// #####  FOR IMAGE  ######
	// build for docker image
	if deployment.UpstreamType == core.UpstreamTypeImage {
		return m.buildApplicationForDockerImage(deployment, ctx, *db, dbWithoutTx, pubSubClient)
	}
	// #####  FOR GIT  ######
	if deployment.UpstreamType == core.UpstreamTypeGit {
		return m.buildApplicationForGit(deployment, ctx, *db, dbWithoutTx, pubSubClient)
	}
	// #####  FOR SOURCE CODE TARBALL  ######
	if deployment.UpstreamType == core.UpstreamTypeSourceCode {
		return m.buildApplicationForTarball(deployment, ctx, *db, dbWithoutTx, pubSubClient)
	}
	return nil
}

// private functions
func (m Manager) buildApplicationForDockerImage(deployment *core.Deployment, ctx context.Context, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client) error {
	// TODO: add support for registry authentication
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "As the upstream type is image, no build is required", false)
	err := deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusDeployPending)
	if err != nil {
		log.Println("failed to update deployment status")
		log.Println(err)
		return err
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment has been triggered. Waiting for deployment to complete", false)
	// push task to queue for deployment
	err = m.ServiceManager.TaskQueueClient.EnqueueTask("deploy_application", DeployApplicationRequest{
		AppId: deployment.ApplicationID,
	})
	if err != nil {
		return err
	}
	// commit the transaction
	return db.Commit().Error
}

func (m Manager) buildApplicationForGit(deployment *core.Deployment, ctx context.Context, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client) error {
	gitUsername := ""
	gitPassword := ""

	if deployment.GitCredentialID != nil {
		// fetch git credentials
		gitCredentials := &core.GitCredential{}
		err := gitCredentials.FindById(ctx, db, *deployment.GitCredentialID)
		if err != nil {
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch git credentials", true)
			err := deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
			if err != nil {
				return err
			}
			return nil
		}
		gitUsername = gitCredentials.Username
		gitPassword = gitCredentials.Password
	}
	// create temporary directory for git clone
	tempDirectory := "/tmp/" + uuid.New().String()
	err := os.Mkdir(tempDirectory, 0777)
	if err != nil {
		return err
	}
	// defer removing the temporary directory
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Println("Failed to remove temporary directory", err)
		}
	}(tempDirectory)
	// fetch commit hash
	commitHash, err := gitmanager.FetchLatestCommitHash(deployment.GitRepositoryURL(), deployment.RepositoryBranch, gitUsername, gitPassword)
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch latest commit hash", true)
		err := deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		if err != nil {
			return err
		}
		// retuning nil as we don't want to requeue the job because it may fail for same reason
		return nil
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Fetched latest commit hash > "+commitHash, false)
	deployment.CommitHash = commitHash
	// clone git repository
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Cloning git repository > "+deployment.GitRepositoryURL(), false)
	err = gitmanager.CloneRepository(deployment.GitRepositoryURL(), deployment.RepositoryBranch, gitUsername, gitPassword, tempDirectory)
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to clone git repository", true)
		err := deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		if err != nil {
			return err
		}
		// because if commit hash fetched successfully, then it is not possible to fail here for wrong credentials
		return err
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Cloned git repository successfully", false)
	// build docker image
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Started building docker image", false)
	// fetch build args
	var buildArgs []*core.BuildArg
	err = db.Where("deployment_id = ?", deployment.ID).Find(&buildArgs).Error
	if err != nil {
		return err
	}
	var buildArgsMap = make(map[string]string)
	for _, buildArg := range buildArgs {
		buildArgsMap[buildArg.Key] = buildArg.Value
	}

	// start building docker image
	scanner, err := m.ServiceManager.DockerManager.CreateImage(deployment.Dockerfile, buildArgsMap, tempDirectory, deployment.DeployableDockerImageURI())
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to build docker image", true)
		return err
	}
	isErrorEncountered := false
	if scanner != nil {
		var data map[string]interface{}
		for scanner.Scan() {
			err = json.Unmarshal(scanner.Bytes(), &data)
			if err != nil {
				continue
			}
			if data["stream"] != nil {
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["stream"].(string), false)
			}
			if data["error"] != nil {
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["error"].(string), false)
				isErrorEncountered = true
				break
			}
		}
	}
	if isErrorEncountered {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build failed", true)
		// update status
		err = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		if err != nil {
			return err
		}
		// retuning nil as we don't want to requeue the job because it may fail for same reason
		return nil
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image built successfully", false)
	// update status
	err = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusDeployPending)
	if err != nil {
		return err
	}
	// commit the transaction
	err = db.Commit().Error
	if err != nil {
		_ = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		return nil
	}
	// push task to queue for deployment
	err = m.ServiceManager.TaskQueueClient.EnqueueTask("deploy_application", DeployApplicationRequest{
		AppId: deployment.ApplicationID,
	})
	if err != nil {
		// set status to failed
		_ = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		return nil
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment has been triggered. Waiting for deployment to complete", false)
	return nil
}

func (m Manager) buildApplicationForTarball(deployment *core.Deployment, ctx context.Context, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client) error {
	tarballPath := filepath.Join(m.SystemConfig.ServiceConfig.DataDir, deployment.SourceCodeCompressedFileName)
	// Verify file exists
	if _, err := os.Stat(tarballPath); os.IsNotExist(err) {
		// mark as failed
		err = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		if err != nil {
			return err
		}
		// if file not exists, then return nil as we don't want to requeue the job
		return nil
	}
	// create temporary directory for extracting tarball
	tempDirectory := "/tmp/" + uuid.New().String()
	err := os.Mkdir(tempDirectory, 0777)
	if err != nil {
		return err
	}
	// defer removing the temporary directory
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Println("Failed to remove temporary directory", err)
		}
	}(tempDirectory)
	// extract tarball
	err = dockerconfiggenerator.ExtractTar(tarballPath, tempDirectory)
	if err != nil {
		// mark as failed, as we don't want to requeue the job
		return nil
	}
	// build docker image
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Started building docker image", false)
	// fetch build args
	var buildArgs []*core.BuildArg
	err = db.Where("deployment_id = ?", deployment.ID).Find(&buildArgs).Error
	if err != nil {
		return err
	}
	var buildArgsMap = make(map[string]string)
	for _, buildArg := range buildArgs {
		buildArgsMap[buildArg.Key] = buildArg.Value
	}

	// start building docker image
	scanner, err := m.ServiceManager.DockerManager.CreateImage(deployment.Dockerfile, buildArgsMap, tempDirectory, deployment.DeployableDockerImageURI())
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to build docker image", true)
		return err
	}
	isErrorEncountered := false
	if scanner != nil {
		var data map[string]interface{}
		for scanner.Scan() {
			err = json.Unmarshal(scanner.Bytes(), &data)
			if err != nil {
				continue
			}
			if data["stream"] != nil {
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["stream"].(string), false)
			}
			if data["error"] != nil {
				println(data["error"].(string))
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["error"].(string), false)
				isErrorEncountered = true
				break
			}
		}
	}
	if isErrorEncountered {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build failed", true)
		// update status
		err = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		if err != nil {
			return err
		}
		// retuning nil as we don't want to requeue the job because it may fail for same reason
		return nil
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image built successfully", false)
	// update status
	err = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusDeployPending)
	if err != nil {
		return err
	}
	// commit the transaction
	err = db.Commit().Error
	if err != nil {
		_ = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		return nil
	}
	// push task to queue for deployment
	err = m.ServiceManager.TaskQueueClient.EnqueueTask("deploy_application", DeployApplicationRequest{
		AppId: deployment.ApplicationID,
	})
	if err != nil {
		// set status to failed
		_ = deployment.UpdateStatus(ctx, dbWithoutTx, core.DeploymentStatusFailed)
		return nil
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment has been triggered. Waiting for deployment to complete", false)
	return nil
}

func addDeploymentLog(dbClient gorm.DB, pubSubClient pubsub.Client, deploymentId string, content string, terminate bool) {
	deploymentLog := &core.DeploymentLog{
		DeploymentID: deploymentId,
		Content:      content,
	}
	err := dbClient.Create(deploymentLog).Error
	if err != nil {
		log.Println("failed to add deployment log")
	}
	err = pubSubClient.Publish(fmt.Sprintf("deployment-log-%s", deploymentId), content)
	if err != nil {
		log.Println("failed to publish deployment log")
	}
	if terminate {
		err := pubSubClient.RemoveTopic(fmt.Sprintf("deployment-log-%s", deploymentId))
		if err != nil {
			log.Println("failed to remove topic")
		}
	}
}
