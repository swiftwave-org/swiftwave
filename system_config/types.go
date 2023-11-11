package system_config

type Config struct {
	Version           string            `yaml:"version"`
	Mode              Mode              `yaml:"mode"`
	Environment       Environment       `yaml:"environment"`
	ServiceConfig     ServiceConfig     `yaml:"service_config"`
	HAProxyConfig     HAProxyConfig     `yaml:"haproxy_config"`
	PostgresqlConfig  PostgresqlConfig  `yaml:"postgresql_config"`
	LetsEncryptConfig LetsEncryptConfig `yaml:"lets_encrypt_config"`
	PubSubConfig      PubSubConfig      `yaml:"pubsub_config"`
	TaskQueueConfig   TaskQueueConfig   `yaml:"task_queue_config"`
}

type ServiceConfig struct {
	AutoTLS              bool   `yaml:"auto_tls"`
	BindAddress          string `yaml:"bind_address"`
	BindPort             int    `yaml:"bind_port"`
	NetworkName          string `yaml:"network_name"`
	DataDir              string `yaml:"data_dir"`
	DockerUnixSocketPath string `yaml:"docker_unix_socket_path"`
	RestrictedPorts      []int  `yaml:"restricted_ports"`
}

type PostgresqlConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	TimeZone string `yaml:"time_zone"`
}

type HAProxyConfig struct {
	UnixSocketPath string `yaml:"unix_socket_path"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
}

type LetsEncryptConfig struct {
	StagingEnvironment bool   `yaml:"staging_environment"`
	EmailID            string `yaml:"email_id"`
	PrivateKeyPath     string `yaml:"private_key_path"`
}

type PubSubConfig struct {
	Mode         PubSubMode  `yaml:"mode"`
	BufferLength int         `yaml:"buffer_length"`
	RedisConfig  RedisConfig `yaml:"redis_config"`
}

type TaskQueueConfig struct {
	Mode                           TaskQueueMode `yaml:"mode"`
	MaxOutstandingMessagesPerQueue int           `yaml:"max_outstanding_messages_per_queue"`
	RedisConfig                    RedisConfig
}

type RedisConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Password   string `yaml:"password"`
	DatabaseID int    `yaml:"database_id"`
}

type AMQPConfig struct {
	Protocol   AMQPProtocol `yaml:"protocol"`
	Host       string       `yaml:"host"`
	User       string       `yaml:"user"`
	Password   string       `yaml:"password"`
	VHost      string       `yaml:"vhost"`
	ClientName string       `yaml:"client_name"`
}
