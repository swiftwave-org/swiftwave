package haproxymanager

import (
	"net/url"
	"strings"
)

/*
IsPortRestrictedForManualConfig :
This function is used to check if a port is restricted or not for application.

There are some ports that are restricted.
because those port are pre-occupied by Swarm services or other required services.
So, binding to those ports will cause errors.
That's why we need to restrict those ports before apply the config.
*/
func IsPortRestrictedForManualConfig(port int, restrictedPorts []int) bool {
	for _, p := range restrictedPorts {
		if port == p {
			return true
		}
	}
	return false
}

// Convert QueryParameters -> List <QueryParameter> to query string
func queryParamsToString(queryParams QueryParameters) string {
	tmp := "?"
	for _, param := range queryParams {
		tmp += param.key + "=" + url.QueryEscape(param.value) + "&"
	}
	tmp = strings.TrimSuffix(tmp, "&")
	return tmp
}

// Add a new QueryParameter to QueryParameters
func (q *QueryParameters) add(key string, value string) {
	*q = append(*q, QueryParameter{key, value})
}

// Check if a status code is OK or not
// Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#successful_responses
func isValidStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
