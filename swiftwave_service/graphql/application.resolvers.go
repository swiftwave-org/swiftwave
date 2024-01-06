package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.42

import (
	"context"
	"errors"

	gitmanager "github.com/swiftwave-org/swiftwave/git_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// EnvironmentVariables is the resolver for the environmentVariables field.
func (r *applicationResolver) EnvironmentVariables(ctx context.Context, obj *model.Application) ([]*model.EnvironmentVariable, error) {
	// fetch record
	records, err := core.FindEnvironmentVariablesByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.EnvironmentVariable, 0)
	for _, record := range records {
		result = append(result, environmentVariableToGraphqlObject(record))
	}
	return result, nil
}

// PersistentVolumeBindings is the resolver for the persistentVolumeBindings field.
func (r *applicationResolver) PersistentVolumeBindings(ctx context.Context, obj *model.Application) ([]*model.PersistentVolumeBinding, error) {
	// fetch record
	records, err := core.FindPersistentVolumeBindingsByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.PersistentVolumeBinding, 0)
	for _, record := range records {
		result = append(result, persistentVolumeBindingToGraphqlObject(record))
	}
	return result, nil
}

// RealtimeInfo is the resolver for the realtimeInfo field.
func (r *applicationResolver) RealtimeInfo(ctx context.Context, obj *model.Application) (*model.RealtimeInfo, error) {
	info, err := r.ServiceManager.DockerManager.RealtimeInfoService(obj.Name, true)
	if err != nil {
		return &model.RealtimeInfo{
			InfoFound:       false,
			DesiredReplicas: 0,
			RunningReplicas: 0,
			DeploymentMode:  model.DeploymentModeGlobal,
		}, nil
	}
	var deploymentMode model.DeploymentMode
	if info.ReplicatedService {
		deploymentMode = model.DeploymentModeReplicated
	} else {
		deploymentMode = model.DeploymentModeGlobal
	}
	return &model.RealtimeInfo{
		InfoFound:       true,
		DesiredReplicas: info.DesiredReplicas,
		RunningReplicas: info.RunningReplicas,
		DeploymentMode:  deploymentMode,
	}, nil
}

// LatestDeployment is the resolver for the latestDeployment field.
func (r *applicationResolver) LatestDeployment(ctx context.Context, obj *model.Application) (*model.Deployment, error) {
	// fetch record
	record, err := core.FindLatestDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	return deploymentToGraphqlObject(record), nil
}

// Deployments is the resolver for the deployments field.
func (r *applicationResolver) Deployments(ctx context.Context, obj *model.Application) ([]*model.Deployment, error) {
	// fetch record
	records, err := core.FindDeploymentsByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.Deployment, 0)
	for _, record := range records {
		result = append(result, deploymentToGraphqlObject(record))
	}
	return result, nil
}

// IngressRules is the resolver for the ingressRules field.
func (r *applicationResolver) IngressRules(ctx context.Context, obj *model.Application) ([]*model.IngressRule, error) {
	// fetch record
	records, err := core.FindIngressRulesByApplicationID(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.IngressRule, 0)
	for _, record := range records {
		result = append(result, ingressRuleToGraphqlObject(record))
	}
	return result, nil
}

// CreateApplication is the resolver for the createApplication field.
func (r *mutationResolver) CreateApplication(ctx context.Context, input model.ApplicationInput) (*model.Application, error) {
	record := applicationInputToDatabaseObject(&input)
	// create transaction
	transaction := r.ServiceManager.DbClient.Begin()
	err := record.Create(ctx, *transaction, r.ServiceManager.DockerManager, r.ServiceConfig.ServiceConfig.DataDir)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	err = transaction.Commit().Error
	if err != nil {
		return nil, err
	}
	// fetch latest deployment
	latestDeployment, err := core.FindLatestDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, record.ID)
	if err != nil {
		return nil, errors.New("failed to fetch latest deployment")
	}
	// push build request to worker
	err = r.WorkerManager.EnqueueBuildApplicationRequest(record.ID, latestDeployment.ID)
	if err != nil {
		return nil, errors.New("failed to process application build request")
	}
	return applicationToGraphqlObject(record), nil
}

