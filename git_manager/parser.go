package gitmanager

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var UnknownParseError = errors.New("unknown parse error")
var invalidGitUrlError = errors.New("invalid git url")

var sshGitUrlRegexStr = `^.+@.+\:.+\/.+$`
var sshGitUrlRegex *regexp.Regexp
var httpGitUrlRegexStr = `^(https://|http://|).+/.+/.+$`
var httpGitUrlRegex *regexp.Regexp

func init() {
	sshGitUrlRegex = regexp.MustCompile(sshGitUrlRegexStr)
	httpGitUrlRegex = regexp.MustCompile(httpGitUrlRegexStr)
}

type GitRepoInfo struct {
	IsParsed      bool
	Provider      string
	Owner         string
	Name          string
	Endpoint      string
	SshUser       string
	IsSshEndpoint bool
}

func ParseGitRepoInfo(gitUrl string) (*GitRepoInfo, error) {
	// handle any panic
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	// clean up the git url
	gitUrl = strings.TrimSpace(gitUrl)

	/*
	* Example of git url:
	* https://github.com/swiftwave-org/swiftwave.git
	* https://github.com/swiftwave-org/swiftwave
	* github.com/swiftwave-org/swiftwave
	* git@github.com:swiftwave-org/swiftwave.git
	 */

	var gitRepoInfo GitRepoInfo
	if isValidSSHGitUrl(gitUrl) {
		isSeparator := func(c rune) bool {
			return c == '@' || c == ':' || c == '/'
		}
		splits := strings.FieldsFunc(gitUrl, isSeparator)
		if len(splits) != 4 {
			return nil, invalidGitUrlError
		}
		gitRepoInfo.SshUser = splits[0]
		gitRepoInfo.Endpoint = splits[1]
		gitRepoInfo.Owner = splits[2]
		gitRepoInfo.Name = splits[3]
		gitRepoInfo.Name = strings.TrimSuffix(gitRepoInfo.Name, ".git")
		gitRepoInfo.Provider = gitProvider(gitRepoInfo.Endpoint)
		gitRepoInfo.IsSshEndpoint = true
		gitRepoInfo.IsParsed = true
		return &gitRepoInfo, nil
	} else if isValidHttpGitUrl(gitUrl) {
		isHttps := true
		if strings.HasPrefix(gitUrl, "http://") {
			isHttps = false
		}
		// strip http:// or https://
		gitUrl = strings.TrimPrefix(gitUrl, "http://")
		gitUrl = strings.TrimPrefix(gitUrl, "https://")
		// strip if ends has / or .git
		gitUrl = strings.TrimSuffix(gitUrl, "/")
		splits := strings.Split(gitUrl, "/")
		if len(splits) < 3 {
			return nil, invalidGitUrlError
		}
		gitRepoInfo.Endpoint = splits[0]
		if isHttps {
			gitRepoInfo.Endpoint = "https://" + gitRepoInfo.Endpoint
		} else {
			gitRepoInfo.Endpoint = "http://" + gitRepoInfo.Endpoint
		}
		gitRepoInfo.Owner = splits[1]
		gitRepoInfo.Name = splits[2]
		gitRepoInfo.Name = strings.TrimSuffix(gitRepoInfo.Name, ".git")
		gitRepoInfo.Provider = gitProvider(gitRepoInfo.Endpoint)
		gitRepoInfo.IsSshEndpoint = false
		gitRepoInfo.IsParsed = true
		return &gitRepoInfo, nil
	}

	return nil, UnknownParseError
}

func isValidSSHGitUrl(gitUrl string) bool {
	return sshGitUrlRegex.MatchString(gitUrl)
}

func isValidHttpGitUrl(gitUrl string) bool {
	return httpGitUrlRegex.MatchString(gitUrl)
}

func gitProvider(endpoint string) string {
	if strings.Contains(endpoint, "github.com") {
		return "github"
	} else if strings.Contains(endpoint, "gitlab.com") {
		return "gitlab"
	} else if strings.Contains(endpoint, "bitbucket.org") {
		return "bitbucket"
	} else {
		return endpoint
	}
}

func (gitRepoInfo *GitRepoInfo) URL() string {
	if !gitRepoInfo.IsParsed {
		return ""
	}
	if gitRepoInfo.IsSshEndpoint {
		return fmt.Sprintf("%s@%s:%s/%s", gitRepoInfo.SshUser, gitRepoInfo.Endpoint, gitRepoInfo.Owner, gitRepoInfo.Name)
	}
	return fmt.Sprintf("%s/%s/%s", gitRepoInfo.Endpoint, gitRepoInfo.Owner, gitRepoInfo.Name)
}
