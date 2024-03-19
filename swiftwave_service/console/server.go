package console

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"golang.org/x/net/websocket"
	"net/http"
	"strconv"
)

// Initialize : Initialize the server and its routes
func (server *Server) Initialize() {
	server.initiateAssetRoutes()
	server.EchoServer.POST("/console/token/server/:id", server.generateAuthTokenForServer)
	server.EchoServer.POST("/console/token/application/:id", server.generateAuthTokenForApplication)
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
	// generate token
	token, err := core.GenerateConsoleTokenForServer(server.ServiceManager.DbClient, serverIdUint)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}
	// return request id and token
	resp := map[string]interface{}{
		"request_id": token.ID,
		"token":      token.Token,
	}
	return c.JSON(http.StatusOK, resp)
}

func (server *Server) generateAuthTokenForApplication(c echo.Context) error {
	return nil
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
