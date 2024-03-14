package system_config

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

var config *SystemConfig
var configVersion uint = 0

func Fetch(db *gorm.DB) (*SystemConfig, error) {
	if config != nil {
		// Fetch the latest version of the config
		var record SystemConfig
		tx := db.First(&record).Select("config_version")
		if tx.Error != nil {
			return nil, tx.Error
		}
		// if the version is the same, return the cached config
		if record.ConfigVersion == configVersion {
			return config, nil
		}
	}
	// fetch first record
	var record SystemConfig
	tx := db.First(&record)
	if tx.Error != nil {
		return nil, tx.Error
	}
	config = &record
	configVersion = record.ConfigVersion
	return config, nil
}

func (config *SystemConfig) Create(db *gorm.DB) error {
	// check if there is only one record
	var count int64
	tx := db.Model(&SystemConfig{}).Count(&count)
	if tx.Error != nil {
		return tx.Error
	}
	if count > 0 {
		return fmt.Errorf("system config already exists! consider updating it instead")
	}
	tx = db.Create(config)
	return tx.Error
}

func (config *SystemConfig) Update(db *gorm.DB) error {
	// fetch the latest version of the config
	var record SystemConfig
	tx := db.First(&record)
	if tx.Error != nil {
		return tx.Error
	}
	// update the id
	config.ID = record.ID
	config.ConfigVersion++
	tx = db.Updates(config)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (a AMQPConfig) URI() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s", a.Protocol, a.User, a.Password, a.Host, a.Port, a.VHost)
}

func (r ImageRegistryConfig) URI() string {
	if strings.Compare(r.Namespace, "") == 0 {
		return r.Endpoint
	}
	return fmt.Sprintf("%s/%s", r.Endpoint, r.Namespace)
}

func (r ImageRegistryConfig) IsConfigured() bool {
	return strings.Compare(r.Endpoint, "") != 0
}

func (config *SystemConfig) PublicSSHKey() (string, error) {
	// Decode the PEM-encoded private key
	block, _ := pem.Decode([]byte(config.SshPrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	// Parse the private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Convert the private key to a public key
	publicKey := &privateKey.PublicKey

	// Get the public key in OpenSSH format
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	pubKey := fmt.Sprintf("ssh-rsa %X swiftwave", pubKeyBytes)
	return pubKey, nil
}
