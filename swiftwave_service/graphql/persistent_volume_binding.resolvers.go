package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.46

import (
	"context"

	dbmodel "github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// PersistentVolume is the resolver for the persistentVolume field.
func (r *persistentVolumeBindingResolver) PersistentVolume(ctx context.Context, obj *model.PersistentVolumeBinding) (*model.PersistentVolume, error) {
	var persistentVolume = &dbmodel.PersistentVolume{}
	err := persistentVolume.FindById(ctx, r.ServiceManager.DbClient, obj.PersistentVolumeID)
	if err != nil {
		return nil, err
	}
	return persistentVolumeToGraphqlObject(persistentVolume), nil
}

// Application is the resolver for the application field.
func (r *persistentVolumeBindingResolver) Application(ctx context.Context, obj *model.PersistentVolumeBinding) (*model.Application, error) {
	var application = &dbmodel.Application{}
	err := application.FindById(ctx, r.ServiceManager.DbClient, obj.ApplicationID)
	if err != nil {
		return nil, err
	}
	return applicationToGraphqlObject(application), nil
}

// PersistentVolumeBinding returns PersistentVolumeBindingResolver implementation.
func (r *Resolver) PersistentVolumeBinding() PersistentVolumeBindingResolver {
	return &persistentVolumeBindingResolver{r}
}

type persistentVolumeBindingResolver struct{ *Resolver }
