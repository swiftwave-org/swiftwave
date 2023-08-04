package server

import (
	"encoding/json"
	"errors"
	DOCKER_MANAGER "keroku/m/container_manager"
	DOCKER_CONFIG_GENERATOR "keroku/m/docker_config_generator"
	GIT_MANAGER "keroku/m/git_manager"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Server) ProcessGenerateSSLRequestFromQueue(name string) error {
	var domainRecord Domain
	if err := s.DB_CLIENT.Where("name = ?", name).First(&domainRecord).Error; err != nil {
		log.Println("Failed to fetch domain record from database")
		return err
	}
	// Send request to Let's Encrypt
	cert, err := s.SSL_MANAGER.ObtainCertificate(domainRecord.Name, domainRecord.SSLPrivateKey)
	if err != nil {
		log.Println("Failed to obtain certificate from Let's Encrypt")
		return err
	}
	// Update domain in database
	domainRecord.SSLStatus = DomainSSLStatusIssued
	domainRecord.SSLFullChain = cert
	domainRecord.SSLIssuedAt = time.Now()
	domainRecord.SSLIssuer = "Let's Encrypt"
	tx3 := s.DB_CLIENT.Save(&domainRecord)
	if tx3.Error != nil {
		log.Println("Failed to update domain ssl certificate in database")
		return tx3.Error
	}
	// Move certificate to certificates folder
	err = s.AddDomainToSSLUpdateHAProxyQueue(domainRecord.Name)
	if err != nil {
		log.Println("Failed to enqueue domain for ssl certificate update")
	}
	return nil
}

func (s *Server) ProcessUpdateSSLHAProxyRequestFromQueue(name string) error {
	var domainRecord Domain
	if err := s.DB_CLIENT.Where("name = ?", name).First(&domainRecord).Error; err != nil {
		log.Println("Failed to fetch domain record from database")
		return err
	}
	// Move certificate to certificates folder
	transaction_id, err := s.HAPROXY_MANAGER.FetchNewTransactionId()
	if err != nil {
		log.Println("Failed to fetch new transaction id")
		return err
	}
	// Update SSL certificate
	err = s.HAPROXY_MANAGER.UpdateSSL(transaction_id, domainRecord.Name, []byte(domainRecord.SSLPrivateKey), []byte(domainRecord.SSLFullChain))
	if err != nil {
		log.Println("Failed to update SSL certificate in HAProxy")
		return err
	}
	// Commit transaction
	err = s.HAPROXY_MANAGER.CommitTransaction(transaction_id)
	if err != nil {
		log.Println("Failed to commit transaction")
		return err
	}
	return nil
}

