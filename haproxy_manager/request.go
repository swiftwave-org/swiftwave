package haproxymanager

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// Generate Base URI for HAProxy Server
func (s Manager) URI() string {
	if s.isUnix {
		return "unix://" + s.unixSocketPath
	}
	return "http://" + s.Host + ":" + strconv.Itoa(s.Port) + "/v2"
}

// Wrapper to send request to HAProxy Server
func (s Manager) getRequest(route string, queryParams QueryParameters) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	client := &http.Client{}
	return client.Do(req)
}

// Wrapper to send request to HAProxy Server
func (s Manager) deleteRequest(route string, queryParams QueryParameters) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	client := &http.Client{}
	return client.Do(req)
}

// Wrapper to send request to HAProxy Server
func (s Manager) postRequest(route string, queryParams QueryParameters, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

// Wrapper to send request to HAProxy Server
func (s Manager) putRequest(route string, queryParams QueryParameters, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = s.URI() + route + queryParamsToString(queryParams)
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

// Upload SSL certificate to HAProxy Server
func (s Manager) uploadSSL(route string, domain string, file io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = s.URI() + route

	// Prepare body
	body := &bytes.Buffer{}
	// Add file
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file_upload", filepath.Base(domain))
	if err != nil {
		return nil, errors.New("error creating field file")
	}
	// Copy file to body
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, errors.New("error copying file to body")
	}

	// Close writer
	err = writer.Close()
	if err != nil {
		return nil, errors.New("error closing writer")
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, errors.New("error creating request")
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	return client.Do(req)
}

// Replace SSL certificate to HAProxy Server
func (s Manager) replaceSSL(route string, domain string, file io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = s.URI() + route

	// Prepare body
	body := bytes.Buffer{}
	// Add file
	_, err := io.Copy(&body, file)
	if err != nil {
		return nil, errors.New("error copying file to body")
	}

	req, err := http.NewRequest("PUT", url, &body)
	if err != nil {
		return nil, errors.New("error creating request")
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Add("Content-Type", "text/plain")
	client := &http.Client{}
	return client.Do(req)
}
