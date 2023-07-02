package haproxymanager

import (
	"net/url"
	"strings"
)

func queryParamsToString(queryParams QueryParameters) string {
	tmp := "?"
	for _, param := range queryParams {
		tmp += param.key + "=" + url.QueryEscape(param.value) + "&"
	}
	tmp = strings.TrimSuffix(tmp, "&")
	return tmp
}

func (q *QueryParameters) add(key string, value string) {
	*q = append(*q, QueryParameter{key, value})
}
