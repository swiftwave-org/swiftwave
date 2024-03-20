package core

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

// SystemLog : hold log of system
type SystemLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Metadata  string    `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// Server : hold information about server
type Server struct {
	ID                    uint                   `json:"id" gorm:"primaryKey"`
	IP                    string                 `json:"ip"`
	HostName              string                 `json:"host_name" gorm:"unique"`
	User                  string                 `json:"user"`
	ScheduleDeployments   bool                   `json:"schedule_deployments" gorm:"default:true"`
	DockerUnixSocketPath  string                 `json:"docker_unix_socket_path"`
	SwarmMode             SwarmMode              `json:"swarm_mode"`
	ProxyConfig           ProxyConfig            `json:"proxy_config" gorm:"embedded;embeddedPrefix:proxy_"`
	Status                ServerStatus           `json:"status"`
	LastPing              time.Time              `json:"last_ping"`
	Logs                  []ServerLog            `json:"logs" gorm:"foreignKey:ServerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ConsoleTokens         []ConsoleToken         `json:"console_tokens" gorm:"foreignKey:ServerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AnalyticsServiceToken *AnalyticsServiceToken `json:"analytics_service_token" gorm:"foreignKey:ServerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ResourceStats         []ServerResourceStat   `json:"resource_stats" gorm:"foreignKey:ServerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// ServerLog : hold logs of server
type ServerLog struct {
	*gorm.Model
	ID       uint   `json:"id" gorm:"primaryKey"`
	ServerID uint   `json:"serverID"`
	Title    string `json:"title"` // can be empty, but will be useful if we want to
	Content  string `json:"content"`
}

// User : hold information about user
type User struct {
	ID           uint     `json:"id" gorm:"primaryKey"`
	Username     string   `json:"username" gorm:"unique"`
	Role         UserRole `json:"role" gorm:"default:'user'"`
	PasswordHash string   `json:"password_hash"`
}

// ************************************************************************************* //
//                                Application Level Table       		   			     //
// ************************************************************************************* //

// GitCredential : credential for git client
type GitCredential struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:GitCredentialID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" `
}

// ImageRegistryCredential : credential for docker image registry
type ImageRegistryCredential struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Url         string       `json:"url"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:ImageRegistryCredentialID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Domain : hold information about domain
type Domain struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Name          string          `json:"name" gorm:"unique"`
	SSLStatus     DomainSSLStatus `json:"ssl_status"`
	SSLPrivateKey string          `json:"ssl_private_key"`
	SSLFullChain  string          `json:"ssl_full_chain"`
	SSLIssuedAt   time.Time       `json:"ssl_issued_at"`
	SSLIssuer     string          `json:"ssl_issuer"`
	SSLAutoRenew  bool            `json:"ssl_auto_renew" gorm:"default:false"`
	IngressRules  []IngressRule   `json:"ingress_rules" gorm:"foreignKey:DomainID"`
	RedirectRules []RedirectRule  `json:"redirect_rules" gorm:"foreignKey:DomainID"`
}

// IngressRule : hold information about Ingress rule for service
type IngressRule struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	DomainID      *uint             `json:"domain_id,omitempty" gorm:"default:null"`
	ApplicationID string            `json:"application_id"`
	Protocol      ProtocolType      `json:"protocol"`
	Port          uint              `json:"port"`        // external port
	TargetPort    uint              `json:"target_port"` // port of the application
	Status        IngressRuleStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// RedirectRule : hold information about Redirect rules for domain
type RedirectRule struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	DomainID    uint               `json:"domain_id"`
	Protocol    ProtocolType       `json:"protocol"`
	RedirectURL string             `json:"redirect_url"`
	Status      RedirectRuleStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// PersistentVolume : hold information about persistent volume
type PersistentVolume struct {
	ID                       uint                      `json:"id" gorm:"primaryKey"`
	Name                     string                    `json:"name" gorm:"unique"`
	Type                     PersistentVolumeType      `json:"type" gorm:"default:'local'"`
	NFSConfig                NFSConfig                 `json:"nfs_config" gorm:"embedded;embeddedPrefix:nfs_config_"`
	PersistentVolumeBindings []PersistentVolumeBinding `json:"persistent_volume_bindings" gorm:"foreignKey:PersistentVolumeID"`
	PersistentVolumeBackups  []PersistentVolumeBackup  `json:"persistent_volume_backups" gorm:"foreignKey:PersistentVolumeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	PersistentVolumeRestores []PersistentVolumeRestore `json:"persistent_volume_restores" gorm:"foreignKey:PersistentVolumeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// PersistentVolumeBinding : hold information about persistent volume binding
