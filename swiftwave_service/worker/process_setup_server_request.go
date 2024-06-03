package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"gorm.io/gorm"
	"time"
)

func (m Manager) SetupServer(request SetupServerRequest, ctx context.Context, _ context.CancelFunc) error {
	err := m.setupServerHelper(request, ctx, nil)
	// fetch server
	server, fetchServerErr := core.FetchServerByID(&m.ServiceManager.DbClient, request.ServerId)
	if fetchServerErr != nil {
		logger.CronJobLoggerError.Println("Failed to fetch server by id", request.ServerId)
		return nil
	}
	if err != nil {
		// update server status
		server.Status = core.ServerNeedsSetup
		_ = core.UpdateServer(&m.ServiceManager.DbClient, server)
	} else {
		// update server status
		server.Status = core.ServerOnline
		_ = core.UpdateServer(&m.ServiceManager.DbClient, server)
	}
	return nil
}

func (m Manager) setupServerHelper(request SetupServerRequest, ctx context.Context, _ context.CancelFunc) error {
	// fetch server
	server, err := core.FetchServerByID(&m.ServiceManager.DbClient, request.ServerId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// fetch server log
	serverLog, err := core.FetchServerLogByID(&m.ServiceManager.DbClient, request.LogId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// log
	logText := "Preparing server for deployment\n"
	// spawn a goroutine to update server log each 5 seconds
	go func() {
		lastSent := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if time.Since(lastSent) > 5*time.Second {
					serverLog.Content = logText
					_ = serverLog.Update(&m.ServiceManager.DbClient)
					lastSent = time.Now()
				}
			}
		}
	}()
	// defer to push final log
	defer func() {
		serverLog.Content = logText
		_ = serverLog.Update(&m.ServiceManager.DbClient)
	}()

	// Proceed request logic (reject in any other case)
	// Create all the required directories, use ssh exec with sudo
	directories := []string{
		m.Config.LocalConfig.ServiceConfig.DataDirectory,
		m.Config.LocalConfig.ServiceConfig.SocketPathDirectory,
		m.Config.LocalConfig.ServiceConfig.LocalPostgresDataDirectory,
		m.Config.LocalConfig.ServiceConfig.TarballDirectoryPath,
		m.Config.LocalConfig.ServiceConfig.LogDirectoryPath,
		m.Config.LocalConfig.ServiceConfig.PVBackupDirectoryPath,
		m.Config.LocalConfig.ServiceConfig.PVRestoreDirectoryPath,
		m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath,
		m.Config.LocalConfig.ServiceConfig.HAProxyUnixSocketDirectory,
		m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath,
		m.Config.LocalConfig.ServiceConfig.UDPProxyUnixSocketDirectory,
		m.Config.LocalConfig.ServiceConfig.SSLCertDirectoryPath,
		m.Config.LocalConfig.ServiceConfig.LocalImageRegistryDirectoryPath,
	}

	for _, dir := range directories {
		stdoutBuf := bytes.Buffer{}
		stderrBuf := bytes.Buffer{}
		err := ssh_toolkit.ExecCommandOverSSH(fmt.Sprintf("mkdir -p %s && chmod -R 0711 %s", dir, dir), &stdoutBuf, &stderrBuf, 5, server.IP, server.SSHPort, server.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logText += "Failed to create folder " + dir + "\n"
			logText += stdoutBuf.String() + "\n" + stderrBuf.String() + "\n"
			return err
		} else {
			logText += "Folder created > " + dir + "\n"
		}
	}

	// check docker socket
	conn, err := ssh_toolkit.NetConnOverSSH("unix", server.DockerUnixSocketPath, 5, server.IP, server.SSHPort, server.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logText += "Failed to connect to docker socket\n"
		logText += fmt.Sprintf("%s should have acess to %s\n", server.User, server.DockerUnixSocketPath)
		logText += err.Error() + "\n"
		return err
	}

	// create a docker client
	dockerClient, err := containermanger.New(ctx, conn)
	if err != nil {
		logText += "Failed to create docker client\n"
		logText += err.Error() + "\n"
		return err
	}

	defer func() {
		_ = dockerClient.Close()
	}()

	// Try to list volume [just to check if the docker client is working]
	_, err = dockerClient.FetchVolumes()
	if err != nil {
		logText += "Failed to connect to docker daemon\n"
		logText += err.Error() + "\n"
		return err
	} else {
		logText += "Docker client connected\n"
	}

	// Proceed request logic (reject in any other case)
	// - if, want to be manager
	//    - if, there are some managers already, need to be online any of them
	//    - if, no servers, then it will be the first manager
	// - if, want to be worker
	//   - there need to be at least one manager
	var swarmManagerServer *core.Server
	if server.SwarmMode == core.SwarmManager {
		// Check if there are some servers already
		exists, err := core.IsPreparedServerExists(&m.ServiceManager.DbClient)
		if err != nil {
			logText += "Failed to check if there are some servers already\n"
			logText += err.Error() + "\n"
			return err
		}
		if exists {
			// Try to find out if there is any manager online
			r, err := core.FetchSwarmManager(&m.ServiceManager.DbClient)
			if err != nil {
				logText += "Failed to find out if there is any swarm manager online\n"
				logText += err.Error() + "\n"
				return err
			}
			swarmManagerServer = &r
		}
	} else {
		// Check if there is any manager
		r, err := core.FetchSwarmManager(&m.ServiceManager.DbClient)
		if err != nil {
			logText += "Failed to find out if there is any swarm manager\n"
			logText += err.Error() + "\n"
			return err
		}
		swarmManagerServer = &r
	}

	if swarmManagerServer == nil && server.SwarmMode == core.SwarmWorker {
		logText += "No manager found\n"
		logText += "At least one active swarm manager is required in cluster to add a worker\n"
		return err
	}

	// NOTE: From here, if `swarmManagerServer` is nil, then this new server can be initialized as first swarm manager
	if swarmManagerServer == nil {
		// Initialize as first swarm manager
		err = dockerClient.InitializeAsManager(request.AdvertiseIP)
		if err != nil {
			logText += "Failed to initialize as first swarm manager\n"
			logText += err.Error() + "\n"
			return err
		}
	} else {
		// Get docker client of swarm manager
		swarmManagerConn, err := ssh_toolkit.NetConnOverSSH("unix", swarmManagerServer.DockerUnixSocketPath, 5, swarmManagerServer.IP, swarmManagerServer.SSHPort, swarmManagerServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logText += "Failed to connect to swarm manager\n"
			logText += err.Error() + "\n"
			return err
		}
		swarmManagerDockerClient, err := containermanger.New(ctx, swarmManagerConn)
		if err != nil {
			logText += "Failed to create docker client for swarm manager\n"
			logText += err.Error() + "\n"
			return err
		}
		// Fetch cluster join token from swarm manager
		var joinToken string
		if server.SwarmMode == core.SwarmManager {
			token, err := swarmManagerDockerClient.GenerateManagerJoinToken()
			if err != nil {
				logText += "Failed to generate manager join token\n"
				logText += err.Error() + "\n"
				return err
			}
			joinToken = token
		} else {
			token, err := swarmManagerDockerClient.GenerateWorkerJoinToken()
			if err != nil {
				logText += "Failed to generate worker join token\n"
				logText += err.Error() + "\n"
				return err
			}
			joinToken = token
		}
		// Add node to swarm cluster
		err = dockerClient.JoinSwarm(fmt.Sprintf("%s:2377", swarmManagerServer.IP), joinToken, request.AdvertiseIP)
		if err != nil {
			logText += "Failed to join swarm cluster\n"
			logText += err.Error() + "\n"
			return err
		}
	}

	// create all the volume in the server
	pvVolumes, err := core.FindAllPersistentVolumes(ctx, m.ServiceManager.DbClient)
	if err != nil {
		logText += "Failed to find all persistent volumes\nUser may need to create them manually\n"
		logText += err.Error() + "\n"
	}
	for _, persistentVolume := range pvVolumes {
		// remove volume (try)
		_ = dockerClient.RemoveVolume(persistentVolume.Name)
		// create volume
		var err error
		if persistentVolume.Type == core.PersistentVolumeTypeLocal {
			err = dockerClient.CreateLocalVolume(persistentVolume.Name)
		} else if persistentVolume.Type == core.PersistentVolumeTypeNFS {
			err = dockerClient.CreateNFSVolume(persistentVolume.Name, persistentVolume.NFSConfig.Host, persistentVolume.NFSConfig.Path, persistentVolume.NFSConfig.Version)
		} else if persistentVolume.Type == core.PersistentVolumeTypeCIFS {
			err = dockerClient.CreateCIFSVolume(persistentVolume.Name, persistentVolume.CIFSConfig.Host, persistentVolume.CIFSConfig.Share, persistentVolume.CIFSConfig.Username, persistentVolume.CIFSConfig.Password, persistentVolume.CIFSConfig.FileMode, persistentVolume.CIFSConfig.DirMode, persistentVolume.CIFSConfig.Uid, persistentVolume.CIFSConfig.Gid)
		}
		if err != nil {
			logText += "Failed to add persistent volume " + persistentVolume.Name + "\n"
			logText += err.Error() + "\n"
		} else {
			logText += "Persistent volume " + persistentVolume.Name + " added successfully\n"
		}
	}

	// set log
	logText += "Server is ready for deployment\n"
	return nil
}
