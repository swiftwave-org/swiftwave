package console

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"golang.org/x/net/websocket"
	"net/http"
	"strconv"
)

// Initialize : Initialize the server and its routes
func (server *Server) Initialize() {
	server.initiateAssetRoutes()
	server.EchoServer.POST("/console/token/server/:id", server.generateAuthTokenForServer)
	server.EchoServer.POST("/console/token/application/:id/:server_id", server.generateAuthTokenForApplication)
	server.EchoServer.GET("/console/application/:id/servers", server.fetchServersForApplication)
	server.EchoServer.GET("/console/ws/:requestId/:token/:rows/:cols", server.consoleWebsocket)
}

// Handler for generate auth token
func (server *Server) generateAuthTokenForServer(c echo.Context) error {
	serverIdStr := c.Param("id")
	serverId, err := strconv.Atoi(serverIdStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid server id")
	}
	// try to convert to uint
	serverIdUint := uint(serverId)
	// fetch server
	serverRecord, err := core.FetchServerByID(&server.ServiceManager.DbClient, serverIdUint)
	if err != nil {
		return c.String(http.StatusNotFound, "Server not found")
	}
	// generate token
	token, err := core.GenerateConsoleTokenForServer(server.ServiceManager.DbClient, serverRecord.ID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}
	// return request id and token
	resp := map[string]interface{}{
		"request_id": token.ID,
		"token":      token.Token,
		"target": map[string]interface{}{
			"type":     "server",
			"ip":       serverRecord.IP,
			"user":     serverRecord.User,
			"hostname": serverRecord.HostName,
		},
	}
	return c.JSON(http.StatusOK, resp)
}

// Handler to fetch servers where application is deployed
// so that user can select a server to connect to console
func (server *Server) fetchServersForApplication(c echo.Context) error {
	// get application id
	applicationIdStr := c.Param("id")
	// fetch application
	applicationRecord := &core.Application{}
	var err error
	err = applicationRecord.FindById(c.Request().Context(), server.ServiceManager.DbClient, applicationIdStr)
	if err != nil {
		return c.String(http.StatusNotFound, "Application not found")
	}
	// fetch a swarm manager
	swarmManagerServer, err := core.FetchSwarmManager(&server.ServiceManager.DbClient)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to fetch swarm manager, maybe none online")
	}
	// fetch docker manager
	dockerManager, err := manager.DockerClient(c.Request().Context(), swarmManagerServer)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to connect to docker")
	}
	// fetch servers hostname
	serverHostnames, err := dockerManager.ServiceRunningServers(applicationRecord.Name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to fetch servers")
	}
	// create a map of servers <id, hostname>
	serverHostnamesMap := make(map[string]uint)
	for _, hostname := range serverHostnames {
		serverId, err := core.FetchServerIDByHostName(&server.ServiceManager.DbClient, hostname)
		if err != nil {
			continue
		}
		serverHostnamesMap[hostname] = serverId
	}
	return c.JSON(http.StatusOK, serverHostnamesMap)
}

// Handler for generate auth token for application
func (server *Server) generateAuthTokenForApplication(c echo.Context) error {
	// get application id
	applicationIdStr := c.Param("id")
	targetServerIdStr := c.Param("server_id")
	// try to convert to uint
	targetServerId, err := strconv.Atoi(targetServerIdStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid application id")
	}
	targetServerIdUint := uint(targetServerId)
	// fetch application
	applicationRecord := &core.Application{}
	err = applicationRecord.FindById(c.Request().Context(), server.ServiceManager.DbClient, applicationIdStr)
	if err != nil {
		return c.String(http.StatusNotFound, "Application not found")
	}
	// check if target server id is provided
	serverRecord, err := core.FetchServerByID(&server.ServiceManager.DbClient, targetServerIdUint)
	if err != nil {
		return c.String(http.StatusNotFound, "Server not found")
	}

	// generate token
	token, err := core.GenerateConsoleTokenForApplication(server.ServiceManager.DbClient, applicationRecord.ID, targetServerIdUint)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}
	// return request id and token
	resp := map[string]interface{}{
		"request_id": token.ID,
		"token":      token.Token,
		"target": map[string]interface{}{
			"type":        "application",
			"application": applicationRecord.Name,
			"server": map[string]interface{}{
				"ip":       serverRecord.IP,
				"user":     serverRecord.User,
				"hostname": serverRecord.HostName,
			},
		},
	}
	return c.JSON(http.StatusOK, resp)
}

