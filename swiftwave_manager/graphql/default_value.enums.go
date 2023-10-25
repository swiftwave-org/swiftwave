package graphql

import "github.com/swiftwave-org/swiftwave/swiftwave_manager/graphql/model"

func DefaultGitProvider(value *model.GitProvider) model.GitProvider {
	if value == nil {
		return model.GitProviderNone
	}
	return *value
}
