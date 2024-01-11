package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.42

import (
	"context"
	"errors"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// Application is the resolver for the application field.
func (r *deploymentResolver) Application(ctx context.Context, obj *model.Deployment) (*model.Application, error) {
	// fetch record
	var application = &core.Application{}
	err := application.FindById(ctx, r.ServiceManager.DbClient, obj.ApplicationID)
	if err != nil {
		return nil, err
	}
	return applicationToGraphqlObject(application), nil
}

// GitCredential is the resolver for the gitCredential field.
func (r *deploymentResolver) GitCredential(ctx context.Context, obj *model.Deployment) (*model.GitCredential, error) {
	var gitCredential = &core.GitCredential{}
	if obj.GitCredentialID != 0 {
		err := gitCredential.FindById(ctx, r.ServiceManager.DbClient, obj.GitCredentialID)
		if err != nil {
			return nil, err
		}
	}
	return gitCredentialToGraphqlObject(gitCredential), nil
}

// ImageRegistryCredential is the resolver for the imageRegistryCredential field.
func (r *deploymentResolver) ImageRegistryCredential(ctx context.Context, obj *model.Deployment) (*model.ImageRegistryCredential, error) {
	var imageRegistryCredential = &core.ImageRegistryCredential{}
	if obj.ImageRegistryCredentialID != 0 {
		err := imageRegistryCredential.FindById(ctx, r.ServiceManager.DbClient, obj.ImageRegistryCredentialID)
		if err != nil {
			return nil, err
		}
	}
	return imageRegistryCredentialToGraphqlObject(imageRegistryCredential), nil
}

// BuildArgs is the resolver for the buildArgs field.
func (r *deploymentResolver) BuildArgs(ctx context.Context, obj *model.Deployment) ([]*model.BuildArg, error) {
	// fetch record
	records, err := core.FindBuildArgsByDeploymentId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.BuildArg, 0)
	for _, record := range records {
		result = append(result, buildArgToGraphqlObject(record))
	}
	return result, nil
}

// CancelDeployment is the resolver for the cancelDeployment field.
func (r *mutationResolver) CancelDeployment(ctx context.Context, id string) (bool, error) {
	deployment := &core.Deployment{}
	deployment.ID = id
	err := deployment.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	if deployment.Status != core.DeploymentStatusPending {
		return false, errors.New("pending deployment only can be cancelled")
	}
	err = r.ServiceManager.PubSubClient.Publish(r.ServiceManager.CancelImageBuildTopic, id)
	if err != nil {
		return false, errors.New("failed to request deployment cancellation")
	}
	return true, nil
}

// Deployment is the resolver for the deployment field.
func (r *queryResolver) Deployment(ctx context.Context, id string) (*model.Deployment, error) {
	var deployment = &core.Deployment{}
	err := deployment.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	return deploymentToGraphqlObject(deployment), nil
}

// Deployment returns DeploymentResolver implementation.
func (r *Resolver) Deployment() DeploymentResolver { return &deploymentResolver{r} }

type deploymentResolver struct{ *Resolver }
