package bootstrap

import (
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"encoding/pem"
	"errors"
	"github.com/lib/pq"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"math/rand"
	"strconv"
	"strings"
)

func IsSystemSetupRequired() (bool, error) {
	if err := loadConfig(); err != nil {
		return false, err
	}
	// Create db connection
	dbConnection, err := db.GetClient(localConfig, 1)
	if err != nil {
		return false, errors.New("failed to connect to database")
	}
	// check if system is already setup
	var count int64
	if err := dbConnection.Model(&system_config.SystemConfig{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

func generateRandomString(length int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func portsStringToArray(ports string) []int64 {
	var portsMap = make(map[int64]bool)
	portsSplit := strings.Split(ports, ",")
	for _, port := range portsSplit {
		p := strings.TrimSpace(port)
		pInt, e := strconv.ParseInt(p, 10, 64)
		if e == nil {
			portsMap[pInt] = true
		}
	}
	// add default ports
	portsMap[22] = true
	portsMap[80] = true
	portsMap[443] = true
	portsMap[2376] = true
	portsMap[2377] = true
	portsMap[4789] = true
	portsMap[7946] = true
	// bind port
	if err := loadConfig(); err != nil {
		panic("You shouldn't call this function without having local config loaded !")
	}
	bindPort := localConfig.ServiceConfig.BindPort
	portsMap[int64(bindPort)] = true
	// convert map to array
	portsArr := pq.Int64Array{}
	for k := range portsMap {
		portsArr = append(portsArr, k)
	}
	return portsArr
}

func generateRSAPrivateKey() (string, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return "", errors.New("failed to generate RSA private key")
	}
	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	keyStr := string(privateKeyPEM)
	// add newline at the end of the key
	keyStr = keyStr + "\n"
	return keyStr, nil
}

func portsArrayToString(ports []int64) string {
	var portsStr = ""
	for _, port := range ports {
		portsStr = portsStr + strconv.FormatInt(port, 10) + ","
	}
	return portsStr
}

func generateRandomStringIfEmpty(s string, length int) string {
	if s == "" {
		return generateRandomString(length)
	}
	return s
}

func payloadToDBRecord(payload SystemConfigurationPayload) (system_config.SystemConfig, error) {
	// validations
	if isEmptyString(payload.NetworkName) {
		return system_config.SystemConfig{}, errors.New("network name is required")
	}
	if isEmptyString(payload.LetsEncrypt.EmailAddress) {
		return system_config.SystemConfig{}, errors.New("letsencrypt email address is required")
	}
	if isEmptyString(payload.HAProxyConfig.Image) {
		return system_config.SystemConfig{}, errors.New("haproxy image is required")
	}
	if isEmptyString(payload.UDPProxyConfig.Image) {
		return system_config.SystemConfig{}, errors.New("udp proxy image is required")
	}
	imageRegistryConfig := system_config.ImageRegistryConfig{}
	if payload.ImageRegistry.Type == RemoteRegistry {
		imageRegistryConfig = system_config.ImageRegistryConfig{
			Endpoint:  payload.ImageRegistry.Endpoint,
			Namespace: payload.ImageRegistry.Namespace,
			Username:  payload.ImageRegistry.Username,
			Password:  payload.ImageRegistry.Password,
		}
	}

	// generate ssh private key
	var sshPrivateKey string
	if isEmptyString(payload.SSHPrivateKey) {
		key, err := generateRSAPrivateKey()
		if err != nil {
			return system_config.SystemConfig{}, err
		}
		sshPrivateKey = key
	} else {
		sshPrivateKey = payload.SSHPrivateKey
	}
	// generate letsencrypt private key
	var letsEncryptPrivateKey string
	if isEmptyString(payload.LetsEncrypt.PrivateKey) {
		key, err := generateRSAPrivateKey()
		if err != nil {
			return system_config.SystemConfig{}, err
		}
		letsEncryptPrivateKey = key
	} else {
		letsEncryptPrivateKey = payload.LetsEncrypt.PrivateKey
	}

	return system_config.SystemConfig{
		NetworkName:     payload.NetworkName,
		ConfigVersion:   1,
		JWTSecretKey:    generateRandomStringIfEmpty(payload.JWTSecretKey, 32),
		SshPrivateKey:   sshPrivateKey,
		RestrictedPorts: portsStringToArray(payload.ExtraRestrictedPorts),
		LetsEncryptConfig: system_config.LetsEncryptConfig{
			EmailID:    payload.LetsEncrypt.EmailAddress,
			Staging:    payload.LetsEncrypt.StagingEnv,
			PrivateKey: letsEncryptPrivateKey,
		},
		HAProxyConfig: system_config.HAProxyConfig{
			Image:    payload.HAProxyConfig.Image,
			Username: generateRandomStringIfEmpty(payload.HAProxyConfig.Username, 16),
			Password: generateRandomStringIfEmpty(payload.HAProxyConfig.Password, 16),
		},
		UDPProxyConfig: system_config.UDPProxyConfig{
			Image: payload.UDPProxyConfig.Image,
		},
		PersistentVolumeBackupConfig: system_config.PersistentVolumeBackupConfig{
			S3BackupConfig: system_config.S3BackupConfig{
				Enabled:         payload.PvBackupConfig.S3Config.Enabled,
				Endpoint:        payload.PvBackupConfig.S3Config.Endpoint,
				Region:          payload.PvBackupConfig.S3Config.Region,
				Bucket:          payload.PvBackupConfig.S3Config.BucketName,
				AccessKeyID:     payload.PvBackupConfig.S3Config.AccessKeyId,
				SecretAccessKey: payload.PvBackupConfig.S3Config.SecretKey,
				ForcePathStyle:  payload.PvBackupConfig.S3Config.ForcePathStyle,
			},
		},
		PubSubConfig: system_config.PubSubConfig{
			Mode:         system_config.PubSubMode(payload.PubsubConfig.Type),
			BufferLength: payload.PubsubConfig.BufferLength,
			RedisConfig: system_config.RedisConfig{
				Host:       payload.PubsubConfig.RedisConfig.Host,
				Port:       payload.PubsubConfig.RedisConfig.Port,
				Password:   payload.PubsubConfig.RedisConfig.Password,
				DatabaseID: payload.PubsubConfig.RedisConfig.Database,
			},
		},
		TaskQueueConfig: system_config.TaskQueueConfig{
			Mode:                           system_config.TaskQueueMode(payload.TaskQueueConfig.Type),
			MaxOutstandingMessagesPerQueue: payload.TaskQueueConfig.MaxOutstandingMessagesPerQueue,
			NoOfWorkersPerQueue:            payload.TaskQueueConfig.NoOfWorkersPerQueue,
			AMQPConfig: system_config.AMQPConfig{
				Protocol: system_config.AMQPProtocol(payload.TaskQueueConfig.AmqpConfig.Protocol),
				Host:     payload.TaskQueueConfig.AmqpConfig.Host,
				Port:     payload.TaskQueueConfig.AmqpConfig.Port,
				User:     payload.TaskQueueConfig.AmqpConfig.Username,
				Password: payload.TaskQueueConfig.AmqpConfig.Password,
				VHost:    payload.TaskQueueConfig.AmqpConfig.Vhost,
			},
		},
		ImageRegistryConfig: imageRegistryConfig,
	}, nil
}

func dbRecordToPayload(record *system_config.SystemConfig) SystemConfigurationPayload {
	var imageRegistry = ImageRegistryConfig{
		Type: LocalRegistry,
	}
	if record.ImageRegistryConfig.IsConfigured() {
		imageRegistry = ImageRegistryConfig{
			Type:      RemoteRegistry,
			Endpoint:  record.ImageRegistryConfig.Endpoint,
			Namespace: record.ImageRegistryConfig.Namespace,
			Username:  record.ImageRegistryConfig.Username,
			Password:  record.ImageRegistryConfig.Password,
		}
	}
	var pubsubConfig = PubsubConfig{
		Type:         LocalPubsub,
		BufferLength: record.PubSubConfig.BufferLength,
	}
	if record.PubSubConfig.Mode == system_config.RemotePubSub {
		pubsubConfig = PubsubConfig{
			Type:         RemotePubsub,
			BufferLength: record.PubSubConfig.BufferLength,
			RedisConfig: RedisConfig{
				Host:     record.PubSubConfig.RedisConfig.Host,
				Port:     record.PubSubConfig.RedisConfig.Port,
				Password: record.PubSubConfig.RedisConfig.Password,
				Database: record.PubSubConfig.RedisConfig.DatabaseID,
			},
		}
	}
	var taskQueueConfig = TaskQueueConfig{
		Type:                           LocalTaskQueue,
		MaxOutstandingMessagesPerQueue: record.TaskQueueConfig.MaxOutstandingMessagesPerQueue,
		NoOfWorkersPerQueue:            record.TaskQueueConfig.NoOfWorkersPerQueue,
	}
	if record.TaskQueueConfig.Mode == system_config.RemoteTaskQueue {
		taskQueueConfig = TaskQueueConfig{
			Type:                           RemoteTaskQueue,
			MaxOutstandingMessagesPerQueue: record.TaskQueueConfig.MaxOutstandingMessagesPerQueue,
			NoOfWorkersPerQueue:            record.TaskQueueConfig.NoOfWorkersPerQueue,
			AmqpConfig: AmqpConfig{
				Protocol: TaskQueueQueueProtocol(record.TaskQueueConfig.AMQPConfig.Protocol),
				Host:     record.TaskQueueConfig.AMQPConfig.Host,
				Port:     record.TaskQueueConfig.AMQPConfig.Port,
				Username: record.TaskQueueConfig.AMQPConfig.User,
				Password: record.TaskQueueConfig.AMQPConfig.Password,
				Vhost:    record.TaskQueueConfig.AMQPConfig.VHost,
			},
		}
	}
	return SystemConfigurationPayload{
		NetworkName:          record.NetworkName,
		ExtraRestrictedPorts: portsArrayToString(record.RestrictedPorts),
		LetsEncrypt: LetsEncryptConfig{
			EmailAddress: record.LetsEncryptConfig.EmailID,
			StagingEnv:   record.LetsEncryptConfig.Staging,
		},
		ImageRegistry: imageRegistry,
		HAProxyConfig: HAProxyConfig{
			Image: record.HAProxyConfig.Image,
		},
		UDPProxyConfig: UDPProxyConfig{
			Image: record.UDPProxyConfig.Image,
		},
		PvBackupConfig: PvBackupConfig{
			S3Config: S3Config{
				Enabled:        record.PersistentVolumeBackupConfig.S3BackupConfig.Enabled,
				Endpoint:       record.PersistentVolumeBackupConfig.S3BackupConfig.Endpoint,
				Region:         record.PersistentVolumeBackupConfig.S3BackupConfig.Region,
				BucketName:     record.PersistentVolumeBackupConfig.S3BackupConfig.Bucket,
				AccessKeyId:    record.PersistentVolumeBackupConfig.S3BackupConfig.AccessKeyID,
				SecretKey:      record.PersistentVolumeBackupConfig.S3BackupConfig.SecretAccessKey,
				ForcePathStyle: record.PersistentVolumeBackupConfig.S3BackupConfig.ForcePathStyle,
			},
		},
		PubsubConfig:    pubsubConfig,
		TaskQueueConfig: taskQueueConfig,
	}
}

func isEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
