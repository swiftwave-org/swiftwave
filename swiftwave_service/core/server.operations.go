package core

import (
	"errors"
	"gorm.io/gorm"
	"net"
	"time"
)

// CreateServer creates a new server in the database
func CreateServer(db *gorm.DB, server *Server) error {
	if server.IP == "" {
		return errors.New("IP is required")
	}
	if server.User == "" {
		return errors.New("user is required")
	}
	server.LastPing = time.Now()
	return db.Create(server).Error
}

// DeleteServer deletes a server from the database
func DeleteServer(db *gorm.DB, id uint) error {
	server, err := FetchServerByID(db, id)
	if err != nil {
		return err
	}
	return db.Delete(server).Error
}

// ChangeServerIP changes the IP of a server in the database
func ChangeServerIP(db *gorm.DB, server *Server, newIp string) error {
	if ip := net.ParseIP(newIp); ip == nil {
		return errors.New("invalid IP address")
	}
	return db.Model(server).Update("ip", newIp).Error
}

// ChangeSSHPort changes the SSH port of a server in the database
func ChangeSSHPort(db *gorm.DB, server *Server, newPort int) error {
	return db.Model(server).Update("ssh_port", newPort).Error
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

// FetchServerIDByHostName fetches a server by its hostname from the database
func FetchServerIDByHostName(db *gorm.DB, hostName string) (uint, error) {
	var server Server
	err := db.Select("id").Where("host_name = ?", hostName).First(&server).Error
	return server.ID, err
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

// FetchSwarmManagerExceptServer fetches the swarm manager from the database except the given server
func FetchSwarmManagerExceptServer(db *gorm.DB, serverId uint) (Server, error) {
	var swarmManager Server
	err := db.Where("status = ?", ServerOnline).Where("swarm_mode = ?", SwarmManager).Where("id != ?", serverId).Order("RANDOM()").First(&swarmManager).Error
	return swarmManager, err

}

// FetchProxyActiveServers fetches all active servers from the database
func FetchProxyActiveServers(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("status = ?", ServerOnline).Where("proxy_enabled = ?", true).Where("proxy_type = ?", ActiveProxy).Find(&servers).Error
	return servers, err
}

// FetchRandomActiveProxyServer fetches a random active server from the database
func FetchRandomActiveProxyServer(db *gorm.DB) (Server, error) {
	var server Server
	err := db.Where("status = ?", ServerOnline).Where("proxy_enabled = ?", true).Where("proxy_type = ?", ActiveProxy).Order("RANDOM()").First(&server).Error
	return server, err

}

// FetchBackupProxyServers fetches all backup servers from the database
func FetchBackupProxyServers(db *gorm.DB) ([]Server, error) {
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

// FetchAllProxyServersIrrespectiveOfStatus fetches all proxy servers from the database irrespective of status
func FetchAllProxyServersIrrespectiveOfStatus(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Where("proxy_enabled = ?", true).Find(&servers).Error
	return servers, err
}

// MarkServerAsOnline marks a server as online in the database
func MarkServerAsOnline(db *gorm.DB, server *Server) error {
	return db.Model(server).Updates(map[string]interface{}{"status": ServerOnline, "last_ping": time.Now()}).Error
}

// MarkServerAsOffline marks a server as offline in the database
func MarkServerAsOffline(db *gorm.DB, server *Server) error {
	return db.Model(server).Update("status", ServerOffline).Error
}

// ChangeProxyType changes the proxy type of server in the database
func ChangeProxyType(db *gorm.DB, server *Server, proxyType ProxyType) error {
	return db.Model(server).Update("proxy_type", proxyType).Error
}
