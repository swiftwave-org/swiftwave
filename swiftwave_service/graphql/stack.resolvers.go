package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/stack_parser"
)

// CleanupStack is the resolver for the cleanupStack field.
func (r *mutationResolver) CleanupStack(ctx context.Context, input model.StackInput) (string, error) {
	content := input.Content
	stack, err := stack_parser.ParseStackYaml(content)
	if err != nil {
		return "", err
	}
	return stack.String(true)
}

// VerifyStack is the resolver for the verifyStack field.
func (r *mutationResolver) VerifyStack(ctx context.Context, input model.StackInput) (*model.StackVerifyResult, error) {
	// parse yaml
	content := input.Content
	stack, err := stack_parser.ParseStackYaml(content)
	if err != nil {
		return nil, err
	}
	// fill variable
	variableMapping := make(map[string]string)
	for _, variable := range input.Variables {
		variableMapping[variable.Name] = variable.Value
	}
	stackFilled, err := stack.FillAndVerifyVariables(&variableMapping, r.ServiceManager)
	if err != nil {
		return nil, err
	}
	// create result
	result := &model.StackVerifyResult{
		Success:         true,
		Message:         "",
		Error:           "",
		ValidVolumes:    make([]string, 0), // volumes that
		InvalidVolumes:  make([]string, 0),
		ValidServices:   make([]string, 0),
		InvalidServices: make([]string, 0),
	}
	// fetch all the service names
	serviceNames := stackFilled.ServiceNames()
	// fetch docker manager
	dockerManager, err := FetchDockerManager(ctx, &r.ServiceManager.DbClient)
	if err != nil {
		return nil, err
	}
	// check if any service name is existing in database
	for _, serviceName := range serviceNames {
		isExist, err := core.IsExistApplicationName(ctx, r.ServiceManager.DbClient, *dockerManager, serviceName)
		if err != nil {
			return nil, err
		}
		if !isExist {
			result.ValidServices = append(result.ValidServices, serviceName)
		} else {
			result.InvalidServices = append(result.InvalidServices, serviceName)
		}
	}
	// check volume names
	volumeNames := stackFilled.VolumeNames()
	for _, volumeName := range volumeNames {
		isExist, err := core.IsExistPersistentVolume(ctx, r.ServiceManager.DbClient, volumeName, *dockerManager)
		if err != nil {
			return nil, err
		}
		if isExist {
			result.ValidVolumes = append(result.ValidVolumes, volumeName)
		} else {
			result.InvalidVolumes = append(result.InvalidVolumes, volumeName)
		}
	}
	// set message
	if len(result.InvalidServices) == 0 {
		result.Success = result.Success && true
		result.Message = "All services are verified"
	} else {
		result.Success = result.Success && false
		unverifiedServiceStr := ""
		for _, service := range result.InvalidServices {
			unverifiedServiceStr += service + ", "
		}
		if len(unverifiedServiceStr) > 2 {
			unverifiedServiceStr = unverifiedServiceStr[:len(unverifiedServiceStr)-2]
		}
		result.Error = fmt.Sprintf("%s\nConflicting services -> %s . Please change stack name", result.Error, unverifiedServiceStr)
	}

	if len(result.InvalidVolumes) == 0 {
		result.Success = result.Success && true
		result.Message = fmt.Sprintf("%s\nAll volumes are verified", result.Message)
	} else {
		result.Success = result.Success && false
		unverifiedVolumeStr := ""
		for _, volume := range result.InvalidVolumes {
			unverifiedVolumeStr += volume + ", "
		}
		if len(unverifiedVolumeStr) > 2 {
			unverifiedVolumeStr = unverifiedVolumeStr[:len(unverifiedVolumeStr)-2]
		}
		result.Error = fmt.Sprintf("%s\nThese volumes doesn't exist -> %s . Please create volumes from dashboard.", result.Error, unverifiedVolumeStr)
	}
	result.Message = strings.TrimSpace(result.Message)
	result.Error = strings.TrimSpace(result.Error)
	return result, nil
}

// DeployStack is the resolver for the deployStack field.
func (r *mutationResolver) DeployStack(ctx context.Context, input model.StackInput) ([]*model.ApplicationDeployResult, error) {
	// parse stack
	stack, err := stack_parser.ParseStackYaml(input.Content)
	if err != nil {
		return nil, errors.New("stack configuration is not valid")
	}
	// verify stack
	verifyResult, err := r.VerifyStack(ctx, input)
	if err != nil {
		return nil, err
	} else {
		if !verifyResult.Success {
			return nil, fmt.Errorf("stack configuration is not valid. %s", verifyResult.Error)
		}
	}
	// Fill variable
	variableMapping := make(map[string]string)
	for _, variable := range input.Variables {
		variableMapping[variable.Name] = variable.Value
	}
	stackFilled, err := stack.FillAndVerifyVariables(&variableMapping, r.ServiceManager)
	if err != nil {
		return nil, err
	}

	// convert to application input
	applicationsInput, err := stackToApplicationsInput(stackFilled, r.ServiceManager.DbClient)
	if err != nil {
		return nil, err
	}
	// result
	results := make([]*model.ApplicationDeployResult, 0)
	// create application
	for _, applicationInput := range applicationsInput {
		application, err := r.CreateApplication(ctx, applicationInput)
		if err != nil {
			results = append(results, &model.ApplicationDeployResult{
				Success:     false,
				Message:     err.Error(),
				Application: application,
			})
		} else {
			results = append(results, &model.ApplicationDeployResult{
				Success:     true,
				Message:     "Application created successfully",
				Application: application,
			})
		}
	}
	return results, nil
}
