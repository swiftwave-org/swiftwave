package bootstrap

import (
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"math/rand"
)

func IsSystemSetupRequired() (bool, error) {
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

func payloadToDBRecord(payload SystemConfigurationPayload) system_config.SystemConfig {
	imageRegistryConfig := system_config.ImageRegistryConfig{}
	if payload.ImageRegistry.Type == RemoteRegistry {
		imageRegistryConfig = system_config.ImageRegistryConfig{
			Endpoint:  payload.ImageRegistry.Endpoint,
			Namespace: payload.ImageRegistry.Namespace,
			Username:  payload.ImageRegistry.Username,
			Password:  payload.ImageRegistry.Password,
		}
	}
	return system_config.SystemConfig{
		NetworkName:   payload.NetworkName,
		ConfigVersion: 1,
		JWTSecretKey:  generateRandomString(64),
		SshPrivateKey: "",
		//ExtraRestrictedPorts: payload.ExtraRestrictedPorts,
		LetsEncryptConfig: system_config.LetsEncryptConfig{
			EmailID:    payload.LetsEncrypt.EmailAddress,
			Staging:    payload.LetsEncrypt.StagingEnv,
			PrivateKey: "",
		},
		HAProxyConfig: system_config.HAProxyConfig{
			Image: payload.HAProxyConfig.Image,
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
	}
}
