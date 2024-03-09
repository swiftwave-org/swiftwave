package local_config

type Config struct {
	IsDevelopmentMode bool             `yaml:"dev_mode"`
	ServiceConfig     ServiceConfig    `yaml:"service"`
	PostgresqlConfig  PostgresqlConfig `yaml:"postgresql"`
}

type ServiceConfig struct {
	UseTLS                    bool   `yaml:"use_tls"`
	ManagementNodeAddress     string `yaml:"management_node_address"`
	BindAddress               string `yaml:"bind_address"`
	BindPort                  int    `yaml:"bind_port"`
	SocketPathDirectory       string `yaml:"-"`
	DataDirectory             string `yaml:"-"`
	NetworkName               string `yaml:"-"`
	HAProxyServiceName        string `yaml:"-"`
	HAProxyUnixSocketPath     string `yaml:"-"`
	HAProxyDataDirectoryPath  string `yaml:"-"`
	UDPProxyServiceName       string `yaml:"-"`
	UDPProxyDataDirectoryPath string `yaml:"-"`
	SSLCertDirectoryPath      string `yaml:"-"`
	LogDirectoryPath          string `yaml:"-"`
	InfoLogFilePath           string `yaml:"-"`
	ErrorLogFilePath          string `yaml:"-"`
}

type PostgresqlConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	User                   string `yaml:"user"`
	Password               string `yaml:"password"`
	Database               string `yaml:"database"`
	TimeZone               string `yaml:"time_zone"`
	SSLMode                string `yaml:"ssl_mode"`
	AutoStartLocalPostgres bool   `yaml:"auto_start_local_postgres"`
}
