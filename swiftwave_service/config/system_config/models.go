package system_config

import "github.com/lib/pq"

// ************************************************************************************* //
//                                Swiftwave System Configuration 		   			     //
// ************************************************************************************* //

// SystemConfig : hold information about system configuration
type SystemConfig struct {
	ID                           uint                         `json:"id" gorm:"primaryKey"`
	ConfigVersion                uint                         `json:"config_version" gore:"default:1"`
	SWVersion                    string                       `json:"sw_version"`
	SetupCompleted               bool                         `json:"setup_completed" gorm:"default:false"`
	ConfigHash                   string                       `json:"config_hash"`
	NetworkName                  string                       `json:"network_name"`
	RestrictedPorts              pq.Int64Array                `json:"restricted_ports" gorm:"type:integer[]"`
	JWTSecretKey                 string                       `json:"jwt_secret_key"`
	SshPrivateKey                string                       `json:"ssh_private_key"`
	FirewallConfig               FirewallConfig               `json:"firewall_config" gorm:"embedded;embeddedPrefix:firewall_config_"`
	LetsEncryptConfig            LetsEncryptConfig            `json:"lets_encrypt_config" gorm:"embedded;embeddedPrefix:lets_encrypt_config_"`
	HAProxyConfig                HAProxyConfig                `json:"haproxy_config" gorm:"embedded;embeddedPrefix:haproxy_config_"`
	UDPProxyConfig               UDPProxyConfig               `json:"udp_proxy_config" gorm:"embedded;embeddedPrefix:udp_proxy_config_"`
	PersistentVolumeBackupConfig PersistentVolumeBackupConfig `json:"persistent_volume_backup_config" gorm:"embedded;embeddedPrefix:persistent_volume_backup_config_"`
	PubSubConfig                 PubSubConfig                 `json:"pub_sub_config" gorm:"embedded;embeddedPrefix:pub_sub_config_"`
	TaskQueueConfig              TaskQueueConfig              `json:"task_queue_config" gorm:"embedded;embeddedPrefix:task_queue_config_"`
	ImageRegistryConfig          ImageRegistryConfig          `json:"image_registry_config" gorm:"embedded;embeddedPrefix:image_registry_config_"`
}
