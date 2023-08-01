package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/labstack/echo/v4"
)

// Types

type Domain struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Name          string          `json:"name"`
	SSLStatus     DomainSSLStatus `json:"ssl_status"`
	SSLPrivateKey string          `json:"ssl_private_key"`
	SSLFullChain  string          `json:"ssl_full_chain"`
	SSLIssuedAt   time.Time       `json:"ssl_issued_at"`
	SSLIssuer     string          `json:"ssl_issuer"`
}

type DomainSSLStatus string

const (
	DomainSSLStatusNone    DomainSSLStatus = "none"
	DomainSSLStatusIssued  DomainSSLStatus = "issued"
	DomainSSLStatusIssuing DomainSSLStatus = "issuing"
)

// Init functions

func (server *Server) InitDomainRestAPI() {
	server.ECHO_SERVER.GET("/domains", server.GetDomains)
	server.ECHO_SERVER.GET("/domains/:id", server.GetDomain)
	server.ECHO_SERVER.POST("/domains", server.CreateDomain)
	server.ECHO_SERVER.DELETE("/domains/:id", server.DeleteDomain)
	server.ECHO_SERVER.POST("/domains/:id/ssl/issue", server.IssueDomainSSL)
}

// REST API functions

func (server *Server) MigrateDomainDB() {
	server.DB_CLIENT.AutoMigrate(&Domain{})
}

// GET /domains
func (server *Server) GetDomains(c echo.Context) error {
	// Fetch all domains from database
	var domains []Domain
	tx := server.DB_CLIENT.Find(&domains)
	if tx.Error != nil {
		c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to fetch domains from database",
		})
		return nil
	}
	// Return domains
	return c.JSON(200, domains)
}

// GET /domains/:id
func (server *Server) GetDomain(c echo.Context) error {
	if c.Param("id") == "" {
		c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
		return nil
	}
	var domain Domain
	tx := server.DB_CLIENT.First(&domain, c.Param("id"))
	if tx.Error != nil {
		c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "domain not found",
		})
		return nil
	}
	return c.JSON(200, domain)
}

// POST /domains
func (server *Server) CreateDomain(c echo.Context) error {
	// JSON decode request body
	var domain Domain
	tx := c.Bind(&domain)
	if tx != nil {
		c.JSON(400, map[string]interface{}{
			"error":   tx.Error(),
			"message": "Failed to decode request body",
		})
		return nil
	}
	// Validate request body
	if domain.Name == "" {
		c.JSON(400, map[string]interface{}{
			"message": "name field is required",
		})
		return nil
	}
	// Cleanup extra fields
	domain.ID = 0
	domain.SSLStatus = DomainSSLStatusNone
	domain.SSLPrivateKey = ""
	domain.SSLFullChain = ""
	domain.SSLIssuedAt = time.Time{}
	domain.SSLIssuer = ""
	// Check if domain already exists
	var existingDomain Domain
	tx2 := server.DB_CLIENT.Where("name = ?", domain.Name).First(&existingDomain)
	if tx2.Error == nil {
		c.JSON(409, map[string]interface{}{
			"message": "Domain already exists",
		})
		return nil
	}
	// Create domain
	tx3 := server.DB_CLIENT.Create(&domain)
	if tx3.Error != nil {
		c.JSON(500, map[string]interface{}{
			"error":   tx3.Error.Error(),
			"message": "Failed to create domain",
		})
		return nil
	}
	// Return domain
	return c.JSON(200, domain)
}

// DELETE /domains/:id
func (server *Server) DeleteDomain(c echo.Context) error {
	if c.Param("id") == "" {
		c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
		return nil
	}
	// Fetch domain from database
	var domain Domain
	tx := server.DB_CLIENT.First(&domain, c.Param("id"))
	if tx.Error != nil {
		c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "domain not found",
		})
		return nil
	}
	// Delete domain
	tx2 := server.DB_CLIENT.Delete(&domain)
	if tx2.Error != nil {
		c.JSON(500, map[string]interface{}{
			"error":   tx2.Error.Error(),
			"message": "Failed to delete domain",
		})
		return nil
	}
	// Return domain
	return c.JSON(200, domain)
}

// TODO: if SSL certificate is already issued, only reissue if it's expired or force is true
// POST /domains/:id/ssl/issue
func (server *Server) IssueDomainSSL(c echo.Context) error {
	if c.Param("id") == "" {
		c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
		return nil
	}
	// Fetch domain from database
	var domain Domain
	tx := server.DB_CLIENT.First(&domain, c.Param("id"))
	if tx.Error != nil {
		c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "domain not found",
		})
	}
	// If no private key is set, generate one
	if domain.SSLPrivateKey == "" {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			c.JSON(500, map[string]interface{}{
				"error":   err.Error(),
				"message": "Failed to generate private key",
			})
			return nil
		}
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		pemKey := pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		}
		privateKeyBytes = pem.EncodeToMemory(&pemKey)
		domain.SSLPrivateKey = string(privateKeyBytes)
		// Update domain in database
		tx2 := server.DB_CLIENT.Save(&domain)
		if tx2.Error != nil {
			c.JSON(500, map[string]interface{}{
				"error":   tx2.Error.Error(),
				"message": "Failed to update domain ssl private key",
			})
			return nil
		}
	}
	// TODO: move to queue
	// Obtain certificate from Let's Encrypt
	cert, err := server.SSL_MANAGER.ObtainCertificate(domain.Name, domain.SSLPrivateKey)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error":   err.Error(),
			"message": "Failed to obtain certificate",
		})
		return nil
	}
	// Update domain in database
	domain.SSLStatus = DomainSSLStatusIssued
	domain.SSLFullChain = cert
	domain.SSLIssuedAt = time.Now()
	domain.SSLIssuer = "Let's Encrypt"
	tx3 := server.DB_CLIENT.Save(&domain)
	if tx3.Error != nil {
		c.JSON(500, map[string]interface{}{
			"error":   tx3.Error.Error(),
			"message": "Failed to update domain ssl certificate",
		})
	}
	// TODO: move to queue
	// Move certificate to certificates folder
	transaction_id, err := server.HAPROXY_MANAGER.FetchNewTransactionId()
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error":   err.Error(),
			"message": "failed to update SSL certificate in HAProxy",
		})
		return nil
	}
	// Update SSL certificate
	err = server.HAPROXY_MANAGER.UpdateSSL(transaction_id, domain.Name, []byte(domain.SSLPrivateKey), []byte(domain.SSLFullChain))
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error":   err.Error(),
			"message": "failed to update SSL certificate in HAProxy",
		})
		return nil
	}
	// Commit transaction
	err = server.HAPROXY_MANAGER.CommitTransaction(transaction_id)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"error":   err.Error(),
			"message": "failed to update SSL certificate in HAProxy",
		})
		return nil
	}
	// Return domain
	return c.JSON(200, domain)
}
