package main

import (
	"errors"
	"fmt"
)

type NodeType string

const (
	MasterNode NodeType = "master"
	WorkerNode NodeType = "worker"
)

type AgentConfig struct {
	ID                      uint                    `gorm:"primaryKey"`
	NodeType                NodeType                `json:"node_type" gorm:"column:node_type"`
	SwiftwaveServiceAddress string                  `json:"swiftwave_service_address" gorm:"column:swiftwave_service_address"`
	WireguardConfig         WireguardConfig         `json:"wireguard_config" gorm:"embedded;embeddedPrefix:wireguard_"`
	MasterNodeConnectConfig MasterNodeConnectConfig `json:"master_node_connect_config" gorm:"embedded;embeddedPrefix:master_node_connect_config_"`
	DockerNetwork           DockerNetworkConfig     `json:"docker_network" gorm:"embedded;embeddedPrefix:docker_network_"`
	HaproxyConfig           HAProxyConfig           `json:"haproxy_config" gorm:"embedded;embeddedPrefix:haproxy_"`
}

type WireguardConfig struct {
	PrivateKey string `json:"private_key" gorm:"column:private_key"`
	Address    string `json:"address" gorm:"column:address"`
}

type MasterNodeConnectConfig struct {
	Endpoint   string `json:"endpoint" gorm:"column:endpoint"`
	PublicKey  string `json:"public_key" gorm:"column:public_key"`
	AllowedIPs string `json:"allowed_ips" gorm:"column:allowed_ips"`
}

type DockerNetworkConfig struct {
	BridgeId       string `json:"bridge_id" gorm:"column:bridge_id"`
	GatewayAddress string `json:"gateway_address" gorm:"column:gateway_address"`
	Subnet         string `json:"subnet" gorm:"column:subnet"`
}

type HAProxyConfig struct {
	Enabled  bool   `json:"enabled" gorm:"column:enabled"`
	Username string `json:"username" gorm:"column:username"`
	Password string `json:"password" gorm:"column:password"`
}

func GetConfig() (*AgentConfig, error) {
	var config AgentConfig
	if err := rDB.First(&config).Error; err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("error getting config")
	}
	return &config, nil
}

func SetConfig(config *AgentConfig) error {
	// Check if the config already exists
	var count int64
	if err := rwDB.Model(&AgentConfig{}).Count(&count).Error; err != nil {
		return err
	}
	config.ID = 1 // set so that it can be updated
	if count > 0 {
		// Update the existing config
		return rwDB.Save(config).Error
	}
	return rwDB.Create(config).Error
}

func RemoveConfig() error {
	var count int64
	if err := rwDB.Model(&AgentConfig{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	return rwDB.Delete(&AgentConfig{}).Where("id = ?", 1).Error
}
