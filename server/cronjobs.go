package server

import (
	"log"
	"reflect"
	"time"

	HAPROXY_MANAGER "github.com/swiftwave-org/swiftwave/haproxy_manager"

	"github.com/docker/docker/api/types/swarm"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Server) InitCronJobs() {
	go s.MovePendingApplicationsToImageGenerationQueueCronjob()
	go s.MoveRedeployPendingApplicationsToImageGenerationQueueCronjob()
	go s.MoveDeployingPendingApplicationsToDeployingQueueCronjob()
	go s.ProcessIngressRulesRequestCronjob()
	go s.ProcessRedirectRulesRequestCronjob()
	go s.HAproxyExposedPortsProcessor()
	go s.CleanExpiredSessionTokensCronjob()
}

// Move `pending` applications to `image generation queue` for building docker image
func (s *Server) MovePendingApplicationsToImageGenerationQueueCronjob() {
	var logRecord ApplicationBuildLog
	for {
		// Get all pending applications
		var applications []Application
		tx := s.DB_CLIENT.Where("status = ?", ApplicationStatusPending).Find(&applications)
		if tx.Error != nil {
			log.Println(tx.Error)
			time.Sleep(5 * time.Second)
			continue
		}
		// Move them to image generation queue
		for _, application := range applications {
			log.Println("Moving application to image generation queue: ", application.ServiceName)
			err := s.DB_CLIENT.Transaction(func(tx *gorm.DB) error {
				// reset log record
				logRecord = ApplicationBuildLog{}
				// Update status
				application.Status = ApplicationStatusBuildingImageQueued
				tx2 := tx.Save(&application)
				if tx2.Error != nil {
					log.Println(tx2.Error)
					return tx2.Error
				}
				// Create log record
				logRecord = ApplicationBuildLog{
					ID:            uuid.New().String(),
					ApplicationID: application.ID,
					Logs:          "Queued for image generation\n",
					Time:          time.Now(),
				}
				tx3 := tx.Create(&logRecord)
				if tx3.Error != nil {
					log.Println(tx3.Error)
					return tx3.Error
				}
				return nil
			})
			if err == nil {
				// Enqueue
				err = s.AddServiceToDockerImageGenerationQueue(application.ServiceName, logRecord.ID)
				if err != nil {
					log.Println("failed to enqueue application to image generation task queue: ", err)
				}
			}
			if err != nil {
				log.Println("Error while moving pending applications to image generation queue: ", err)
			} else {
				s.AddLogToApplicationBuildLog(logRecord.ID, "Successfully enqueued for image generation", "success", true)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// Move `redeploy_pending` applications to `image generation queue` for building docker image
func (s *Server) MoveRedeployPendingApplicationsToImageGenerationQueueCronjob() {
	var logRecord ApplicationBuildLog
	for {
		// Get all pending applications
		var applications []Application
		tx := s.DB_CLIENT.Where("status = ?", ApplicationStatusRedeployPending).Find(&applications)
		if tx.Error != nil {
			log.Println(tx.Error)
			time.Sleep(5 * time.Second)
			continue
		}
		// Move them to image generation queue
		for _, application := range applications {
			log.Println("Moving application to image generation queue: ", application.ServiceName)
			err := s.DB_CLIENT.Transaction(func(tx *gorm.DB) error {
				// reset log record
				logRecord = ApplicationBuildLog{}
				// Update status
				application.Status = ApplicationStatusBuildingImageQueued
				tx2 := tx.Save(&application)
				if tx2.Error != nil {
					log.Println(tx2.Error)
					return tx2.Error
				}
				// Create log record
				logRecord = ApplicationBuildLog{
					ID:            uuid.New().String(),
					ApplicationID: application.ID,
					Logs:          "Queued for image generation\n",
					Time:          time.Now(),
				}
				tx3 := tx.Create(&logRecord)
				if tx3.Error != nil {
					log.Println(tx3.Error)
					return tx3.Error
				}
				return nil
			})
			if err == nil {
				// Enqueue
				err = s.AddServiceToDockerImageGenerationQueue(application.ServiceName, logRecord.ID)
				if err != nil {
					log.Println("failed to enqueue application to image generation task queue: ", err)
				}
			}
			if err != nil {
				log.Println("Error while moving redeploy_pending applications to image generation queue: ", err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// Move `deploying_pending` applications to `deploying queue` for deploying the application
func (s *Server) MoveDeployingPendingApplicationsToDeployingQueueCronjob() {
	for {
		// Get all pending applications
		var applications []Application
		tx := s.DB_CLIENT.Where("status = ?", ApplicationStatusDeployingPending).Find(&applications)
		if tx.Error != nil {
			log.Println(tx.Error)
			time.Sleep(5 * time.Second)
			continue
		}
		// Move them to image generation queue
		for _, application := range applications {
			log.Println("Moving application to deploying-queue: ", application.ServiceName)
			err := s.DB_CLIENT.Transaction(func(tx *gorm.DB) error {
				// Update status
				application.Status = ApplicationStatusDeployingQueued
				tx2 := tx.Save(&application)
				if tx2.Error != nil {
					log.Println(tx2.Error)
					return tx2.Error
				}

				// Enqueue
				err := s.AddServiceToDeployQueue(application.ServiceName)
				if err != nil {
					log.Println(err)
					return err
				}
				return nil
			})
			if err != nil {
				log.Println("Error while moving deploying_pending applications to deploying queue: ", err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// Process ingress rules request - `pending` , `delete_pending` status records
func (s *Server) ProcessIngressRulesRequestCronjob() {
	for {
		var ingressRules []IngressRule
		tx := s.DB_CLIENT.Where("status = ? OR status = ?", IngressRuleStatusPending, IngressRuleStatusDeletePending).Find(&ingressRules)
		if tx.Error != nil {
			log.Println(tx.Error)
			time.Sleep(5 * time.Second)
			continue
		}
		for _, ingressRule := range ingressRules {
			transaction_id, err := s.HAPROXY_MANAGER.FetchNewTransactionId()
			if err != nil {
				log.Println(err)
				continue
			}
			if ingressRule.Status == IngressRuleStatusPending {
				// add backend
				backend_name := s.HAPROXY_MANAGER.GenerateBackendName(ingressRule.ServiceName, int(ingressRule.ServicePort))
				// skip if backend already exists - check db for service name and port and status != pending and id != ingressRule.ID
				backendDoesNotExist := false
				var ingressRuleCheck IngressRule
				tx := s.DB_CLIENT.Where("id != ? AND service_name = ? AND service_port = ? AND status != ?", ingressRule.ID, ingressRule.ServiceName, ingressRule.ServicePort, IngressRuleStatusPending).First(&ingressRuleCheck)
				if tx.Error != nil {
					if tx.Error == gorm.ErrRecordNotFound {
						backendDoesNotExist = true
					}
				}
				// if backend does not exist, create it
				if backendDoesNotExist {
					_, err := s.HAPROXY_MANAGER.AddBackend(transaction_id, ingressRule.ServiceName, int(ingressRule.ServicePort), 1)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				}
				// create backend switch rule
				if ingressRule.Protocol == HTTPSProtcol {
					err = s.HAPROXY_MANAGER.AddHTTPSLink(transaction_id, backend_name, ingressRule.DomainName)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				} else if ingressRule.Protocol == HTTPProtcol && ingressRule.Port == 80 {
					err = s.HAPROXY_MANAGER.AddHTTPLink(transaction_id, backend_name, ingressRule.DomainName)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				} else {
					var listenerMode HAPROXY_MANAGER.ListenerMode
					if ingressRule.Protocol == TCPProtcol {
						listenerMode = HAPROXY_MANAGER.TCPMode
					} else {
						listenerMode = HAPROXY_MANAGER.HTTPMode
					}
					err = s.HAPROXY_MANAGER.AddTCPLink(transaction_id, backend_name, int(ingressRule.Port), ingressRule.DomainName, listenerMode, s.RESTRICTED_PORTS)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				}
				// commit transaction
				err = s.HAPROXY_MANAGER.CommitTransaction(transaction_id)
				if err != nil {
					err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
					if err2 != nil {
						log.Println(err2)
					}
					log.Println(err)
					continue
				}
				// update status
				ingressRule.Status = IngressRuleStatusApplied
				tx2 := s.DB_CLIENT.Save(&ingressRule)
				if tx2.Error != nil {
					log.Println(tx2.Error)
				}
			} else if ingressRule.Status == IngressRuleStatusDeletePending {
				backend_name := s.HAPROXY_MANAGER.GenerateBackendName(ingressRule.ServiceName, int(ingressRule.ServicePort))
				// delete backend switch rule
				if ingressRule.Protocol == HTTPSProtcol {
					err = s.HAPROXY_MANAGER.DeleteHTTPSLink(transaction_id, backend_name, ingressRule.DomainName)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				} else if ingressRule.Protocol == HTTPProtcol && ingressRule.Port == 80 {
					err = s.HAPROXY_MANAGER.DeleteHTTPLink(transaction_id, backend_name, ingressRule.DomainName)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				} else {
					err = s.HAPROXY_MANAGER.DeleteTCPLink(transaction_id, backend_name, int(ingressRule.Port), ingressRule.DomainName, s.RESTRICTED_PORTS)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				}
				// ensure this backend is not used by any other ingress rules
				// check by service name and port and status != delete pending
				backendUsedByOther := true
				var ingressRuleCheck IngressRule
				tx := s.DB_CLIENT.Where("id != ? AND service_name = ? AND service_port = ? AND status != ?", ingressRule.ID, ingressRule.ServiceName, ingressRule.ServicePort, IngressRuleStatusDeletePending).First(&ingressRuleCheck)
				if tx.Error != nil {
					if tx.Error == gorm.ErrRecordNotFound {
						backendUsedByOther = false
					}
				}
				if !backendUsedByOther {
					// delete backend
					err = s.HAPROXY_MANAGER.DeleteBackend(transaction_id, backend_name)
					if err != nil {
						err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
						if err2 != nil {
							log.Println(err2)
						}
						log.Println(err)
						continue
					}
				}
				// commit transaction
				err = s.HAPROXY_MANAGER.CommitTransaction(transaction_id)
				if err != nil {
					err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
					if err2 != nil {
						log.Println(err2)
					}
					log.Println(err)
					continue
				}
				// delete ingress rule
				tx2 := s.DB_CLIENT.Delete(&ingressRule)
				if tx2.Error != nil {
					log.Println(tx2.Error)
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// Process to maintain exposed-ports of haproxy-service
func (s *Server) HAproxyExposedPortsProcessor() {
	for {
		// Fetch all ingress rules with only port field
		var ingressRules []IngressRule
		tx := s.DB_CLIENT.Select("port").Where("port IS NOT NULL").Find(&ingressRules)
		if tx.Error != nil {
			log.Println(tx.Error)
			continue
		}
		// Serialize port
		var portsmap map[int]bool = make(map[int]bool)
		for _, ingressRule := range ingressRules {
			portsmap[int(ingressRule.Port)] = true
		}
		// add 80 and 443 to ports
		portsmap[80] = true
		portsmap[443] = true
		if !s.isProductionEnvironment() {
			portsmap[5555] = true
		}
		// Check if ports are changed
		exposedPorts, err := s.DOCKER_MANAGER.FetchPublishedHostPorts(s.HAPROXY_SERVICE)
		if err != nil {
			log.Println(err)
			continue
		}
		exposedPortsMap := make(map[int]bool)
		for _, port := range exposedPorts {
			exposedPortsMap[port] = true
		}
		portsNotChanged := reflect.DeepEqual(exposedPortsMap, portsmap)
		if !portsNotChanged {
			var ports_update_required []swarm.PortConfig = make([]swarm.PortConfig, 0)
			for port := range portsmap {
				ports_update_required = append(ports_update_required, swarm.PortConfig{
					Protocol:      swarm.PortConfigProtocolTCP,
					PublishMode:   swarm.PortConfigPublishModeHost,
					TargetPort:    uint32(port),
					PublishedPort: uint32(port),
				})
			}
			// Update exposed ports
			err := s.DOCKER_MANAGER.UpdatePublishedHostPorts(s.HAPROXY_SERVICE, ports_update_required)
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Exposed ports are changed")
			}
		}
		time.Sleep(20 * time.Second)
	}
}

// Process redirect rules request - `pending` and `delete_pending`
func (s *Server) ProcessRedirectRulesRequestCronjob() {
	for {
		// Fetch all redirect rules with status pending
		var redirectRules []RedirectRule
		tx := s.DB_CLIENT.Where("status != ?", RedirectRuleStatusApplied).Find(&redirectRules)
		if tx.Error != nil {
			log.Println(tx.Error)
			continue
		}
		// Process redirect rules
		for _, redirectRule := range redirectRules {
			// create transaction
			transaction_id, err := s.HAPROXY_MANAGER.FetchNewTransactionId()
			if err != nil {
				log.Println(err)
				continue
			}
			if redirectRule.Status == RedirectRuleStatusPending {
				// create redirect rule
				err = s.HAPROXY_MANAGER.AddHTTPRedirectRule(transaction_id, redirectRule.DomainName, redirectRule.RedirectURL)
				if err != nil {
					err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
					if err2 != nil {
						log.Println(err2)
					}
					log.Println(err)
					continue
				}
				// commit transaction
				err = s.HAPROXY_MANAGER.CommitTransaction(transaction_id)
				if err != nil {
					err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
					if err2 != nil {
						log.Println(err2)
					}
					log.Println(err)
					continue
				}
				// update status
				redirectRule.Status = RedirectRuleStatusApplied
				tx2 := s.DB_CLIENT.Save(&redirectRule)
				if tx2.Error != nil {
					log.Println(tx2.Error)
				}
			} else if redirectRule.Status == RedirectRuleStatusDeletePending {
				// delete redirect rule
				err = s.HAPROXY_MANAGER.DeleteHTTPRedirectRule(transaction_id, redirectRule.DomainName)
				if err != nil {
					err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
					if err2 != nil {
						log.Println(err2)
					}
					log.Println(err)
					continue
				}
				// commit transaction
				err = s.HAPROXY_MANAGER.CommitTransaction(transaction_id)
				if err != nil {
					err2 := s.HAPROXY_MANAGER.DeleteTransaction(transaction_id)
					if err2 != nil {
						log.Println(err2)
					}
					log.Println(err)
					continue
				}
				// delete redirect rule
				tx2 := s.DB_CLIENT.Delete(&redirectRule)
				if tx2.Error != nil {
					log.Println(tx2.Error)
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// Cleanip expired session tokens
func (s *Server) CleanExpiredSessionTokensCronjob() {
	for {
		var removeTokens []string = make([]string, 0)
		for token, sessionToken := range s.SESSION_TOKENS {
			if sessionToken.Before(time.Now()) {
				removeTokens = append(removeTokens, token)
			}
		}
		// Remove expired tokens
		for _, token := range removeTokens {
			delete(s.SESSION_TOKENS, token)
		}
		time.Sleep(10 * time.Second)
	}
}
