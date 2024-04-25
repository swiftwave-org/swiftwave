package gitmanager

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var UnknownParseError = errors.New("unknown parse error")
var invalidGitUrlError = errors.New("invalid git url")

var sshGitUrlRegexV1Str = `^.+@.+\:.+\/.+$`
var sshGitUrlRegexV1 *regexp.Regexp
var sshGitUrlRegexV2 *regexp.Regexp
var sshGitUrlRegexV2Str = `^ssh:\/\/.+@.+\/.+$`
var httpGitUrlRegexStr = `^(https://|http://|).+/.+$`
var httpGitUrlRegex *regexp.Regexp

func init() {
	sshGitUrlRegexV1 = regexp.MustCompile(sshGitUrlRegexV1Str)
	sshGitUrlRegexV2 = regexp.MustCompile(sshGitUrlRegexV2Str)
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
	// clean up the git url
	gitUrl = strings.TrimSpace(gitUrl)

	/*
	* Example of git url:
	* https://github.com/swiftwave-org/swiftwave.git
	* https://github.com/swiftwave-org/swiftwave
	* github.com/swiftwave-org/swiftwave
	* git@github.com:swiftwave-org/swiftwave.git
	* v2 ssh format
	* ssh://git@host.xz:2222/path/to/repo.git/
	 */

	var gitRepoInfo GitRepoInfo
	if isValidSSHGitUrlV2(gitUrl) {
		url := strings.TrimPrefix(gitUrl, "ssh://")
		isSeparator := func(c rune) bool {
			return c == '@' || c == '/'
		}
		splits := strings.FieldsFunc(url, isSeparator)
		if len(splits) < 3 {
			return nil, invalidGitUrlError
		}
		gitRepoInfo.SshUser = splits[0]
		gitRepoInfo.Endpoint = splits[1]
		gitRepoInfo.Owner = strings.Join(splits[2:len(splits)-1], "/")
		gitRepoInfo.Name = splits[len(splits)-1]
		gitRepoInfo.Name = strings.TrimSuffix(gitRepoInfo.Name, ".git")
		gitRepoInfo.Provider = gitProvider(gitRepoInfo.Endpoint)
		gitRepoInfo.IsSshEndpoint = true
		gitRepoInfo.IsParsed = true
		return &gitRepoInfo, nil
	} else if isValidSSHGitUrlV1(gitUrl) {
		isSeparator := func(c rune) bool {
			return c == '@' || c == ':' || c == '/'
		}
		splits := strings.FieldsFunc(gitUrl, isSeparator)
		if len(splits) < 3 {
			return nil, invalidGitUrlError
		}
		gitRepoInfo.SshUser = splits[0]
		gitRepoInfo.Endpoint = splits[1]
		gitRepoInfo.Owner = strings.Join(splits[2:len(splits)-1], "/")
		gitRepoInfo.Name = splits[len(splits)-1]
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
		if len(splits) < 2 {
			return nil, invalidGitUrlError
		}
		gitRepoInfo.Endpoint = splits[0]
		if isHttps {
			gitRepoInfo.Endpoint = "https://" + gitRepoInfo.Endpoint
		} else {
			gitRepoInfo.Endpoint = "http://" + gitRepoInfo.Endpoint
		}
		gitRepoInfo.Owner = strings.Join(splits[1:len(splits)-1], "/")
		gitRepoInfo.Name = splits[len(splits)-1]
		gitRepoInfo.Name = strings.TrimSuffix(gitRepoInfo.Name, ".git")
		gitRepoInfo.Provider = gitProvider(gitRepoInfo.Endpoint)
		gitRepoInfo.IsSshEndpoint = false
		gitRepoInfo.IsParsed = true
		return &gitRepoInfo, nil
	}

	return nil, UnknownParseError
}

func isValidSSHGitUrlV1(gitUrl string) bool {
	return sshGitUrlRegexV1.MatchString(gitUrl)
}

func isValidSSHGitUrlV2(gitUrl string) bool {
	return sshGitUrlRegexV2.MatchString(gitUrl)
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
		if strings.Compare(gitRepoInfo.Owner, "") == 0 {
			if strings.Contains(gitRepoInfo.Endpoint, ":") {
				// if port is present, then use v2 ssh format
				return fmt.Sprintf("ssh://%s@%s/%s", gitRepoInfo.SshUser, gitRepoInfo.Endpoint, gitRepoInfo.Name)
			} else {
				return fmt.Sprintf("%s@%s:%s", gitRepoInfo.SshUser, gitRepoInfo.Endpoint, gitRepoInfo.Name)
			}
		} else {
			if strings.Contains(gitRepoInfo.Endpoint, ":") {
				// if port is present, then use v2 ssh format
				return fmt.Sprintf("ssh://%s@%s/%s/%s", gitRepoInfo.SshUser, gitRepoInfo.Endpoint, gitRepoInfo.Owner, gitRepoInfo.Name)
			} else {
				return fmt.Sprintf("%s@%s:%s/%s", gitRepoInfo.SshUser, gitRepoInfo.Endpoint, gitRepoInfo.Owner, gitRepoInfo.Name)

			}
		}
	} else {
		if strings.Compare(gitRepoInfo.Owner, "") == 0 {
			return fmt.Sprintf("%s/%s", gitRepoInfo.Endpoint, gitRepoInfo.Name)

		} else {
			return fmt.Sprintf("%s/%s/%s", gitRepoInfo.Endpoint, gitRepoInfo.Owner, gitRepoInfo.Name)
		}
	}
}

func isGitSSHAgentForwardingEnabled() bool {
	// check if SSH_AUTH_SOCK is set
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if strings.Compare(sshAuthSock, "") == 0 {
		return false
	}
	// check if SSH_KNOWN_HOSTS is set
	sshKnownHosts := os.Getenv("SSH_KNOWN_HOSTS")
	if strings.Compare(sshKnownHosts, "") == 0 {
		return false
	}
	return true
}
