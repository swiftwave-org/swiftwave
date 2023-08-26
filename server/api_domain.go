package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/labstack/echo/v4"
)



// Init functions

func (server *Server) InitDomainRestAPI() {
	server.ECHO_SERVER.GET("/domains", server.getDomains)
	server.ECHO_SERVER.GET("/domains/:id", server.getDomain)
	server.ECHO_SERVER.GET("/domains/shortlist", server.getShortlistedDomains)
	server.ECHO_SERVER.POST("/domains", server.createDomain)
	server.ECHO_SERVER.DELETE("/domains/:id", server.deleteDomain)
	server.ECHO_SERVER.POST("/domains/:id/ssl/issue", server.issueDomainSSL)
}

// REST API functions

// GET /domains
func (server *Server) getDomains(c echo.Context) error {
	// Fetch all domains from database
	var domains []Domain
	tx := server.DB_CLIENT.Find(&domains)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to fetch domains from database",
		})
	}
	// Return domains
	return c.JSON(200, domains)
}

// GET /domains/:id
func (server *Server) getDomain(c echo.Context) error {
	if c.Param("id") == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
	}
	var domain Domain
	tx := server.DB_CLIENT.First(&domain, c.Param("id"))
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "domain not found",
		})
	}
	return c.JSON(200, domain)
}

// GET /domains/shortlist
func (server *Server) getShortlistedDomains(c echo.Context) error {
	// Fetch all domains from database
	var domains []Domain
	tx := server.DB_CLIENT.Select("name").Find(&domains)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to fetch domains from database",
		})
	}
	// Filter domains
	var shortlistedDomains []string = []string{}
	for _, domain := range domains {
		shortlistedDomains = append(shortlistedDomains, domain.Name)
	}
	// Return domains
	return c.JSON(200, shortlistedDomains)
}

// POST /domains
func (server *Server) createDomain(c echo.Context) error {
	// JSON decode request body
	var domain Domain
	tx := c.Bind(&domain)
	if tx != nil {
		return c.JSON(400, map[string]interface{}{
			"error":   tx.Error(),
			"message": "Failed to decode request body",
		})
	}
	// Validate request body
	if domain.Name == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "name field is required",
		})
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
		return c.JSON(409, map[string]interface{}{
			"message": "Domain already exists",
		})
	}
	// Create domain
	tx3 := server.DB_CLIENT.Create(&domain)
	if tx3.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx3.Error.Error(),
			"message": "Failed to create domain",
		})
	}
	// Return domain
	return c.JSON(200, domain)
}

// DELETE /domains/:id
func (server *Server) deleteDomain(c echo.Context) error {
	if c.Param("id") == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
	}
	// Fetch domain from database
	var domain Domain
	tx := server.DB_CLIENT.First(&domain, c.Param("id"))
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "domain not found",
		})
	}
	// Delete domain
	tx2 := server.DB_CLIENT.Delete(&domain)
	if tx2.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx2.Error.Error(),
			"message": "Failed to delete domain",
		})
	}
	// Return domain
	return c.JSON(200, domain)
}

// POST /domains/:id/ssl/issue
func (server *Server) issueDomainSSL(c echo.Context) error {
	if c.Param("id") == "" {
		return c.JSON(400, map[string]interface{}{
			"message": "id parameter is required",
		})
	}
	// Fetch domain from database
	var domain Domain
	tx := server.DB_CLIENT.First(&domain, c.Param("id"))
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "domain not found",
		})
	}
	// If no private key is set, generate one
	if domain.SSLPrivateKey == "" {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return c.JSON(500, map[string]interface{}{
				"error":   err.Error(),
				"message": "Failed to generate private key",
			})
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
			return c.JSON(500, map[string]interface{}{
				"error":   tx2.Error.Error(),
				"message": "Failed to update domain ssl private key",
			})
		}
	}
	// Update status
	domain.SSLStatus = DomainSSLStatusIssuing
	tx3 := server.DB_CLIENT.Save(&domain)
	if tx3.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx3.Error.Error(),
			"message": "Failed to update domain ssl status",
		})
	}
	// Add domain to task queue
	err := server.AddDomainToSSLGenerateQueue(domain.Name)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   err.Error(),
			"message": "Failed to enqueue domain for ssl certificate generation",
		})
	}

	// Return domain
	return c.JSON(200, domain)
}
