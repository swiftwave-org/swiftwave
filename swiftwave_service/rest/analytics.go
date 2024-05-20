package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
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
		logger.HTTPLogger.Println("Error reading response body:", err.Error())
		return c.String(http.StatusBadRequest, "invalid request")
	}
	var data ResourceStatsData
	requestBytes := buf.Bytes()
	if err := json.Unmarshal(requestBytes, &data); err != nil {
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
		logger.HTTPLogger.Println(err.Error())
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

	recvKB := data.SystemStat.NetStat.RecvKB
	sentKB := data.SystemStat.NetStat.SentKB

	/*
		Little hack -
		sometimes wrong data can be reported, due to overflow issues
		10000000000000000KB = 10000000000GB

		We are assuming that in 1 minute a server can't have 10000000000GB of data transfer in Tx/Rx.
		So if something reported, ignore that data.
	*/

	if recvKB > 10000000000000000 || sentKB > 10000000000000000 {
		logger.HTTPLoggerError.Println("Ignoring data, because anomaly detected in net stats")
		logger.HTTPLoggerError.Println(string(requestBytes))
		return c.String(http.StatusBadRequest, "invalid request")
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
		cpuUsagePercent := (serviceStat.ServiceCpuTime / serviceStat.SystemCpuTime) * 100
		appStats = append(appStats, &core.ApplicationServiceResourceStat{
			ApplicationID:        application.ID,
			ServiceCpuTime:       serviceStat.ServiceCpuTime,
			SystemCpuTime:        serviceStat.SystemCpuTime,
			CpuUsagePercent:      uint8(cpuUsagePercent),
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
