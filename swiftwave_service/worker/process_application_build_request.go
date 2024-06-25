package worker

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	dockerconfiggenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	gitmanager "github.com/swiftwave-org/swiftwave/git_manager"
	"github.com/swiftwave-org/swiftwave/pubsub"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

func (m Manager) BuildApplication(request BuildApplicationRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	// fetch docker manager
	dockerManager, err := containermanger.NewLocalClient(ctx)
	if err != nil {
		return err
	}
	isFailed, err := core.IsDeploymentFailed(context.Background(), m.ServiceManager.DbClient, request.DeploymentId)
	if err == nil {
		if isFailed {
			return nil
		}
	}
	subscriptionId, subscriptionChannel, _ := m.ServiceManager.PubSubClient.Subscribe(m.ServiceManager.CancelImageBuildTopic)
	defer func(id string) {
		err := m.ServiceManager.PubSubClient.Unsubscribe(m.ServiceManager.CancelImageBuildTopic, id)
		if err != nil {
			log.Println("failed to unsubscribe from topic > " + m.ServiceManager.CancelImageBuildTopic)
		}
	}(subscriptionId)

	isHelperExited := make(chan bool, 1)
	go func(deploymentId string, cancelContext context.CancelFunc) {
		for {
			select {
			case <-isHelperExited:
				return
			case id := <-subscriptionChannel:
				if id == deploymentId {
					cancelContext()
					return
				}
			}
		}
	}(request.DeploymentId, cancelContext)

	err = m.buildApplicationHelper(request, ctx, cancelContext, dockerManager)
	isHelperExited <- true
	if err != nil {
		addPersistentDeploymentLog(m.ServiceManager.DbClient, m.ServiceManager.PubSubClient, request.DeploymentId, "Failed to build application\n"+err.Error()+"\n", true)
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
func (m Manager) buildApplicationHelper(request BuildApplicationRequest, ctx context.Context, cancelContext context.CancelFunc, dockerManager *containermanger.Manager) error {
	// database client to work without transaction
	dbWithoutTx := m.ServiceManager.DbClient
	// pubSub client
	pubSubClient := m.ServiceManager.PubSubClient
	// start a database transaction
	db := m.ServiceManager.DbClient.Begin()
	defer func() {
		db.Rollback()
	}()
	// find out the deployment
	deployment := &core.Deployment{}
	err := deployment.FindById(ctx, *db, request.DeploymentId)
	if err != nil {
		return err
	}
	// #####  FOR IMAGE  ######
	// build for docker image
	if deployment.UpstreamType == core.UpstreamTypeImage {
		err = m.buildApplicationForDockerImage(deployment, *db, dbWithoutTx, pubSubClient, ctx, cancelContext, dockerManager)
		if err != nil {
			return err
		}
	}
	// #####  FOR GIT  ######
	if deployment.UpstreamType == core.UpstreamTypeGit {
		err = m.buildApplicationForGit(deployment, *db, dbWithoutTx, pubSubClient, ctx, cancelContext, dockerManager)
		if err != nil {
			return err
		}
	}
	// #####  FOR SOURCE CODE TARBALL  ######
	if deployment.UpstreamType == core.UpstreamTypeSourceCode {
		err = m.buildApplicationForTarball(deployment, *db, dbWithoutTx, pubSubClient, ctx, cancelContext, dockerManager)
		if err != nil {
			return err
		}
	}
	// Push image to registry
	if deployment.UpstreamType == core.UpstreamTypeGit || deployment.UpstreamType == core.UpstreamTypeSourceCode {
		err = m.pushImageToRegistry(deployment, *db, dbWithoutTx, pubSubClient, ctx, cancelContext, dockerManager)
		if err != nil {
			return err
		}
	}

	// Update deployment status
	err = deployment.UpdateStatus(ctx, *db, core.DeploymentStatusDeployPending)
	if err != nil {
		return err
	}
	// commit the transaction
	err = db.Commit().Error
	if err != nil {
		return err
	}

	// push task to queue for deployment
	err = m.EnqueueDeployApplicationRequest(deployment.ApplicationID, deployment.ID)
	if err == nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment has been triggered. Waiting for deployment to complete\n", false)
	}
	return err
}

func (m Manager) buildApplicationForDockerImage(deployment *core.Deployment, _ gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client, _ context.Context, _ context.CancelFunc, _ *containermanger.Manager) error {
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "As the upstream type is image, no build is required\n", false)
	return nil
}

