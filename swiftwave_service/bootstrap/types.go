package bootstrap

type ImageRegistryType string
type PubsubType string
type TaskQueueType string
type TaskQueueQueueProtocol string

const (
	LocalRegistry   ImageRegistryType      = "local"
	RemoteRegistry  ImageRegistryType      = "remote"
	LocalPubsub     PubsubType             = "local"
	RemotePubsub    PubsubType             = "remote"
	LocalTaskQueue  TaskQueueType          = "local"
	RemoteTaskQueue TaskQueueType          = "remote"
	AMQP            TaskQueueQueueProtocol = "amqp"
	AMQPS           TaskQueueQueueProtocol = "amqps"
)

type SystemConfigurationPayload struct {
	NetworkName          string `json:"network_name"`
	ExtraRestrictedPorts string `json:"extra_restricted_ports"`
	LetsEncrypt          struct {
		EmailAddress string `json:"email_address"`
		StagingEnv   bool   `json:"staging_env"`
	} `json:"lets_encrypt"`
	ImageRegistry struct {
		Type      ImageRegistryType `json:"type"`
		Endpoint  string            `json:"endpoint"`
		Namespace string            `json:"namespace"`
		Username  string            `json:"username"`
		Password  string            `json:"password"`
	} `json:"image_registry"`
	HAProxyConfig struct {
		Image string `json:"image"`
	} `json:"haproxy_config"`
	UDPProxyConfig struct {
		Image string `json:"image"`
	} `json:"udpproxy_config"`
	PvBackupConfig struct {
		S3Config struct {
			Enabled        bool   `json:"enabled"`
			Endpoint       string `json:"endpoint"`
			Region         string `json:"region"`
			BucketName     string `json:"bucket_name"`
			AccessKeyId    string `json:"access_key_id"`
			SecretKey      string `json:"secret_key"`
			ForcePathStyle bool   `json:"force_path_style"`
		} `json:"s3_config"`
	} `json:"pv_backup_config"`
	PubsubConfig struct {
		Type         PubsubType `json:"type"`
		BufferLength uint       `json:"buffer_length"`
		RedisConfig  struct {
			Host     string `json:"host"`
			Port     uint   `json:"port"`
			Password string `json:"password"`
			Database uint   `json:"database"`
		} `json:"redis_config"`
	} `json:"pubsub_config"`
	TaskQueueConfig struct {
		Type                           TaskQueueType `json:"type"`
		MaxOutstandingMessagesPerQueue uint          `json:"max_outstanding_messages_per_queue"`
		NoOfWorkersPerQueue            uint          `json:"no_of_workers_per_queue"`
		AmqpConfig                     struct {
			Protocol TaskQueueQueueProtocol `json:"protocol"`
			Host     string                 `json:"host"`
			Port     uint                   `json:"port"`
			Username string                 `json:"username"`
			Password string                 `json:"password"`
			Vhost    string                 `json:"vhost"`
		} `json:"amqp_config"`
	} `json:"task_queue_config"`
}
