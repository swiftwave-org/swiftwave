package rest

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
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
	// unmarshal request body
	var data ResourceStatsData
	if err := c.Bind(&data); err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusBadRequest, "invalid request")
	}
	// fetch server id from database
	serverId, err := core.FetchServerIDByHostName(&server.ServiceManager.DbClient, serverHostName)
	if err != nil {
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
	err = serverStat.Create(c.Request().Context(), server.ServiceManager.DbClient)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to create server resource stat")
	}
	return c.String(200, "ok")
}
