package worker

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
)

// Manager : hold references to other functions of service
type Manager struct {
	Config         *config.Config
	ServiceManager *service_manager.ServiceManager
}

// Queue names
const (
	buildApplicationQueueName            = "build_application"
	deployApplicationQueueName           = "deploy_application"
	deleteApplicationQueueName           = "delete_application"
	ingressRuleApplyQueueName            = "ingress_rule_apply"
	ingressRuleDeleteQueueName           = "ingress_rule_delete"
	redirectRuleApplyQueueName           = "redirect_rule_apply"
	redirectRuleDeleteQueueName          = "redirect_rule_delete"
	sslGenerateQueueName                 = "ssl_generate"
	persistentVolumeBackupQueueName      = "persistent_volume_backup"
	persistentVolumeRestoreQueueName     = "persistent_volume_restore"
	installDependenciesOnServerQueueName = "install_dependencies_on_server"
	setupServerQueueName                 = "setup_server"
	setupAndEnableProxyQueueName         = "setup_and_enable_proxy"
	deletePersistentVolumeQueueName      = "delete_persistent_volume"
)

// Request Payload

// DeployApplicationRequest : request payload for deploy application
type DeployApplicationRequest struct {
	AppId             string `json:"app_id"`
	DeploymentId      string `json:"deployment_id"`
	IgnoreProxyUpdate bool   `json:"ignore_proxy_update"`
}

// BuildApplicationRequest : request payload for deploy application
type BuildApplicationRequest struct {
	AppId        string `json:"app_id"`
	DeploymentId string `json:"deployment_id"`
}

// IngressRuleApplyRequest : request payload for ingress rule apply
type IngressRuleApplyRequest struct {
	Id uint `json:"id"`
}

// IngressRuleDeleteRequest : request payload for ingress rule delete
type IngressRuleDeleteRequest struct {
	Id uint `json:"id"`
}

// RedirectRuleApplyRequest : request payload for redirect rule apply
type RedirectRuleApplyRequest struct {
	Id uint `json:"id"`
}

// RedirectRuleDeleteRequest : request payload for redirect rule delete
type RedirectRuleDeleteRequest struct {
	Id uint `json:"id"`
}

// SSLGenerateRequest : request payload for ssl generate
type SSLGenerateRequest struct {
	DomainId uint `json:"domain_id"`
}

// DeleteApplicationRequest : request payload for application delete
type DeleteApplicationRequest struct {
	Id string `json:"id"`
}

// PersistentVolumeBackupRequest : request payload for persistent volume backup
type PersistentVolumeBackupRequest struct {
	Id uint `json:"id"`
}

// PersistentVolumeRestoreRequest : request payload for persistent volume restore
type PersistentVolumeRestoreRequest struct {
	Id uint `json:"id"`
}

// InstallDependenciesOnServerRequest : request payload for install dependencies on server
type InstallDependenciesOnServerRequest struct {
	ServerId uint `json:"server_id"`
	LogId    uint `json:"log_id"`
}

// SetupServerRequest : request payload for setup server
type SetupServerRequest struct {
	ServerId uint `json:"server_id"`
	LogId    uint `json:"log_id"`
}

// SetupAndEnableProxyRequest : request payload for setup server
type SetupAndEnableProxyRequest struct {
	ServerId uint `json:"server_id"`
	LogId    uint `json:"log_id"`
}

// PersistentVolumeDeletionRequest : request payload for delete persistent volume
type PersistentVolumeDeletionRequest struct {
	Id uint `json:"id"`
}
