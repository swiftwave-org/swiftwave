package gitmanager

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"os"
	"sort"
)

func FetchLatestCommitHash(gitUrl string, branch string, username string, password string) (string, error) {
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
	// ls-remote the repo
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{gitUrl},
	})
	refs, err := remote.List(&git.ListOptions{
		Auth:            httpAuth,
		InsecureSkipTLS: true,
		PeelingOption:   git.IgnorePeeled,
	})
	if err != nil {
		return "", err
	}
	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name().Short() == branch {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.New("branch not found")
}

func FetchBranches(gitUrl string, username string, password string) ([]string, error) {
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
	// ls-remote the repo
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{gitUrl},
	})
	refs, err := remote.List(&git.ListOptions{
		Auth:            httpAuth,
		InsecureSkipTLS: true,
		PeelingOption:   git.IgnorePeeled,
	})
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
	}
	// sort the branches
	sort.Strings(branches)
	return branches, nil
}

func CloneRepository(gitUrl string, branch string, username string, password string, destFolder string) error {
	// check if folder exists
	if _, err := os.Stat(destFolder); os.IsNotExist(err) {
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
	_, err := git.PlainClone(destFolder, false, &git.CloneOptions{
		URL:               gitUrl,
		Progress:          nil,
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		Auth:              httpAuth,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		ShallowSubmodules: true,
	})
	if err != nil {
		return errors.New("failed to clone repository")
	}
	return nil
}
