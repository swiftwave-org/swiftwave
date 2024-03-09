package system_config

// PubSubMode : mode of the pub-sub system
type PubSubMode string

const (
	LocalPubSub  PubSubMode = "local"
	RemotePubSub PubSubMode = "remote"
)

// AMQPProtocol : protocol for AMQP
type AMQPProtocol string

const (
	AMQP  AMQPProtocol = "amqp"
	AMQPS AMQPProtocol = "amqps"
)

// TaskQueueMode : mode of the task queue system
type TaskQueueMode string

const (
	LocalTaskQueue  TaskQueueMode = "local"
	RemoteTaskQueue TaskQueueMode = "remote"
)

// ImageRegistryConfig : configuration for image registry
type ImageRegistryConfig struct {
	Endpoint  string `json:"endpoint"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Namespace string `json:"namespace"`
}

// LetsEncryptConfig : hold information about lets encrypt configuration
type LetsEncryptConfig struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Staging    bool   `json:"staging" gorm:"default:false"`
	EmailID    string `json:"email_id"`
	PrivateKey string `json:"private_key"`
}

// FirewallConfig : hold information about firewall configuration
type FirewallConfig struct {
	Enabled          bool   `json:"enabled" gorm:"default:false"`
	AllowPortCommand string `json:"allow_port_command"` // can contain {{port}} as placeholder
	DenyPortCommand  string `json:"deny_port_command"`  // can contain {{port}} as placeholder
}

// PersistentVolumeBackupConfig : configuration for persistent volume backup
type PersistentVolumeBackupConfig struct {
	S3BackupConfig S3BackupConfig `json:"s3_backup_config" gorm:"embedded;embeddedPrefix:s3_backup_"`
}

// S3BackupConfig : configuration for S3 backup
type S3BackupConfig struct {
	Enabled         bool   `json:"enabled"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	ForcePathStyle  bool   `json:"force_path_style"`
}

// PubSubConfig : configuration for pub-sub system
type PubSubConfig struct {
	Mode         PubSubMode  `json:"mode" gorm:"default:'local'"`
	BufferLength uint        `json:"buffer_length" gorm:"default:2000"`
	RedisConfig  RedisConfig `json:"redis_config" gorm:"embedded;embeddedPrefix:redis_"`
}

// TaskQueueConfig : configuration for task queue system
type TaskQueueConfig struct {
	Mode                           TaskQueueMode `json:"mode" gorm:"default:'local'"`
	MaxOutstandingMessagesPerQueue uint          `json:"max_outstanding_messages_per_queue" gorm:"default:2"`
	NoOfWorkersPerQueue            uint          `json:"no_of_workers_per_queue"`
	AMQPConfig                     AMQPConfig    `json:"amqp_config" gorm:"embedded;embeddedPrefix:amqp_"`
}

// RedisConfig : configuration for Redis
type RedisConfig struct {
	Host       string `json:"host"`
	Port       uint   `json:"port"`
	Password   string `json:"password"`
	DatabaseID uint   `json:"database_id"`
}

// AMQPConfig : configuration for AMQP
type AMQPConfig struct {
	Protocol AMQPProtocol `json:"protocol"`
	Host     string       `json:"host"`
	User     string       `json:"user"`
	Password string       `json:"password"`
	VHost    string       `json:"vhost"`
}

// HAProxyConfig : configuration for HAProxy
type HAProxyConfig struct {
	Image          string `json:"image"`
	UnixSocketPath string `json:"unix_socket_path"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}

// UDPProxyConfig : configuration for UDP Proxy
type UDPProxyConfig struct {
	UnixSocketPath string `json:"unix_socket_path"`
	Image          string `json:"image"`
}
