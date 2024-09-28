package core

import (
	"errors"
	"fmt"
	"net"
	"time"

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
	server.LastPing = time.Now()
	return db.Create(server).Error
}

// DeleteServer deletes a server from the database
func DeleteServer(db *gorm.DB, id uint) error {
	server, err := FetchServerByID(db, id)
	if err != nil {
		return err
	}
	var applications []Application

	tx := db.Raw("SELECT name FROM applications WHERE preferred_server_hostnames @> ARRAY[?]", server.HostName).Scan(&applications)
	if tx.Error != nil {
		return fmt.Errorf("failed to fetch linked apps : %s", tx.Error.Error())
	}
	if len(applications) > 0 {
		applicationString := ""
		for i, application := range applications {
			applicationString = applicationString + application.Name
			if i != len(applications)-1 {
				applicationString = applicationString + ", "
			} else {
				applicationString = applicationString + " "
			}
		}
		return fmt.Errorf("server is linked to application(s) : %s\nPlease remove this server from preferred servers of the application(s) before deleting the server", applicationString)
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

// NoOfServers returns the number of servers in the database
func NoOfServers(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Server{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// NoOfPreparedServers returns the number of prepared servers in the database
func NoOfPreparedServers(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Server{}).Where("status = ?", ServerOnline).Or("status = ?", ServerOffline).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("server not found")
	}
	return &server, err
}

// FetchServerByIP fetches a server by its IP from the database
func FetchServerByIP(db *gorm.DB, ip string) (*Server, error) {
	var server Server
	err := db.Where("ip = ?", ip).First(&server).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("server not found")
	}
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
	isAnyActiveProxyServerOffline, err := IsAnyActiveProxyServerOffline(db)
	if err != nil {
		return nil, err
	}
	if isAnyActiveProxyServerOffline {
		// PS: This has been done to mitigate security risks
		// Suppose we are adding authentication for an app to proxy
		// We failed to apply that on a specific proxy because the proxy was offline
		// But think that, the proxy become online after a while and it has no authentication for specific app
		// So, any user can access the app via that specific proxy
		// To mitigate this, we will abort the operation if any active proxy is offline
		return nil, errors.New("all proxy servers need to be online to perform this action")
	}
	var servers []Server
	err = db.Where("status = ?", ServerOnline).Where("proxy_enabled = ?", true).Where("proxy_type = ?", ActiveProxy).Find(&servers).Error
	return servers, err
}

// IsAnyActiveProxyServerOffline checks if any active proxy server is offline
func IsAnyActiveProxyServerOffline(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Model(&Server{}).Where("status = ?", ServerOffline).Where("proxy_enabled = ?", true).Where("proxy_type = ?", ActiveProxy).Count(&count).Error
	return count > 0, err
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

// FetchDisabledDeploymentServerHostNames fetches the hostnames of all servers that are not in deployment mode
func FetchDisabledDeploymentServerHostNames(db *gorm.DB) ([]string, error) {
	var disabledServers []Server
	err := db.Where("schedule_deployments = ?", false).Select("host_name").Distinct("host_name").Find(&disabledServers).Error
	if err != nil {
		return nil, err
	}
	var hostNames []string
	for _, server := range disabledServers {
		hostNames = append(hostNames, server.HostName)
	}
	return hostNames, nil
}
