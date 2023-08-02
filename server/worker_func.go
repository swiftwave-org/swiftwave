package server

import (
	"log"
	"time"
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