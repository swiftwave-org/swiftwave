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
	// Initiating Routes for Auth
	e.POST("/auth/login", server.login)
	// Initiating Routes for Project
	e.POST("/upload/code", server.uploadTarFile)
}