// Websocket handler for console
func (server *Server) consoleWebsocket(c echo.Context) error {
	requestId := c.Param("requestId")
	token := c.Param("token")
	rowsStr := c.Param("rows")
	colsStr := c.Param("cols")
	rows, err := strconv.Atoi(rowsStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid rows")
	}
	cols, err := strconv.Atoi(colsStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid cols")
	}
	// find token
	tokenRecord, err := core.FindConsoleToken(server.ServiceManager.DbClient, requestId, token)
	if err != nil {
		return c.String(http.StatusForbidden, "Invalid token")
	}

	if tokenRecord.Target == core.ConsoleTargetTypeServer {
		return server.handleSSHConsoleRequestToServer(c, rows, cols, tokenRecord)
	} else if tokenRecord.Target == core.ConsoleTargetTypeApplication {
		return server.handleSSHConsoleRequestToApplication(c, rows, cols, tokenRecord)
	} else {
		return c.String(http.StatusNotImplemented, "Not implemented")
	}
}

func (server *Server) handleSSHConsoleRequestToServer(c echo.Context, rows int, cols int, tokenRecord *core.ConsoleToken) error {
	// fetch server
	remoteServer, err := core.FetchServerByID(&server.ServiceManager.DbClient, *tokenRecord.ServerID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to connect to server")
	}
	// create context with cancel
	ctx, ctxCancel := context.WithCancel(context.Background())
	// create ssh
	session, stdin, stdout, stderr, err := ssh_toolkit.DirectSSH(ctx, cols, rows, remoteServer.IP, 22, remoteServer.User, server.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		ctxCancel()
		return c.String(http.StatusInternalServerError, "Failed to connect to server")
	}
	if stdin == nil || stdout == nil || stderr == nil {
		ctxCancel()
		return c.String(http.StatusInternalServerError, "Failed to connect to server")
	}
	// accept websocket
	websocket.Handler(func(ws *websocket.Conn) {
		defer func(ws *websocket.Conn) {
			_ = ws.Close()
		}(ws)

		// accept websocket connection
		ws.PayloadType = websocket.BinaryFrame

		// write stdout to websocket
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					buf := make([]byte, 1024)
					n, err := (*stdout).Read(buf)
					if err != nil {
						ctxCancel()
					}
					err = websocket.Message.Send(ws, buf[:n])
					if err != nil {
						ctxCancel()
					}
				}

			}
		}()

		// write stderr to websocket
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					buf := make([]byte, 1024)
					n, err := (*stderr).Read(buf)
					if err != nil {
						ctxCancel()
						return
					}
					err = websocket.Message.Send(ws, buf[:n])
					if err != nil {
						ctxCancel()
					}
				}
			}
		}()

		// read until close in a go routine
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					var buf = make([]byte, 1024)
					err := websocket.Message.Receive(ws, &buf)
					if err != nil {
						ctxCancel()
						return
					}
					// check if starts with EOT (hacky way to get resize info from binary message)
					isAnySysReq := false
					if string(buf[:1]) == "\x04" {
						// take other part of the buffer
						buf = buf[1:]
						dimension := PTYDimension{}
						// marsha to json
						err = json.Unmarshal(buf, &dimension)
						if err == nil {
							isAnySysReq = true
							// resize terminal
							_ = session.WindowChange(dimension.Rows, dimension.Cols)
						}
					}
					if !isAnySysReq {
						_, err = (*stdin).Write(buf)
						if err != nil {
							ctxCancel()
							return
						}
					}
				}
			}
		}()
		// write until close
		<-ctx.Done()
	}).ServeHTTP(c.Response(), c.Request())
	// upgrade to websocket
	return nil
}

