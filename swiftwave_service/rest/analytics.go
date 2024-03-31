package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"io"
	"log"
	"net/http"
	"time"
)

// POST /service/analytics
func (server *Server) analytics(c echo.Context) error {
	if c.Get("hostname") == nil {
		return c.String(http.StatusBadRequest, "invalid request")
	}
	// fetch hostname from context
	serverHostName := c.Get("hostname").(string)
	// parse request body
	var buf bytes.Buffer
	// Copy the response body to the buffer
	_, err := io.Copy(&buf, c.Request().Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return c.String(http.StatusBadRequest, "invalid request")
	}
	if err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusBadRequest, "invalid request")
	}
	var data ResourceStatsData
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusBadRequest, "invalid request")
	}
	// create a transaction
	tx := server.ServiceManager.DbClient.Begin()
	defer func() {
		tx.Rollback()
	}()

	// fetch server id from database
	serverId, err := core.FetchServerIDByHostName(tx, serverHostName)
	if err != nil {
		log.Println(err.Error())
		return c.String(http.StatusInternalServerError, "failed to fetch server id")
	}
	// create new host resource stat
	diskStats := make([]core.ServerDiskStat, len(data.SystemStat.DiskStats))
	for i, diskStat := range data.SystemStat.DiskStats {
		diskStats[i] = core.ServerDiskStat{
			Path:       diskStat.Path,
			MountPoint: diskStat.MountPoint,
			TotalGB:    diskStat.TotalGB,
			UsedGB:     diskStat.UsedGB,
		}
	}
	serverStat := core.ServerResourceStat{
		ServerID:        serverId,
		CpuUsagePercent: data.SystemStat.CpuUsagePercent,
		MemStat: core.ServerMemoryStat{
			TotalGB:  data.SystemStat.MemStat.TotalGB,
			UsedGB:   data.SystemStat.MemStat.UsedGB,
			CachedGB: data.SystemStat.MemStat.CachedGB,
		},
		DiskStats: diskStats,
		NetStat: core.ServerNetStat{
			RecvKB:   data.SystemStat.NetStat.RecvKB,
			SentKB:   data.SystemStat.NetStat.SentKB,
			RecvKBPS: data.SystemStat.NetStat.RecvKB / 60,
			SentKBPS: data.SystemStat.NetStat.SentKB / 60,
		},
		RecordedAt: time.Unix(int64(data.TimeStamp), 0),
	}
	err = serverStat.Create(c.Request().Context(), *tx)
	if err != nil {
		log.Println(err.Error())
		return c.String(http.StatusInternalServerError, "failed to create server resource stat")
	}

	// create application resource stat
	appStats := make([]*core.ApplicationServiceResourceStat, 0)
	for serviceName, serviceStat := range data.ServiceStats {
		application := core.Application{}
		err := application.FindByName(c.Request().Context(), *tx, serviceName)
		if err != nil {
			continue
		}
		appStats = append(appStats, &core.ApplicationServiceResourceStat{
			ApplicationID:        application.ID,
			CpuUsagePercent:      serviceStat.CpuUsagePercent,
			ReportingServerCount: 1,
			UsedMemoryMB:         serviceStat.UsedMemoryMB,
			NetStat: core.ApplicationServiceNetStat{
				RecvKB:   serviceStat.NetStat.RecvKB,
				SentKB:   serviceStat.NetStat.SentKB,
				RecvKBPS: serviceStat.NetStat.RecvKB / 60,
				SentKBPS: serviceStat.NetStat.SentKB / 60,
			},
			RecordedAt: time.Unix(int64(data.TimeStamp), 0),
		})
	}
	// create application resource stat
	err = core.CreateApplicationServiceResourceStat(c.Request().Context(), *tx, appStats)
	if err != nil {
		log.Println(err.Error())
		return c.String(http.StatusInternalServerError, "failed to create application resource stat")
	}
	// commit transaction
	tx.Commit()
	return c.String(200, "ok")
}