type PersistentVolumeBinding struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	ApplicationID      string `json:"application_id"`
	PersistentVolumeID uint   `json:"persistent_volume_id"`
	MountingPath       string `json:"mounting_path"`
}

// PersistentVolumeBackup : hold information about persistent volume backup
type PersistentVolumeBackup struct {
	ID                 uint         `json:"id" gorm:"primaryKey"`
	Type               BackupType   `json:"type"`
	Status             BackupStatus `json:"status"`
	File               string       `json:"file"`
	FileSizeMB         float64      `json:"file_size_mb"`
	PersistentVolumeID uint         `json:"persistent_volume_id"`
	CreatedAt          time.Time    `json:"created_at"`
	CompletedAt        time.Time    `json:"completed_at"`
}

// PersistentVolumeRestore : hold information about persistent volume restore
type PersistentVolumeRestore struct {
	ID                 uint          `json:"id" gorm:"primaryKey"`
	Type               RestoreType   `json:"type"`
	File               string        `json:"file"`
	Status             RestoreStatus `json:"status"`
	PersistentVolumeID uint          `json:"persistent_volume_id"`
	CreatedAt          time.Time     `json:"created_at"`
	CompletedAt        time.Time     `json:"completed_at"`
}

// EnvironmentVariable : hold information about environment variable
type EnvironmentVariable struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ApplicationID string `json:"application_id"`
	Key           string `json:"key"`
	Value         string `json:"value"`
}

// BuildArg : hold information about build args
type BuildArg struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	DeploymentID string `json:"deployment_id"`
	Key          string `json:"key"`
	Value        string `json:"value"`
}