// Application deployment tasks
func (s *Server) ProcessDockerImageGenerationRequestFromQueue(app_id uint, log_id string) error {
	var application Application
	if err := s.DB_CLIENT.Preload("Source.GitCredential").Preload(clause.Associations).Where("id = ?", app_id).First(&application).Error; err != nil {
		log.Println("Failed to fetch application record from database")
		s.AddLogToApplicationDeployLog(log_id, "Failed to fetch application record from database", "error")
		return err
	}
	// Update application status
	application.Status = ApplicationStatusBuildingImage
	tx := s.DB_CLIENT.Save(&application)
	if tx.Error != nil {
		s.AddLogToApplicationDeployLog(log_id, "Failed to update application status in database", "warn")
		log.Println("Failed to update application status in database")
	}

	// Generate docker image -- declare as failed if error
	var buildargs map[string]string
	err := json.Unmarshal([]byte(application.BuildArgs), &buildargs)
	if err != nil {
		log.Println("Failed to unmarshal build args")
		s.AddLogToApplicationDeployLog(log_id, "Failed to unmarshal build args", "error")
		application.Status = ApplicationStatusBuildingImageFailed
		tx := s.DB_CLIENT.Save(&application)
		if tx.Error != nil {
			s.AddLogToApplicationDeployLog(log_id, "Failed to update application status in database", "error")
			log.Println("Failed to update application status in database")
		}
		return err
	}
	// Start building based on source type
	if application.Source.Type == ApplicationSourceTypeGit {
		// Create temporary directory
		tempDirectory := "/tmp/" + uuid.New().String()
		os.Mkdir(tempDirectory, 0777)
		// Defer remove temporary directory
		defer os.RemoveAll(tempDirectory)
		// Clone git repository
		err = GIT_MANAGER.CloneRepository(application.Source.RepositoryURL(), application.Source.Branch, application.Source.GitCredential.Username, application.Source.GitCredential.Password, tempDirectory)
		if err != nil {
			failImageBuildUpdateStatus(&application, s.DB_CLIENT)
			s.AddLogToApplicationDeployLog(log_id, "Failed to clone git repository", "error")
			return err
		}
		// Fetch latest commit hash
		commitHash, err := GIT_MANAGER.FetchLatestCommitHash(application.Source.RepositoryURL(), application.Source.Branch, application.Source.GitCredential.Username, application.Source.GitCredential.Password)
		if err != nil {
			failImageBuildUpdateStatus(&application, s.DB_CLIENT)
			s.AddLogToApplicationDeployLog(log_id, "Failed to fetch latest commit hash", "error")
			return err
		}
		s.AddLogToApplicationDeployLog(log_id, "Fetched latest commit hash: "+commitHash, "info")
		// Image name
		imageName := application.ServiceName + ":" + commitHash
		// Build docker image
		scanner, err := s.DOCKER_MANAGER.CreateImage(application.Dockerfile, buildargs, tempDirectory, imageName)
		if err != nil {
			failImageBuildUpdateStatus(&application, s.DB_CLIENT)
			return err
		}
		if scanner != nil {
			var data map[string]interface{}
			for scanner.Scan() {
				err = json.Unmarshal(scanner.Bytes(), &data)
				if err != nil {
					continue
				}
				if data["stream"] != nil {
					s.AddLogToApplicationDeployLog(log_id, data["stream"].(string), "info")
				}
			}
		}
		// Update image name
		application.Image = imageName
		// Update application status
		application.Status = ApplicationStatusBuildingImageCompleted
		tx2 := s.DB_CLIENT.Save(&application)
		if tx2.Error != nil {
			s.AddLogToApplicationDeployLog(log_id, "Failed to update application status in database", "error")
			log.Println("Failed to update application status in database")
		}
		// Update application commit hash
		source := ApplicationSource{
			ID: application.SourceID,
		}
		tx2 = s.DB_CLIENT.Model(&source).Update("last_commit", commitHash)
		if tx2.Error != nil {
			s.AddLogToApplicationDeployLog(log_id, "Failed to update application commit hash in database", "error")
			log.Println("Failed to update application commit hash in database")
		}
	} else if application.Source.Type == ApplicationSourceTypeTarball {
		tarballpath := filepath.Join(s.CODE_TARBALL_DIR, application.Source.TarballFile)
		// Verify file exists
		if _, err := os.Stat(tarballpath); os.IsNotExist(err) {
			log.Println("Tarball file does not exist")
			s.AddLogToApplicationDeployLog(log_id, "Tarball file does not exist", "error")
			failImageBuildUpdateStatus(&application, s.DB_CLIENT)
			return err
		}
		// Create temporary directory
		tempDirectory := "/tmp/" + uuid.New().String()
		os.Mkdir(tempDirectory, 0777)
		// Defer remove temporary directory
		defer os.RemoveAll(tempDirectory)
		// Extract tarball
		err = DOCKER_CONFIG_GENERATOR.ExtractTar(tarballpath, tempDirectory)
		if err != nil {
			log.Println("Failed to extract tarball")
			s.AddLogToApplicationDeployLog(log_id, "Failed to extract tarball", "error")
			failImageBuildUpdateStatus(&application, s.DB_CLIENT)
			return err
		}
		// Image name
		imageName := application.ServiceName + ":" + uuid.NewString()
		// Build docker image
		s.AddLogToApplicationDeployLog(log_id, "Building docker image", "info")
		scanner, err := s.DOCKER_MANAGER.CreateImage(application.Dockerfile, buildargs, tempDirectory, imageName)
		if err != nil {
			s.AddLogToApplicationDeployLog(log_id, "Failed to build docker image", "error")
			failImageBuildUpdateStatus(&application, s.DB_CLIENT)
			return err
		}
		if scanner != nil {
			var data map[string]interface{}
			for scanner.Scan() {
				err = json.Unmarshal(scanner.Bytes(), &data)
				if err != nil {
					continue
				}
				if data["stream"] != nil {
					s.AddLogToApplicationDeployLog(log_id, data["stream"].(string), "info")
				}
			}
		}
		// Update image name
		application.Image = imageName
		// Update application status
		application.Status = ApplicationStatusBuildingImageCompleted
		tx2 := s.DB_CLIENT.Save(&application)
		if tx2.Error != nil {
			s.AddLogToApplicationDeployLog(log_id, "Failed to update application status in database", "error")
			log.Println("Failed to update application status in database")
		}

	} else if application.Source.Type == ApplicationSourceTypeImage {
		log.Println("Application source type is image, skipping image generation")
		s.AddLogToApplicationDeployLog(log_id, "Application source type is image, skipping image generation", "info")
		// Update image name
		application.Image = application.Source.DockerImage
		// Update application status
		application.Status = ApplicationStatusBuildingImageCompleted
		tx2 := s.DB_CLIENT.Save(&application)
		if tx2.Error != nil {
			log.Println("Failed to update application status in database")
		}
	}
	s.AddLogToApplicationDeployLog(log_id, "Successfully built docker image"+application.Image, "info")

	// Deploy service
	// Update application status to deploying_pending
	application.Status = ApplicationStatusDeployingPending
	tx3 := s.DB_CLIENT.Save(&application)
	if tx3.Error != nil {
		log.Println("Failed to update application status in database")
	}
	return nil
}

