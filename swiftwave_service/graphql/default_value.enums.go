package graphql

import "github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"

func DefaultGitProvider(value *model.GitProvider) model.GitProvider {
	if value == nil {
		return model.GitProviderNone
	}
	return *value
}