// Application : hold information about application
type Application struct {
	ID   string `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique"`
	// Environment Variables
	// On change of environment variables, deployment will be triggered by force update
	EnvironmentVariables []EnvironmentVariable `json:"environment_variables" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Persistent Volumes
	// On change of persistent volumes, deployment will be triggered by force update
	PersistentVolumeBindings []PersistentVolumeBinding `json:"persistent_volume_bindings" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// No of replicas to be deployed
	DeploymentMode DeploymentMode `json:"deployment_mode"`
	Replicas       uint           `json:"replicas"`
	// Deployments
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Latest Deployment
	LatestDeployment Deployment `json:"-"`
	// Ingress Rules
	IngressRules []IngressRule `json:"ingress_rules" gorm:"foreignKey:ApplicationID"`
	// Command
	Command string `json:"command"`
	// Capabilities
	Capabilities pq.StringArray `json:"capabilities" gorm:"type:text[]"`
	// Sysctls
	Sysctls pq.StringArray `json:"sysctls" gorm:"type:text[]"`
	// ConsoleTokens
	ConsoleTokens []ConsoleToken `json:"console_tokens" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Is deleted - soft delete - will be deleted from database in background
	IsDeleted bool `json:"is_deleted" gorm:"default:false"`
	// Webhook token
	WebhookToken string `json:"webhook_token"`
	// Sleeping
	IsSleeping bool `json:"is_sleeping" gorm:"default:false"`
	// Resource Stats
	ResourceStats []ApplicationServiceResourceStat `json:"resource_stats" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Deployment : hold information about deployment of application
type Deployment struct {
	ID            string       `json:"id" gorm:"primaryKey"`
	ApplicationID string       `json:"application_id"`
	UpstreamType  UpstreamType `json:"upstream_type"`
	// Fields for UpstreamType = Git
	GitCredentialID  *uint       `json:"git_credential_id"`
	GitProvider      GitProvider `json:"git_provider"`
	RepositoryOwner  string      `json:"repository_owner"`
	RepositoryName   string      `json:"repository_name"`
	RepositoryBranch string      `json:"repository_branch"`
	CodePath         string      `json:"code_path"`
	CommitHash       string      `json:"commit_hash"`
	// Fields for UpstreamType = SourceCode
	SourceCodeCompressedFileName string `json:"source_code_compressed_file_name"`
	// Fields for UpstreamType = Image
	DockerImage               string `json:"docker_image"`
	ImageRegistryCredentialID *uint  `json:"image_registry_credential_id"`
	// Common Fields
	BuildArgs  []BuildArg `json:"build_args" gorm:"foreignKey:DeploymentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Dockerfile string     `json:"dockerfile"`
	// Logs
	Logs []DeploymentLog `json:"logs" gorm:"foreignKey:DeploymentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Deployment Status
	Status DeploymentStatus `json:"status"`
	// Created At
	CreatedAt time.Time `json:"created_at"`
}

// DeploymentLog : hold logs of deployment
type DeploymentLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DeploymentID string    `json:"deployment_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}

// ConsoleToken hold information about console auth tokens, used in establishing websocket connection
// Note this
// If Target == ConsoleTargetTypeServer, ServerID denote which server to ssh into
// If Target == ConsoleTargetTypeApplication, ApplicationID denote which application to connect to and ServerID denote which server to connect to.
// In case of ConsoleTargetTypeApplication, we will connect to ServerID and try to ssh into the application container
// If ServerID server has no container for the application, we will return error
type ConsoleToken struct {
	ID            string        `json:"id" gorm:"primaryKey"`
	Target        ConsoleTarget `json:"target_type"`
	ServerID      *uint         `json:"server_id"`
	ApplicationID *string       `json:"application_id"`
	Token         string        `json:"token" gorm:"unique"`
	ExpiresAt     time.Time     `json:"expires_at"`
}

type AnalyticsServiceToken struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Token     string    `json:"token" gorm:"unique"`
	ServerID  uint      `json:"server_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ************************************************************************************* //
//                                Server Related Stats       		   			         //
// ************************************************************************************* //

// ServerResourceStat : struct to hold host resource stats
type ServerResourceStat struct {
	ID              uint             `json:"id" gorm:"primaryKey"`
	ServerID        uint             `json:"server_id"`
	CpuUsagePercent uint8            `json:"cpu_used_percent"`
	MemStat         ServerMemoryStat `json:"memory" gorm:"embedded;embeddedPrefix:memory_"`
	DiskStats       ServerDiskStats  `json:"disks"`
	NetStat         ServerNetStat    `json:"network" gorm:"embedded;embeddedPrefix:network_"`
	RecordedAt      time.Time        `json:"recorded_at"`
}

// ************************************************************************************* //
//                                Server Related Stats       		   			         //
// ************************************************************************************* //

// ApplicationServiceResourceStat : struct to hold service resource stats
type ApplicationServiceResourceStat struct {
	ID                   uint                      `json:"id" gorm:"primaryKey"`
	ApplicationID        string                    `json:"application_id"`
	CpuUsagePercent      uint8                     `json:"cpu_used_percent"`
	ReportingServerCount uint                      `json:"reporting_server_count"` // to help in calculating average
	UsedMemoryMB         uint64                    `json:"used_memory_mb"`
	NetStat              ApplicationServiceNetStat `json:"network" gorm:"embedded;embeddedPrefix:network_"`
	RecordedAt           time.Time                 `json:"recorded_at"`
}
