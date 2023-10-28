package rest

// Initialize : Initialize the server and its routes
func (server *Server) Initialize() {
	// Initiating Routes for ACME Challenge
	server.ServiceManager.SslManager.InitHttpHandlers(server.EchoServer)
}
