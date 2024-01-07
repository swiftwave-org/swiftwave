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
	err := m.buildApplicationHelper(request)
	if err != nil {
		addDeploymentLog(m.ServiceManager.DbClient, m.ServiceManager.PubSubClient, request.DeploymentId, "Failed to build application\n"+err.Error()+"\n", true)
		// update status
		deployment := &core.Deployment{}
		deployment.ID = request.DeploymentId
		err = deployment.UpdateStatus(context.Background(), m.ServiceManager.DbClient, core.DeploymentStatusFailed)
		if err != nil {
			log.Println("failed to update deployment status. Error: ", err)
		}
	}
	// If it fails, don't requeue the job
	return nil
}

// private functions
func (m Manager) buildApplicationHelper(request BuildApplicationRequest) error {
	// database client to work without transaction
	dbWithoutTx := m.ServiceManager.DbClient
	// pubSub client
	pubSubClient := m.ServiceManager.PubSubClient
	// start a database transaction
	db := m.ServiceManager.DbClient.Begin()
	defer func() {
		db.Rollback()
	}()
	ctx := context.Background()
	// find out the deployment
	deployment := &core.Deployment{}
	err := deployment.FindById(ctx, *db, request.DeploymentId)
	if err != nil {
		return err
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

func (m Manager) buildApplicationForDockerImage(deployment *core.Deployment, ctx context.Context, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client) error {
	// TODO: add support for registry authentication
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "As the upstream type is image, no build is required\n", false)
	err := deployment.UpdateStatus(ctx, db, core.DeploymentStatusDeployPending)
	if err != nil {
		return err
	}
	// commit the transaction
	err = db.Commit().Error
	if err != nil {
		return err
	}

	// push task to queue for deployment
	err = m.ServiceManager.TaskQueueClient.EnqueueTask("deploy_application", DeployApplicationRequest{
		AppId: deployment.ApplicationID,
		DeploymentId: deployment.ID,
	})
	if err == nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment has been triggered. Waiting for deployment to complete\n", false)
	}
	return err
}

func (m Manager) buildApplicationForGit(deployment *core.Deployment, ctx context.Context, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client) error {
	gitUsername := ""
	gitPassword := ""

	if deployment.GitCredentialID != nil {
		// fetch git credentials
		gitCredentials := &core.GitCredential{}
		err := gitCredentials.FindById(ctx, db, *deployment.GitCredentialID)
		if err != nil {
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch git credentials\n", true)
			return err
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
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch latest commit hash\n", true)
		return err
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Fetched latest commit hash > "+commitHash+"\n", false)
	deployment.CommitHash = commitHash
	// update deployment
	err = db.Model(&deployment).Update("commit_hash", commitHash).Error
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to update commit hash\n", true)
		return err
	}
	// clone git repository
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Cloning git repository > "+deployment.GitRepositoryURL()+"\n", false)
	err = gitmanager.CloneRepository(deployment.GitRepositoryURL(), deployment.RepositoryBranch, gitUsername, gitPassword, tempDirectory)
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to clone git repository\n", true)
		return err
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Cloned git repository successfully\n", false)
	// build docker image
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Started building docker image\n", false)
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
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to build docker image\n", true)
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
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build failed\n", true)
		return err
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image built successfully\n", false)
	// update status
	err = deployment.UpdateStatus(ctx, db, core.DeploymentStatusDeployPending)
	if err != nil {
		return err
	}
	// commit the transaction
	err = db.Commit().Error
	if err != nil {
		return err
	}
	// push task to queue for deployment
	err = m.ServiceManager.TaskQueueClient.EnqueueTask("deploy_application", DeployApplicationRequest{
		AppId:        deployment.ApplicationID,
		DeploymentId: deployment.ID,
	})
	if err == nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment has been triggered. Waiting for deployment to complete\n", false)
	}
	return err
}

func (m Manager) buildApplicationForTarball(deployment *core.Deployment, ctx context.Context, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client) error {
	tarballPath := filepath.Join(m.SystemConfig.ServiceConfig.DataDir, deployment.SourceCodeCompressedFileName)
	// Verify file exists
	if _, err := os.Stat(tarballPath); os.IsNotExist(err) {
		return errors.New("tarball file not found")
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
			log.Println("failed to remove temporary directory", err)
		}
	}(tempDirectory)
	// extract tarball
	err = dockerconfiggenerator.ExtractTar(tarballPath, tempDirectory)
	if err != nil {
		return errors.New("failed to extract tarball")
	}
	// build docker image
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Started building docker image\n", false)
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
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to build docker image\n", true)
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
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build failed\n", true)
		return errors.New("unexpected failure at the time of building docker image")
	}
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image built successfully\n", false)
	// update status
	err = deployment.UpdateStatus(ctx, db, core.DeploymentStatusDeployPending)
	if err != nil {
		return err
	}
	// commit the transaction
	err = db.Commit().Error
	if err != nil {
		return err
	}
	// push task to queue for deployment
	err = m.ServiceManager.TaskQueueClient.EnqueueTask("deploy_application", DeployApplicationRequest{
		AppId: deployment.ApplicationID,
		DeploymentId: deployment.ID,
	})
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment has been triggered. Waiting for deployment to complete\n", false)
	}
	return err
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
