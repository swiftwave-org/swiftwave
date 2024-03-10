package core

import "gorm.io/gorm"

// FetchAllServers fetches all servers from the database
func FetchAllServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Find(&servers).Error
	return servers, err
}

// FetchSwarmManager fetches the swarm manager from the database
func FetchSwarmManager(db *gorm.DB) (Server, error) {
	var server Server
	// The reason behind using Order("RANDOM()") is
	// if any swarm manager is down, the next one will be used
	// so remove the possibility of complete failure
	err := db.Where("role = ?", SwarmManager).Order("RANDOM()").First(&server).Error
	return server, err
}

// FetchProxyActiveServers fetches all active servers from the database
func FetchProxyActiveServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("proxy_enabled = ?", true).Where("proxy_type = ?", ActiveProxy).Find(&servers).Error
	return servers, err
}

// FetchProxyBackupServers fetches all backup servers from the database
func FetchProxyBackupServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("proxy_enabled = ?", true).Where("proxy_type = ?", BackupProxy).Find(&servers).Error
	return servers, err
}
