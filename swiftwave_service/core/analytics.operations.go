package core

import (
	"context"
	"gorm.io/gorm"
)

func (s *ServerResourceStat) Create(_ context.Context, db gorm.DB) error {
	return db.Create(s).Error
}

func CreateApplicationServiceResourceStat(_ context.Context, db gorm.DB, appStats []*ApplicationServiceResourceStat) error {
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
			totalPercent := uint(existingAppStat.CpuUsagePercent) * existingAppStat.ReportingServerCount
			totalPercent += uint(appStat.CpuUsagePercent) + totalPercent
			existingAppStat.CpuUsagePercent = uint8(totalPercent / (existingAppStat.ReportingServerCount + 1))
			existingAppStat.ReportingServerCount++
			existingAppStat.UsedMemoryMB = appStat.UsedMemoryMB + existingAppStat.UsedMemoryMB
			existingAppStat.NetStat = ApplicationServiceNetStat{
				RecvKB:   appStat.NetStat.RecvKB + existingAppStat.NetStat.RecvKB,
				SentKB:   appStat.NetStat.SentKB + existingAppStat.NetStat.SentKB,
				RecvKBPS: (appStat.NetStat.RecvKB + existingAppStat.NetStat.RecvKB) / 60,
				SentKBPS: (appStat.NetStat.SentKB + existingAppStat.NetStat.SentKB) / 60,
			}
			err = db.Save(&existingAppStat).Error
			if err != nil {
				return err
			}
		}
	}

	return db.Create(appStats).Error
}
