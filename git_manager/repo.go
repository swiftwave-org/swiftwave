package gitmanager

import (
	"errors"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Fetch latest commit hash for a repository.
func FetchLatestCommitHash(git_url string, branch string, username string, password string) (string, error) {
	var httpAuth *http.BasicAuth
	// If username and password both are provided, then use only http auth
	if username != "" && password != "" {
		httpAuth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	} else {
		httpAuth = nil
	}
	// clone the repo
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:               git_url,
		SingleBranch:      true,
		Progress:          nil,
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		Auth:              httpAuth,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return "", errors.New("failed to clone repository")
	}
	// get the head of the repo
	ref, err := r.Head()
	if err != nil {
		return "", errors.New("failed to get head")
	}
	// return the hash
	return ref.Hash().String(), nil
}

// Clone repository to local folder
func CloneRepository(git_url string, branch string, username string, password string, dest_folder string) error {
	// check if folder exists
	if _, err := os.Stat(dest_folder); os.IsNotExist(err) {
		return errors.New("destination folder does not exist")
	}
	var httpAuth *http.BasicAuth
	// If username and password both are provided, then use only http auth
	if username != "" && password != "" {
		httpAuth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	} else {
		httpAuth = nil
	}
	// clone the repo
	_, err := git.PlainClone(dest_folder, false, &git.CloneOptions{
		URL:               git_url,
		Progress:          nil,
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		Auth:              httpAuth,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		log.Println(err)
		return errors.New("failed to clone repository")
	}
	return nil
}
