package bootstrap

type ImageRegistryType string
type PubsubType string
type TaskQueueType string
type TaskQueueQueueProtocol string
type RemoteTaskQueueType string

const (
	LocalRegistry   ImageRegistryType      = "local"
	RemoteRegistry  ImageRegistryType      = "remote"
	LocalPubsub     PubsubType             = "local"
	RemotePubsub    PubsubType             = "remote"
	LocalTaskQueue  TaskQueueType          = "local"
	RemoteTaskQueue TaskQueueType          = "remote"
	AMQP            TaskQueueQueueProtocol = "amqp"
	AMQPS           TaskQueueQueueProtocol = "amqps"
	AmqpQueue       RemoteTaskQueueType    = "amqp"
	RedisQueue      RemoteTaskQueueType    = "redis"
	NoneRemoteQueue RemoteTaskQueueType    = "none"
)

type SystemConfigurationPayload struct {
	NetworkName          string              `json:"network_name"`
	ExtraRestrictedPorts string              `json:"extra_restricted_ports"`
	JWTSecretKey         string              `json:"-"`
	SSHPrivateKey        string              `json:"-"`
	LetsEncrypt          LetsEncryptConfig   `json:"lets_encrypt"`
	ImageRegistry        ImageRegistryConfig `json:"image_registry"`
	HAProxyConfig        HAProxyConfig       `json:"haproxy_config"`
	UDPProxyConfig       UDPProxyConfig      `json:"udpproxy_config"`
	PvBackupConfig       PvBackupConfig      `json:"pv_backup_config"`
	PubsubConfig         PubsubConfig        `json:"pubsub_config"`
	TaskQueueConfig      TaskQueueConfig     `json:"task_queue_config"`
	NewAdminCredential   NewAdminCredential  `json:"new_admin_credential"`
}

type LetsEncryptConfig struct {
	EmailAddress string `json:"email_address"`
	StagingEnv   bool   `json:"staging_env"`
	PrivateKey   string `json:"-"`
}

type ImageRegistryConfig struct {
	Type      ImageRegistryType `json:"type"`
	Endpoint  string            `json:"endpoint"`
	Namespace string            `json:"namespace"`
	Username  string            `json:"username"`
	Password  string            `json:"password"`
}

type HAProxyConfig struct {
	Image    string `json:"image"`
	Username string `json:"-"`
	Password string `json:"-"`
}

type UDPProxyConfig struct {
	Image string `json:"image"`
}

type PvBackupConfig struct {
	S3Config S3Config `json:"s3_config"`
}

type S3Config struct {
	Enabled        bool   `json:"enabled"`
	Endpoint       string `json:"endpoint"`
	Region         string `json:"region"`
	BucketName     string `json:"bucket_name"`
	AccessKeyId    string `json:"access_key_id"`
	SecretKey      string `json:"secret_key"`
	ForcePathStyle bool   `json:"force_path_style"`
}

type PubsubConfig struct {
	Type         PubsubType  `json:"type"`
	BufferLength uint        `json:"buffer_length"`
	RedisConfig  RedisConfig `json:"redis_config"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     uint   `json:"port"`
	Password string `json:"password"`
	Database uint   `json:"database"`
}

type TaskQueueConfig struct {
	Type                           TaskQueueType       `json:"type"`
	RemoteTaskQueueType            RemoteTaskQueueType `json:"remote_task_queue_type" gorm:"default:'none'"`
	MaxOutstandingMessagesPerQueue uint                `json:"max_outstanding_messages_per_queue"`
	NoOfWorkersPerQueue            uint                `json:"no_of_workers_per_queue"`
	AmqpConfig                     AmqpConfig          `json:"amqp_config"`
	RedisConfig                    RedisConfig         `json:"redis_config"`
}

type AmqpConfig struct {
	Protocol TaskQueueQueueProtocol `json:"protocol"`
	Host     string                 `json:"host"`
	Port     uint                   `json:"port"`
	Username string                 `json:"username"`
	Password string                 `json:"password"`
	Vhost    string                 `json:"vhost"`
}

type NewAdminCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
