package worker

func (m Manager) EnqueueBuildApplicationRequest(applicationId string, deploymentId string) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(buildApplicationQueueName, BuildApplicationRequest{
		AppId:        applicationId,
		DeploymentId: deploymentId,
	})
}

func (m Manager) EnqueueDeployApplicationRequest(applicationId string, deploymentId string) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(deployApplicationQueueName, DeployApplicationRequest{
		AppId:             applicationId,
		DeploymentId:      deploymentId,
		IgnoreProxyUpdate: false,
	})
}

func (m Manager) EnqueueDeployApplicationRequestWithNoProxyUpdate(applicationId string, deploymentId string) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(deployApplicationQueueName, DeployApplicationRequest{
		AppId:             applicationId,
		DeploymentId:      deploymentId,
		IgnoreProxyUpdate: true,
	})
}

func (m Manager) EnqueueDeleteApplicationRequest(applicationId string) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(deleteApplicationQueueName, DeleteApplicationRequest{
		Id: applicationId,
	})
}

func (m Manager) EnqueueSSLGenerateRequest(domainId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(sslGenerateQueueName, SSLGenerateRequest{
		DomainId: domainId,
	})
}

func (m Manager) EnqueueIngressRuleApplyRequest(ingressRuleId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(ingressRuleApplyQueueName, IngressRuleApplyRequest{
		Id: ingressRuleId,
	})
}

func (m Manager) EnqueueIngressRuleDeleteRequest(ingressRuleId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(ingressRuleDeleteQueueName, IngressRuleDeleteRequest{
		Id: ingressRuleId,
	})
}

func (m Manager) EnqueueIngressRuleHttpsRedirectRequest(ingressRuleId uint, enabled bool) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(ingressRuleHttpsRedirectQueueName, IngressRuleHttpsRedirectRequest{
		Id:      ingressRuleId,
		Enabled: enabled,
	})
}

func (m Manager) EnqueueRedirectRuleApplyRequest(redirectRuleId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(redirectRuleApplyQueueName, RedirectRuleApplyRequest{
		Id: redirectRuleId,
	})
}

func (m Manager) EnqueueRedirectRuleDeleteRequest(redirectRuleId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(redirectRuleDeleteQueueName, RedirectRuleDeleteRequest{
		Id: redirectRuleId,
	})
}

func (m Manager) EnqueuePersistentVolumeBackupRequest(persistentVolumeBackupId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(persistentVolumeBackupQueueName, PersistentVolumeBackupRequest{
		Id: persistentVolumeBackupId,
	})
}

func (m Manager) EnqueuePersistentVolumeRestoreRequest(persistentVolumeRestoreId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(persistentVolumeRestoreQueueName, PersistentVolumeRestoreRequest{
		Id: persistentVolumeRestoreId,
	})
}

func (m Manager) EnqueueInstallDependenciesOnServerRequest(serverId uint, logId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(installDependenciesOnServerQueueName, InstallDependenciesOnServerRequest{
		ServerId: serverId,
		LogId:    logId,
	})
}

func (m Manager) EnqueueSetupServerRequest(serverId uint, logId uint, advertiseIP string) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(setupServerQueueName, SetupServerRequest{
		ServerId:    serverId,
		LogId:       logId,
		AdvertiseIP: advertiseIP,
	})
}

func (m Manager) EnqueueSetupAndEnableProxyRequest(serverId uint, logId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(setupAndEnableProxyQueueName, SetupAndEnableProxyRequest{
		ServerId: serverId,
		LogId:    logId,
	})
}

func (m Manager) EnqueueDeletePersistentVolumeRequest(persistentVolumeId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(deletePersistentVolumeQueueName, PersistentVolumeDeletionRequest{
		Id: persistentVolumeId,
	})
}

func (m Manager) EnqueueUpdateApplicationOnServerScheduleDeploymentUpdateRequest(serverId uint) error {
	return m.ServiceManager.TaskQueueClient.EnqueueTask(updateApplicationOnServerScheduleDeploymentUpdateQueueName, UpdateApplicationOnServerScheduleDeploymentStatusUpdateRequest{
		ServerId: serverId,
	})
}
