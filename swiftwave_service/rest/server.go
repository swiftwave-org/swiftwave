package rest

import "github.com/labstack/echo/v4"

// Initialize : Initialize the server and its routes
func (server *Server) Initialize() {
	// Initiating Routes for ACME Challenge
	server.ServiceManager.SslManager.InitHttpHandlers(server.EchoServer)
	// Initiating Routes for Project
	server.initiateProjectRoutes(server.EchoServer)
}

func (server *Server) initiateProjectRoutes(e *echo.Echo) {
	// Initiating Routes for Healthcheck
	e.GET("/healthcheck", server.healthcheck)
	// Initiating Routes for Auth
	e.POST("/auth/login", server.login)
	e.GET("/verify-auth", server.verifyAuth)
	// Initiating Routes for Project
	e.POST("/upload/code", server.uploadTarFile)
	// Initiating Routes for PersistentVolume
	e.GET("/persistent-volume/backup/:id/download", server.downloadPersistentVolumeBackup)
	e.POST("/persistent-volume/restore/:id/upload", server.uploadPersistentVolumeRestoreFile)
	// Initiating Routes for Webhook
	e.Any("/webhook/redeploy-app/:app-id/:webhook-token", server.redeployApp)
}
