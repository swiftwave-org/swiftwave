package gitmanager

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/go-git/go-git/v5/storage/memory"
	"os"
	"sort"
	"strings"
)

func FetchLatestCommitHash(gitUrl string, branch string, username string, password string, privateKey string) (string, error) {
	// Parse the URL
	repoInfo, err := ParseGitRepoInfo(gitUrl)
	if err != nil {
		return "", err
	}

	// Get the auth method
	auth, err := getAuthMethod(repoInfo, username, password, privateKey)
	if err != nil {
		return "", err
	}

	// ls-remote the repo
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{gitUrl},
	})
	refs, err := remote.List(&git.ListOptions{
		Auth:            auth,
		InsecureSkipTLS: true,
		PeelingOption:   git.IgnorePeeled,
	})
	if err != nil {
		return "", err
	}
	for _, ref := range refs {
		if ref.Name().IsBranch() && strings.Compare(ref.Name().Short(), branch) == 0 {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.New("branch not found")
}

func FetchBranches(gitUrl string, username string, password string, privateKey string) ([]string, error) {
	// Parse the URL
	repoInfo, err := ParseGitRepoInfo(gitUrl)
	if err != nil {
		return nil, err
	}

	// Get the auth method
	auth, err := getAuthMethod(repoInfo, username, password, privateKey)
	if err != nil {
		return nil, err
	}

	// ls-remote the repo
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{gitUrl},
	})
	refs, err := remote.List(&git.ListOptions{
		Auth:            auth,
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

func CloneRepository(gitUrl string, branch string, username string, password string, privateKey string, destFolder string) error {
	// Parse the URL
	repoInfo, err := ParseGitRepoInfo(gitUrl)
	if err != nil {
		return err
	}

	// Get the auth method
	auth, err := getAuthMethod(repoInfo, username, password, privateKey)
	if err != nil {
		return err
	}

	// check if folder exists
	if _, err := os.Stat(destFolder); os.IsNotExist(err) {
		return errors.New("destination folder does not exist")
	}

	// clone the repo
	_, err = git.PlainClone(destFolder, false, &git.CloneOptions{
		URL:               gitUrl,
		Progress:          nil,
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		Auth:              auth,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		ShallowSubmodules: true,
	})
	if err != nil {
		return errors.New("failed to clone repository")
	}
	return nil
}

// private function
func getAuthMethod(repoInfo *GitRepoInfo, username string, password string, privateKey string) (transport.AuthMethod, error) {
	if repoInfo == nil {
		return nil, errors.New("invalid repository info")
	}
	var auth transport.AuthMethod
	// If username and password both are provided, then use only http auth
	if username != "" && password != "" && !repoInfo.IsSshEndpoint {
		httpAuth := &http.BasicAuth{
			Username: username,
			Password: password,
		}
		auth = httpAuth
	} else if privateKey != "" && repoInfo.IsSshEndpoint {
		privateKeyAuth, err := ssh.NewPublicKeys(repoInfo.SshUser, []byte(privateKey), "")
		if err != nil {
			return nil, err
		}
		auth = privateKeyAuth
	} else {
		auth = nil
	}

	return auth, nil
}
