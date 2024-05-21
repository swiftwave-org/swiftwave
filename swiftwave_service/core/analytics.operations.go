package core

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

func (s *ServerResourceStat) Create(_ context.Context, db gorm.DB) error {
	return db.Create(s).Error
}

func CreateApplicationServiceResourceStat(_ context.Context, db gorm.DB, appStats []*ApplicationServiceResourceStat) error {
	if len(appStats) == 0 {
		return nil
	}
	// Try to merge the updates if there is same record for same service at same timestamp
	for _, appStat := range appStats {
		var existingAppStat ApplicationServiceResourceStat
		err := db.Where("application_id = ? AND recorded_at = ?", appStat.ApplicationID, appStat.RecordedAt).First(&existingAppStat).Error
		if err != nil {
			// If no record found, then create record
			err = db.Create(&appStat).Error
			if err != nil {
				return err
			}
		} else {
			existingAppStat.ServiceCpuTime = appStat.ServiceCpuTime + existingAppStat.ServiceCpuTime
			existingAppStat.SystemCpuTime = appStat.SystemCpuTime + existingAppStat.SystemCpuTime
			existingAppStat.CpuUsagePercent = uint8(float64(existingAppStat.ServiceCpuTime) / float64(existingAppStat.SystemCpuTime) * 100)
			existingAppStat.ReportingServerCount++
			existingAppStat.UsedMemoryMB = appStat.UsedMemoryMB + existingAppStat.UsedMemoryMB
			existingAppStat.NetStat = ApplicationServiceNetStat{
				RecvKB:   appStat.NetStat.RecvKB + existingAppStat.NetStat.RecvKB,
				SentKB:   appStat.NetStat.SentKB + existingAppStat.NetStat.SentKB,
				RecvKBPS: uint64(appStat.NetStat.RecvKB+existingAppStat.NetStat.RecvKB) / 60,
				SentKBPS: uint64(appStat.NetStat.SentKB+existingAppStat.NetStat.SentKB) / 60,
			}
			err = db.Where("id = ?", existingAppStat.ID).Save(&existingAppStat).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FetchLatestServerResourceAnalytics fetches the latest server resource analytics
func FetchLatestServerResourceAnalytics(_ context.Context, db gorm.DB, serverId uint) (*ServerResourceStat, error) {
	var serverStat *ServerResourceStat
	err := db.Select("id", "server_id", "cpu_usage_percent",
		"memory_total_gb", "memory_used_gb", "memory_cached_gb",
		"network_sent_kbps", "network_recv_kbps", "recorded_at").Where("server_id = ?", serverId).Order("recorded_at desc").First(&serverStat).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ServerResourceStat{
				ID:              0,
				ServerID:        serverId,
				CpuUsagePercent: 0,
				DiskStats:       ServerDiskStats{},
				MemStat:         ServerMemoryStat{},
				NetStat:         ServerNetStat{},
				RecordedAt:      time.Now(),
			}, nil
		}
	}
	return serverStat, err
}

// FetchLatestServerDiskUsage fetches the latest server disk usage
func FetchLatestServerDiskUsage(_ context.Context, db gorm.DB, serverId uint) (*ServerDiskStats, *time.Time, error) {
	var serverStat *ServerResourceStat
	err := db.Select("id", "server_id", "disk_stats", "recorded_at").Where("server_id = ?", serverId).Order("recorded_at desc").First(&serverStat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			currentTime := time.Now()
			return &ServerDiskStats{}, &currentTime, nil
		}
		return nil, nil, err
	}
	return &serverStat.DiskStats, &serverStat.RecordedAt, err
}

// FetchServerDiskUsage fetches the server disk usage of last 24 hours
func FetchServerDiskUsage(_ context.Context, db gorm.DB, serverId uint) ([]*ServerResourceStat, error) {
	previousUnixTime := time.Now().Unix() - 86400
	var serverResourceStat []*ServerResourceStat
	previousTime := time.Unix(previousUnixTime, 0)
	err := db.Select("id", "server_id", "disk_stats", "recorded_at").Where("server_id = ?", serverId).Where("recorded_at > ?", previousTime).Order("recorded_at desc").Limit(1000).Find(&serverResourceStat).Error
	return serverResourceStat, err
}

// FetchServerResourceAnalytics fetches the server resource analytics
func FetchServerResourceAnalytics(_ context.Context, db gorm.DB, serverId uint, tillTime uint) ([]*ServerResourceStat, error) {
	var serverStats []*ServerResourceStat
	err := db.Select("id", "server_id", "cpu_usage_percent",
		"memory_total_gb", "memory_used_gb", "memory_cached_gb",
		"network_sent_kb", "network_recv_kb",
		"network_sent_kbps", "network_recv_kbps", "recorded_at").Where("server_id = ?", serverId).Where("recorded_at > ?", time.Unix(int64(tillTime), 0)).Order("recorded_at desc").Find(&serverStats).Error
	return serverStats, err
}

// FetchApplicationServiceResourceAnalytics fetches the application service resource analytics
func FetchApplicationServiceResourceAnalytics(_ context.Context, db gorm.DB, applicationId string, tillTime uint) ([]*ApplicationServiceResourceStat, error) {
	var appStats []*ApplicationServiceResourceStat
	err := db.Where("application_id = ?", applicationId).Where("recorded_at > ?", time.Unix(int64(tillTime), 0)).Order("recorded_at desc").Find(&appStats).Error
	return appStats, err
}
