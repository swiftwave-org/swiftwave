package server

import (
	"net/url"
	"strings"
)

func FetchRepositoryUsernameFromURL(repo_url string) string {
	// parse
	url , err := url.Parse(repo_url)
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

func FetchRepositoryNameFromURL(repo_url string) string {
	// parse
	url , err := url.Parse(repo_url)
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