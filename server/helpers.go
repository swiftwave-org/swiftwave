package server

import "strings"

// Migrate database table
func (server *Server) MigrateDatabaseTables() {
}

// Function to check if the server is running in production environment
func (s *Server) isProductionEnvironment() bool {
	return strings.Compare(s.ENVIRONMENT, "production") == 0
}
