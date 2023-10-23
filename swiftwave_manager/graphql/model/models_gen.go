// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Application struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	EnvironmentVariables []*EnvironmentVariable `json:"environmentVariables"`
}

type EnvironmentVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GitCredential struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type GitCredentialInput struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type GitCredentialRepositoryAccessInput struct {
	GitCredentialID  int    `json:"gitCredentialId"`
	RepositoryURL    string `json:"repositoryUrl"`
	RepositoryBranch string `json:"repositoryBranch"`
}

type GitCredentialRepositoryAccessResult struct {
	GitCredentialID  int            `json:"gitCredentialId"`
	GitCredential    *GitCredential `json:"gitCredential"`
	RepositoryURL    string         `json:"repositoryUrl"`
	RepositoryBranch string         `json:"repositoryBranch"`
	Success          bool           `json:"success"`
	Error            string         `json:"error"`
}

type ImageRegistryCredential struct {
	ID       int    `json:"id"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ImageRegistryCredentialInput struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}