package gitmanager

import (
	"context"
	"io"

	"github.com/google/go-github/github"
)

// List all repositories for the user.
func (m Manager) FetchRepositories() ([]Repository, error){
	repositories, _, err := github.NewClient(nil).Repositories.List(context.Background(), m.GitUser.Username, nil)
	if err != nil {
		return []Repository{}, err
	}
	var result []Repository = []Repository{};;

	for _, repo := range repositories {
		result = append(result, Repository{
			Name: *repo.Name,
			Username: m.GitUser.Username,
			Branch: *repo.DefaultBranch,
			IsPrivate: *repo.Private,
		})
	}
	
	return result, nil;
}

// Fetch latest commit hash for a repository.
func (m Manager) FetchLatestCommitHash(repo Repository) (string, error){
	commits, _, err := github.NewClient(nil).Repositories.ListCommits(context.Background(), repo.Username, repo.Name, &github.CommitsListOptions{
		SHA: repo.Branch,
	})
	if err != nil {
		return "", err
	}
	return *commits[0].SHA, nil;
}

// Fetch folder structure for a repository.
func (m Manager) FetchFolderStructure(repo Repository) ([]string, error){
	tree, _, err := github.NewClient(nil).Git.GetTree(context.Background(), repo.Username, repo.Name, repo.Branch, true)
	if err != nil {
		return []string{}, err
	}
	var result []string = []string{};;
	for _, entry := range tree.Entries {
		result = append(result, *entry.Path)
	}
	return result, nil;
}

// Fetch file content for a repository.			
func (m Manager) FetchFileContent(repo Repository, path string) (string, error){
	fileContent, err := github.NewClient(nil).Repositories.DownloadContents(context.Background(), repo.Username, repo.Name, path, &github.RepositoryContentGetOptions{
		Ref: repo.Branch,
	})
	if err != nil {
		return "failed to resolve repository", err
	}
	content, err := io.ReadAll(fileContent)
	if err != nil {
		return "failed to read file content", err
	}
	return string(content), nil;
}