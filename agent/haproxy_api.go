package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func sendRequestToHAProxy(c echo.Context) error {
	config, err := GetConfig()
	if err != nil {
		return c.String(http.StatusNotFound, "Failed to fetch config")
	}
	if !config.HaproxyConfig.Enabled {
		return c.String(http.StatusNotFound, "Haproxy is not enabled")
	}
	// // Get the original request path and strip "/haproxy"
	// newPath := strings.TrimPrefix(c.Request().URL.Path, "/haproxy")

	// Define the target URL
	targetURL, _ := url.Parse(DataplaneAPIBaseAddress)

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	// Modify request before forwarding
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/haproxy")
		req.Host = targetURL.Host

		// Set Basic Auth Header
		req.SetBasicAuth(config.HaproxyConfig.Username, config.HaproxyConfig.Password)
	}

	// Serve the request
	proxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

func getHAProxyStatus(c echo.Context) error {
	config, err := GetConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch config",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Successfully fetched status",
		Data: map[string]interface{}{
			"enabled":                     config.HaproxyConfig.Enabled,
			"haproxy_service_active":      GetServiceStatus("haproxy"),
			"dataplaneapi_service_active": GetServiceStatus("dataplaneapi"),
		},
	})
}