func (m Manager) buildApplicationForGit(deployment *core.Deployment, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client, ctx context.Context, _ context.CancelFunc, dockerManager *containermanger.Manager) error {
	gitUsername := ""
	gitPassword := ""
	gitPrivateKey := ""

	if deployment.GitCredentialID != nil {
		// fetch git credentials
		gitCredentials := &core.GitCredential{}
		err := gitCredentials.FindById(ctx, db, *deployment.GitCredentialID)
		if err != nil {
			addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch git credentials\n", true)
			return err
		}
		gitUsername = gitCredentials.Username
		gitPassword = gitCredentials.Password
		gitPrivateKey = gitCredentials.SshPrivateKey
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
	// clone git repository
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Cloning git repository > "+deployment.GitRepositoryURL()+"\n", false)
	commitHash, commitMessage, err := gitmanager.CloneRepository(deployment.GitRepositoryURL(), deployment.RepositoryBranch, gitUsername, gitPassword, gitPrivateKey, tempDirectory)
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to clone git repository\n", false)
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Reason > "+err.Error()+"\n", true)
		return err
	}
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Cloned git repository successfully\n", false)
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Commit message > "+commitMessage+"\n", false)
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Commit hash > "+commitHash+"\n", false)
	deployment.CommitHash = commitHash
	deployment.CommitMessage = commitMessage
	// update deployment
	err = dbWithoutTx.Model(&deployment).Updates(map[string]interface{}{"commit_hash": commitHash, "commit_message": commitMessage}).Error
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to update commit hash and message in database\n", false)
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Reason > "+err.Error()+"\n", true)
		return err
	}
	// build docker image
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Started building docker image\n", false)
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
	scanner, err := dockerManager.CreateImageWithContext(ctx, deployment.Dockerfile, buildArgsMap, tempDirectory, deployment.CodePath, deployment.DeployableDockerImageURI(m.Config.ImageRegistryURI()))
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to build docker image\n", true)
		return err
	}
	isErrorEncountered := false
	if scanner != nil {
		var data map[string]interface{}
		for scanner.Scan() {
			b := scanner.Bytes()
			err = json.Unmarshal(b, &data)
			if err != nil {
				continue
			}
			if data["stream"] != nil {
				addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["stream"].(string), false)
			}
			if data["error"] != nil {
				addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["error"].(string), false)
				isErrorEncountered = true
				break
			}
		}
	}
	select {
	case <-ctx.Done():
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build cancelled !\n", true)
		return errors.New("docker image build cancelled")
	default:
		if isErrorEncountered {
			addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build failed\n", true)
			return errors.New("docker image build failed\n")
		}
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image built successfully\n", false)
		return nil
	}
}

func (m Manager) buildApplicationForTarball(deployment *core.Deployment, db gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client, ctx context.Context, _ context.CancelFunc, dockerManager *containermanger.Manager) error {
	tarballPath := filepath.Join(m.Config.LocalConfig.ServiceConfig.TarballDirectoryPath, deployment.SourceCodeCompressedFileName)
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
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Started building docker image\n", false)
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
	scanner, err := dockerManager.CreateImageWithContext(ctx, deployment.Dockerfile, buildArgsMap, tempDirectory, deployment.CodePath, deployment.DeployableDockerImageURI(m.Config.ImageRegistryURI()))
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to build docker image\n", true)
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
				addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["stream"].(string), false)
			}
			if data["error"] != nil {
				addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["error"].(string), false)
				isErrorEncountered = true
				break
			}
		}
	}
	select {
	case <-ctx.Done():
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build cancelled !\n", true)
		return errors.New("docker image build cancelled")
	default:
		if isErrorEncountered {
			addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image build failed\n", true)
			return errors.New("docker image build failed\n")
		}
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Docker image built successfully\n", false)
		return nil
	}
}

func (m Manager) pushImageToRegistry(deployment *core.Deployment, _ gorm.DB, dbWithoutTx gorm.DB, pubSubClient pubsub.Client, ctx context.Context, _ context.CancelFunc, dockerManager *containermanger.Manager) error {
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image : "+deployment.DeployableDockerImageURI(m.Config.ImageRegistryURI())+"\n", false)
	scanner, err := dockerManager.PushImage(ctx, deployment.DeployableDockerImageURI(m.Config.ImageRegistryURI()), m.Config.ImageRegistryUsername(), m.Config.ImageRegistryPassword())
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to push image to registry\n", true)
		return err
	} else {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Started pushing image to registry\n", false)
	}
	isErrorEncountered := false
	dockerPushLogger := newDockerPushLogger()
	if scanner != nil {
		var data map[string]interface{}
		for scanner.Scan() {
			b := scanner.Bytes()
			err = json.Unmarshal(b, &data)
			if err != nil {
				continue
			}
			if data["error"] != nil {
				addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, data["error"].(string)+"\n", false)
				isErrorEncountered = true
				break
			}
			addNonPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, dockerPushLogger.push(data), false)
		}
	}
	addPersistentNonRealtimeDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, dockerPushLogger.content(true), false)
	select {
	case <-ctx.Done():
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image push cancelled !\n", true)
		return errors.New("image push cancelled")
	default:
		if isErrorEncountered {
			addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image push failed\n", true)
			return errors.New("image push failed\n")
		}
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image pushed to registry successfully\n", false)
		return nil
	}
}
