package haproxymanager

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (s HAProxySocket) URI() string {
	if s.isUnix {
		return "unix://" + s.unixSocketPath
	}
	return "http://" + s.Host + ":" + strconv.Itoa(s.Port) + "/v2"
}

func (s HAProxySocket) getRequest(route string, queryParams QueryParameters) (*http.Response, error) {
	if !strings.HasPrefix(route, "/"){
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams);
	req, err := http.NewRequest("GET", url, nil);
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password);
	client := &http.Client{}
	return client.Do(req)
}

func (s HAProxySocket) deleteRequest(route string, queryParams QueryParameters) (*http.Response, error) {
	if !strings.HasPrefix(route, "/"){
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams);
	req, err := http.NewRequest("DELETE", url, nil);
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password);
	client := &http.Client{}
	return client.Do(req)
}

func (s HAProxySocket) postRequest(route string, queryParams QueryParameters, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/"){
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams);
	req, err := http.NewRequest("POST", url, body);
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password);
	req.Header.Add("Content-Type", "application/json");
	client := &http.Client{}
	return client.Do(req)
}

func (s HAProxySocket) putRequest(route string, queryParams QueryParameters, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/"){
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams);
	req, err := http.NewRequest("PUT", url, body);
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password);
	req.Header.Add("Content-Type", "application/json");
	client := &http.Client{}
	return client.Do(req)
}