package system_config

type Config struct {
	Version           string            `yaml:"version"`
	IsDevelopmentMode bool              `yaml:"-"`
	Mode              Mode              `yaml:"mode"`
	ServiceConfig     ServiceConfig     `yaml:"service"`
	HAProxyConfig     HAProxyConfig     `yaml:"haproxy"`
	PostgresqlConfig  PostgresqlConfig  `yaml:"postgresql"`
	LetsEncryptConfig LetsEncryptConfig `yaml:"lets_encrypt"`
	PubSubConfig      PubSubConfig      `yaml:"pubsub"`
	TaskQueueConfig   TaskQueueConfig   `yaml:"task_queue"`
}

type ServiceConfig struct {
	AutoMigrateDatabase  bool   `yaml:"auto_migrate_database"`
	UseTLS               bool   `yaml:"use_tls"`
	SSLCertificateDir    string `yaml:"ssl_certificate_dir"`
	AddressOfCurrentNode string `yaml:"address_of_current_node"`
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
	SSLMode  string `yaml:"ssl_mode"`
}

type HAProxyConfig struct {
	ServiceName    string `yaml:"service_name"`
	DockerImage    string `yaml:"image"`
	UnixSocketPath string `yaml:"unix_socket_path"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DataDir        string `yaml:"data_dir"`
}

type LetsEncryptConfig struct {
	StagingEnvironment    bool   `yaml:"staging_environment"`
	EmailID               string `yaml:"email_id"`
	AccountPrivateKeyPath string `yaml:"account_private_key_path"`
}

type PubSubConfig struct {
	Mode         PubSubMode  `yaml:"mode"`
	BufferLength int         `yaml:"buffer_length"`
	RedisConfig  RedisConfig `yaml:"redis"`
}

type TaskQueueConfig struct {
	Mode                           TaskQueueMode `yaml:"mode"`
	MaxOutstandingMessagesPerQueue int           `yaml:"max_outstanding_messages_per_queue"`
	AMQPConfig                     AMQPConfig    `yaml:"amqp"`
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