// UpdateApplication is the resolver for the updateApplication field.
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, input model.ApplicationInput) (*model.Application, error) {
	// fetch record
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	// convert input to database object
	var databaseObject = applicationInputToDatabaseObject(&input)
	databaseObject.ID = record.ID
	databaseObject.LatestDeployment.ApplicationID = record.ID
	if databaseObject.LatestDeployment.UpstreamType == core.UpstreamTypeGit {
		gitUsername := ""
		gitPassword := ""
		if databaseObject.LatestDeployment.GitCredentialID != nil {
			var gitCredential core.GitCredential
			if err := gitCredential.FindById(ctx, r.ServiceManager.DbClient, *databaseObject.LatestDeployment.GitCredentialID); err != nil {
				return nil, errors.New("invalid git credential provided")
			}
			gitUsername = gitCredential.Username
			gitPassword = gitCredential.Password
		}
		
		commitHash, err := gitmanager.FetchLatestCommitHash(databaseObject.LatestDeployment.GitRepositoryURL(), databaseObject.LatestDeployment.RepositoryBranch, gitUsername, gitPassword)
		if err != nil {
			return nil, errors.New("failed to fetch latest commit hash")
		}
		databaseObject.LatestDeployment.CommitHash = commitHash
	}

	// update record
	result, err := databaseObject.Update(ctx, r.ServiceManager.DbClient, r.ServiceManager.DockerManager)
	if err != nil {
		return nil, err
	} else {
		if result.RebuildRequired {
			// fetch latest deployment
			latestDeployment, err := core.FindLatestDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, record.ID)
			if err != nil {
				return nil, err
			}
			err = r.WorkerManager.EnqueueBuildApplicationRequest(record.ID, latestDeployment.ID)
			if err != nil {
				return nil, errors.New("failed to process application build request")
			}
		} else if result.ReloadRequired {
			err = r.WorkerManager.EnqueueDeployApplicationRequest(record.ID)
			if err != nil {
				return nil, errors.New("failed to process application deploy request")
			}
		}
	}
	return applicationToGraphqlObject(databaseObject), nil
}

// DeleteApplication is the resolver for the deleteApplication field.
func (r *mutationResolver) DeleteApplication(ctx context.Context, id string) (bool, error) {
	// fetch record
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// delete record
	err = record.SoftDelete(ctx, r.ServiceManager.DbClient, r.ServiceManager.DockerManager)
	if err != nil {
		return false, err
	}
	// push delete request to worker
	err = r.WorkerManager.EnqueueDeleteApplicationRequest(record.ID)
	if err != nil {
		return false, errors.New("failed to process application delete request")
	}
	return true, nil
}

// Application is the resolver for the application field.
func (r *queryResolver) Application(ctx context.Context, id string) (*model.Application, error) {
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	return applicationToGraphqlObject(record), nil
}

// Applications is the resolver for the applications field.
func (r *queryResolver) Applications(ctx context.Context) ([]*model.Application, error) {
	var records []*core.Application
	records, err := core.FindAllApplications(ctx, r.ServiceManager.DbClient)
	if err != nil {
		return nil, err
	}
	var result = make([]*model.Application, 0)
	for _, record := range records {
		result = append(result, applicationToGraphqlObject(record))
	}
	return result, nil
}

// IsExistApplicationName is the resolver for the isExistApplicationName field.
func (r *queryResolver) IsExistApplicationName(ctx context.Context, name string) (bool, error) {
	return core.IsExistApplicationName(ctx, r.ServiceManager.DbClient, r.ServiceManager.DockerManager, name)
}

// Application returns ApplicationResolver implementation.
func (r *Resolver) Application() ApplicationResolver { return &applicationResolver{r} }

type applicationResolver struct{ *Resolver }
