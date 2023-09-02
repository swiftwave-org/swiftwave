package server

import (
	"net/url"
	"path/filepath"
	"strings"
)

// FetchRepositoryUsernameFromURL returns the username from a repository url.
func FetchRepositoryUsernameFromURL(repo_url string) string {
	// parse
	url, err := url.Parse(repo_url)
	if err != nil {
		return ""
	}
	// split at /
	splits := strings.Split(url.Path, "/")
	if len(splits) >= 3 {
		return splits[1]
	}
	return ""
}

// FetchRepositoryNameFromURL returns the repository name from a repository url.
func FetchRepositoryNameFromURL(repo_url string) string {
	// parse
	url, err := url.Parse(repo_url)
	if err != nil {
		return ""
	}
	// split at /
	splits := strings.Split(url.Path, "/")
	if len(splits) >= 3 {
		return splits[2]
	}
	return ""
}

// FetchGitProviderFromURL returns the git provider from a repository url.
func FetchGitProviderFromURL(repo_url string) GitProvider {
	if strings.Contains(repo_url, "github.com") {
		return GitProviderGithub
	}
	if strings.Contains(repo_url, "gitlab.com") {
		return GitProviderGitlab
	}

	return ""
}

// Sanitize the fileName to remove potentially dangerous characters
func SanitizeFileName(fileName string) string {
	// Remove any path components and keep only the file name
	fileName = filepath.Base(fileName)

	// Remove potentially dangerous characters like ".."
	fileName = strings.ReplaceAll(fileName, "..", "")

	// You can add more sanitization rules as needed

	return fileName
}