func (server *Server) handleSSHConsoleRequestToApplication(c echo.Context, rows int, cols int, tokenRecord *core.ConsoleToken) error {
	// fetch server
	remoteServer, err := core.FetchServerByID(&server.ServiceManager.DbClient, *tokenRecord.ServerID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to connect to server")
	}
	// fetch application
	applicationRecord := &core.Application{}
	err = applicationRecord.FindById(c.Request().Context(), server.ServiceManager.DbClient, *tokenRecord.ApplicationID)
	if err != nil {
		return c.String(http.StatusNotFound, "Application not found")
	}
	// create context with cancel
	ctx, ctxCancel := context.WithCancel(context.Background())
	// create docker manager
	dockerManager, err := manager.DockerClient(ctx, *remoteServer)
	if err != nil {
		ctxCancel()
		return c.String(http.StatusInternalServerError, "Failed to connect to docker")
	}
	// create exec id
	containerId, err := dockerManager.RandomServiceContainerID(applicationRecord.Name)
	if err != nil {
		ctxCancel()
		return c.String(http.StatusInternalServerError, "Failed to connect to server")
	}
	// create ssh
	dockerHost := fmt.Sprintf("unix://%s", remoteServer.DockerUnixSocketPath)
	session, stdin, stdout, stderr, err := ssh_toolkit.DirectSSHToContainer(ctx, cols, rows, containerId, dockerHost, remoteServer.IP, 22, remoteServer.User, server.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		ctxCancel()
		return c.String(http.StatusInternalServerError, "Failed to connect to server")
	}
	if stdin == nil || stdout == nil || stderr == nil {
		ctxCancel()
		return c.String(http.StatusInternalServerError, "Failed to connect to server")
	}
	// accept websocket
	websocket.Handler(func(ws *websocket.Conn) {
		defer func(ws *websocket.Conn) {
			_ = ws.Close()
		}(ws)

		// accept websocket connection
		ws.PayloadType = websocket.BinaryFrame

		// write stdout to websocket
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					buf := make([]byte, 1024)
					n, err := (*stdout).Read(buf)
					if err != nil {
						ctxCancel()
					}
					err = websocket.Message.Send(ws, buf[:n])
					if err != nil {
						ctxCancel()
					}
				}

			}
		}()

		// write stderr to websocket
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					buf := make([]byte, 1024)
					n, err := (*stderr).Read(buf)
					if err != nil {
						ctxCancel()
						return
					}
					err = websocket.Message.Send(ws, buf[:n])
					if err != nil {
						ctxCancel()
					}
				}
			}
		}()

		// read until close in a go routine
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					var buf = make([]byte, 1024)
					err := websocket.Message.Receive(ws, &buf)
					if err != nil {
						ctxCancel()
						return
					}
					// check if starts with EOT (hacky way to get resize info from binary message)
					isAnySysReq := false
					if string(buf[:1]) == "\x04" {
						// take other part of the buffer
						buf = buf[1:]
						dimension := PTYDimension{}
						// marsha to json
						err = json.Unmarshal(buf, &dimension)
						if err == nil {
							isAnySysReq = true
							// resize terminal
							_ = session.WindowChange(dimension.Rows, dimension.Cols)
						}
					}
					if !isAnySysReq {
						_, err = (*stdin).Write(buf)
						if err != nil {
							ctxCancel()
							return
						}
					}
				}
			}
		}()
		// write until close
		<-ctx.Done()
	}).ServeHTTP(c.Response(), c.Request())
	// upgrade to websocket
	return nil
}

//go:embed assets/index.html
var consoleHTML string

//go:embed assets/main.js
var consoleJS string

//go:embed assets/xterm.js
var xtermJS string

//go:embed assets/xterm.css
var xtermCSS string

//go:embed assets/xterm-addon-fit.js
var xtermAddonFit string

// Initialize Routes for assets
func (server *Server) initiateAssetRoutes() {
	server.EchoServer.GET("/console", func(c echo.Context) error {
		return c.HTML(200, consoleHTML)
	})
	server.EchoServer.GET("/console/main.js", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/javascript")
		return c.String(200, consoleJS)
	})
	server.EchoServer.GET("/console/xterm.js", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/javascript")
		return c.String(200, xtermJS)
	})
	server.EchoServer.GET("/console/xterm-addon-fit.js", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/javascript")
		return c.String(200, xtermAddonFit)
	})
	server.EchoServer.GET("/console/xterm.css", func(c echo.Context) error {
		// set content type
		c.Response().Header().Set("Content-Type", "text/css")
		return c.String(200, xtermCSS)
	})
}
