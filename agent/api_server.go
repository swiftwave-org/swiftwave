package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func startHttpServer() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	// Auth middleware
	// e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	// 	return func(c echo.Context) error {
	// 		token := c.Request().Header.Get("Authorization")
	// 		if token != "Token" {
	// 			return c.JSON(http.StatusUnauthorized, "Unauthorized")
	// 		}
	// 		return next(c)
	// 	}
	// })

	// Container API

	// Create + Remove + Status + Logs
	// Container is mutable
	// For update, swiftwave will remove previous container and create a new one

	/*
		For zero downtime deployment, we can use the following steps:
		1. Create a new container with the new image
		2. Wait for the new container to be ready
		3. Update the DNS record to point to the new container
		4. Remove the old DNS records
		5. Remove the old containers
	*/

	// Volume API
	e.GET("/volumes", fetchAllVolumes)
	e.GET("/volumes/:uuid", fetchVolume)
	e.POST("/volumes/:uuid/size", fetchVolumeSize)
	e.POST("/volumes", createVolume)
	e.DELETE("/volumes/:uuid", deleteVolume)

	// DNS Record API
	e.GET("/dns", fetchAllDNSRecords)
	e.GET("/dns/:domain", fetchDNSRecordsByDomain)
	e.DELETE("/dns", deleteDNSRecord)
	e.POST("/dns", createDNSRecord)

	// Wireguard Peer API
	e.GET("/wireguard/peers", fetchAllWireguardPeers)
	e.POST("/wireguard/peers", createWireguardPeer)
	e.DELETE("/wireguard/peers/:publicKey", deleteWireguardPeer)
	e.PUT("/wireguard/peers/:publicKey", updateWireguardPeer)
	e.POST("/wireguard/peers/configure", configureWireguardPeers)

	// Static Route API
	e.GET("/static-routes", fetchAllStaticRoutes)
	e.POST("/static-routes", createStaticRoute)
	e.GET("/static-routes/:ip/:cidr", fetchStaticRouteByDestination)
	e.DELETE("/static-routes/:ip/:cidr", deleteStaticRoute)

	// NF Rule API
	e.GET("/nf-rules", fetchAllNFRules)
	e.POST("/nf-rules", createNFRule)
	e.GET("/nf-rules/:uuid", getNFRule)
	e.DELETE("/nf-rules/:uuid", deleteNFRule)

	// HAProxy API
	e.GET("/haproxy/service-status", getHAProxyStatus)
	e.Any("/haproxy/*", sendRequestToHAProxy)

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		e.Logger.Fatal(err)
	}
}
