package graphql

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	swiftwaveMiddleware "github.com/swiftwave-org/swiftwave/swiftwave_service/middleware"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	containermanger "github.com/swiftwave-org/swiftwave/pkg/container_manager"
	dockerconfiggenerator "github.com/swiftwave-org/swiftwave/pkg/docker_config_generator"
	haproxymanager "github.com/swiftwave-org/swiftwave/pkg/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
)

func convertMapToDockerConfigBuildArgs(input map[string]dockerconfiggenerator.Variable) []*model.DockerConfigBuildArg {
	var output = make([]*model.DockerConfigBuildArg, 0)
	for key, value := range input {
		output = append(output, &model.DockerConfigBuildArg{
			Key:          key,
			Type:         value.Type,
			Description:  value.Description,
			DefaultValue: value.Default,
		})
	}
	return output
}

/*
SanitizeFileName Sanitize the fileName to remove potentially dangerous characters
It's meant to be used for filename
Should not use this to sanitize file path
*/
func sanitizeFileName(fileName string) string {
	// Remove any path components and keep only the file name
	fileName = filepath.Base(fileName)

	// Remove potentially dangerous characters like ".."
	fileName = strings.ReplaceAll(fileName, "..", "")

	// Remove potentially dangerous characters like "/"
	fileName = strings.ReplaceAll(fileName, "/", "")

	// You can add more sanitization rules as needed

	return fileName
}

func FetchDockerManager(ctx context.Context, db *gorm.DB) (*containermanger.Manager, error) {
	// Fetch a random swarm manager
	swarmManagerServer, err := core.FetchSwarmManager(db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// no online swarm manager
			logger.GraphQLLogger.Println("failed to fetch docker manager due to no online swarm manager")
			return nil, errors.New("failed to fetch docker manager due to no online swarm manager")
		}
		return nil, errors.New("failed to fetch swarm manager")
	}
	// Fetch docker manager
	dockerManager, err := manager.DockerClient(ctx, swarmManagerServer)
	if err != nil {
		return nil, errors.New("failed to fetch docker manager")
	}
	return dockerManager, nil
}

func AppendPublicSSHKeyLocally(pubKey string) error {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	// add \n to the end of the public key
	pubKey = pubKey + "\n"

	// Construct the path to the .ssh directory
	sshDirPath := filepath.Join(currentUser.HomeDir, ".ssh")

	// Create the .ssh directory if it doesn't exist
	err = os.MkdirAll(sshDirPath, 0700)
	if err != nil {
		return fmt.Errorf("failed to create .ssh directory: %v", err)
	}

	// Construct the path to the authorized_keys file
	authorizedKeysPath := filepath.Join(sshDirPath, "authorized_keys")

	// Open the authorized_keys file for appending
	f, err := os.OpenFile(authorizedKeysPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys file: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("failed to close authorized_keys file: %v", err)
		}
	}(f)

	// Append the public key to the file
	_, err = fmt.Fprintln(f, pubKey)
	if err != nil {
		return fmt.Errorf("failed to append public key to authorized_keys: %v", err)
	}

	return nil
}

func (r *mutationResolver) RunActionsInAllHAProxyNodes(ctx context.Context, db *gorm.DB, innerFunction func(ctx context.Context, db *gorm.DB, transactionId string, manager *haproxymanager.Manager) error) error {
	// fetch all proxy servers
	proxyServers, err := core.FetchProxyActiveServers(&r.ServiceManager.DbClient)
	if err != nil {
		return err
	}
	// don't attempt if no proxy servers are active
	if len(proxyServers) == 0 {
		return errors.New("no proxy servers are active")
	}
	// fetch all haproxy managers
	var haproxyManagers []*haproxymanager.Manager
	haproxyManagers, err = manager.HAProxyClients(context.Background(), proxyServers)
	if err != nil {
		return err
	}
	// map of server ip and transaction id
	transactionIdMap := make(map[*haproxymanager.Manager]string)
	var isFailed bool

	errString := ""

	for _, haproxyManager := range haproxyManagers {
		// create new haproxy transaction
		haproxyTransactionId, err := haproxyManager.FetchNewTransactionId()
		if err != nil {
			isFailed = true
			break
		}
		// add to map
		transactionIdMap[haproxyManager] = haproxyTransactionId
		// run the inner function
		err = innerFunction(ctx, db, haproxyTransactionId, haproxyManager)
		if err != nil {
			errString += err.Error() + "\n"
			isFailed = true
			break
		}
	}

	for haproxyManager, haproxyTransactionId := range transactionIdMap {
		var err error
		if !isFailed {
			err = haproxyManager.CommitTransaction(haproxyTransactionId)
		}
		if isFailed || err != nil {
			isFailed = true
			err2 := haproxyManager.DeleteTransaction(haproxyTransactionId)
			if err2 != nil {
				errString += err2.Error() + "\n"
			}
		}
	}

	if isFailed {
		if strings.Compare(errString, "") != 0 {
			return errors.New("failed to run actions in all haproxy nodes: " + errString)
		}
		return errors.New("failed to run actions in all haproxy nodes due to unknown error")
	} else {
		return nil
	}
}

func GetEchoContext(ctx context.Context) (echo.Context, error) {
	if c, ok := ctx.Value("echoContext").(echo.Context); ok {
		return c, nil
	} else {
		return nil, errors.New("failed to get echo context")
	}
}

func GetAuthInfo(ctx context.Context) swiftwaveMiddleware.AuthInfo {
	eCtx, err := GetEchoContext(ctx)
	if err != nil {
		return swiftwaveMiddleware.AuthInfo{}
	}
	if m, ok := eCtx.Get("auth").(swiftwaveMiddleware.AuthInfo); ok {
		return m
	} else {
		return swiftwaveMiddleware.AuthInfo{}
	}
}
