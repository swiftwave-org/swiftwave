package worker

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

// Manager : hold references to other functions of service
type Manager struct {
	ServiceConfig  *core.ServiceConfig
	ServiceManager *core.ServiceManager
}

// Request Payload

// DeployApplicationRequest : request payload for deploy application
type DeployApplicationRequest struct {
	AppId string `json:"app_id"`
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