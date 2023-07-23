package gitmanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/go-github/github"
	"github.com/google/uuid"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
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

// Clone repository to local folder
func (m Manager) CloneRepository(repo Repository) (string, error){
	// create a tmp folder
	tmpFolder := "/tmp/keroku/" + uuid.New().String()
	err := os.MkdirAll(tmpFolder, 0777)
	if err != nil {
		fmt.Println(err)
		return "", errors.New("failed to create tmp folder")
	}
	// clone the repo
	// TODO: auth
	_, err = git.PlainClone(tmpFolder, false, &git.CloneOptions{
		URL:      m.generateGithubURL(repo),
		Progress: nil,
		ReferenceName: plumbing.NewBranchReferenceName(repo.Branch),
	})
	if err != nil {
		fmt.Println(err)
		// cleanup
		os.RemoveAll(tmpFolder)
		return "", errors.New("failed to clone repository")
	}
	return tmpFolder, nil
}


func (m Manager) generateGithubURL(repo Repository) string{
	return "https://github.com/"+repo.Username+"/"+repo.Name+".git"
}