func (s *Server) ProcessDeployServiceRequestFromQueue(app_id uint) error {
	// Fetch application from database
	var application Application
	if err := s.DB_CLIENT.Preload("Source.GitCredential").Preload(clause.Associations).Where("id = ?", app_id).First(&application).Error; err != nil {
		log.Println("Failed to fetch application record from database")
		return err
	}
	// Verify application status
	if application.Status != ApplicationStatusDeployingPending && application.Status != ApplicationStatusDeployingQueued {
		log.Println("Application status is not deployment pending state")
		failApplicationDeployUpdateStatus(&application, s.DB_CLIENT)
		return errors.New("Application status is not deployment pending state")
	}
	// update status to deploying
	application.Status = ApplicationStatusDeploying
	tx2 := s.DB_CLIENT.Save(&application)
	if tx2.Error != nil {
		log.Println("Failed to update application status in database")
	}
	// Check if image is present
	if application.Image == "" {
		log.Println("Application image is empty")
		failApplicationDeployUpdateStatus(&application, s.DB_CLIENT)
		return errors.New("Application image is empty")
	}
	// Environment variables
	var environmentVariables map[string]string = make(map[string]string)
	err := json.Unmarshal([]byte(application.EnvironmentVariables), &environmentVariables)
	if err != nil {
		log.Println("Failed to unmarshal environment variables")
		failApplicationDeployUpdateStatus(&application, s.DB_CLIENT)
		return err
	}
	// Deploy service
	service := DOCKER_MANAGER.Service{
		Name:         application.ServiceName,
		Image:        application.Image,
		Command:      []string{},
		Env:          environmentVariables,
		Networks:     []string{s.SWARM_NETWORK},
		Replicas:     uint64(application.Replicas),
		VolumeMounts: []DOCKER_MANAGER.VolumeMount{},
	}
	err = s.DOCKER_MANAGER.CreateService(service)
	if err != nil {
		log.Println("Failed to create service, fallback try to update service")
		err = s.DOCKER_MANAGER.UpdateService(service)
		if err != nil {
			log.Println("Failed to update service, fallback try to remove service")
			err = s.DOCKER_MANAGER.RemoveService(service.Name)
			if err != nil {
				log.Println("Failed to remove service")
			}
		}
		failApplicationDeployUpdateStatus(&application, s.DB_CLIENT)
	}
	if err == nil {
		// update status
		application.Status = ApplicationStatusRunning
		tx3 := s.DB_CLIENT.Save(&application)
		if tx3.Error != nil {
			log.Println("Failed to update application status in database")
		}
	}
	return err
}

func failImageBuildUpdateStatus(application *Application, db_client gorm.DB) {
	application.Status = ApplicationStatusBuildingImageFailed
	tx := db_client.Save(&application)
	if tx.Error != nil {
		log.Println("Failed to update application status in database")
	}
}

func failApplicationDeployUpdateStatus(application *Application, db_client gorm.DB) {
	application.Status = ApplicationStatusDeployingFailed
	tx := db_client.Save(&application)
	if tx.Error != nil {
		log.Println("Failed to update application status in database")
	}
}
