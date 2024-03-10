package core

import (
	"errors"
	"gorm.io/gorm"
)

// CreateServer creates a new server in the database
func CreateServer(db *gorm.DB, server *Server) error {
	if server.IP == "" {
		return errors.New("IP is required")
	}
	if server.User == "" {
		return errors.New("user is required")
	}
	return db.Create(server).Error
}

// UpdateServer updates a server in the database
func UpdateServer(db *gorm.DB, server *Server) error {
	return db.Save(server).Error
}

// IsPreparedServerExists checks if a prepared server exists in the database
func IsPreparedServerExists(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Model(&Server{}).Where("status = ?", ServerOnline).Or("status = ?", ServerOffline).Count(&count).Error
	return count > 0, err
}

// FetchAllServers fetches all servers from the database
func FetchAllServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Find(&servers).Error
	return servers, err
}

// FetchServerByID fetches a server by its ID from the database
func FetchServerByID(db *gorm.DB, id uint) (*Server, error) {
	var server Server
	err := db.First(&server, id).Error
	return &server, err
}

// FetchAllOnlineServers fetches all servers from the database
func FetchAllOnlineServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("status = ?", ServerOnline).Find(&servers).Error
	return servers, err
}

// FetchSwarmManager fetches the swarm manager from the database
func FetchSwarmManager(db *gorm.DB) (Server, error) {
	var server Server
	// The reason behind using Order("RANDOM()") is
	// if any swarm manager is down, the next one will be used
	// so remove the possibility of complete failure
	err := db.Where("status = ?", ServerOnline).Where("swarm_mode = ?", SwarmManager).Order("RANDOM()").First(&server).Error
	return server, err
}

// FetchProxyActiveServers fetches all active servers from the database
func FetchProxyActiveServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("status = ?", ServerOnline).Where("proxy_enabled = ?", true).Where("proxy_type = ?", ActiveProxy).Find(&servers).Error
	return servers, err
}

// FetchProxyBackupServers fetches all backup servers from the database
// nolint:unused
func FetchProxyBackupServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("status = ?", ServerOnline).Where("proxy_enabled = ?", true).Where("proxy_type = ?", BackupProxy).Find(&servers).Error
	return servers, err
}

// FetchAllProxyServers fetches all proxy servers from the database
func FetchAllProxyServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("status = ?", ServerOnline).Where("proxy_enabled = ?", true).Find(&servers).Error
	return servers, err
}
