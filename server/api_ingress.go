package server

import (
	HAPROXY_MANAGER "keroku/m/haproxy_manager"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// Init functions
func (server *Server) InitIngressRestAPI() {
	server.ECHO_SERVER.GET("/ingresses", server.getIngressRules)
	server.ECHO_SERVER.GET("/ingresses/:id", server.getIngressRule)
	server.ECHO_SERVER.POST("/ingresses", server.createIngressRule)
	server.ECHO_SERVER.DELETE("/ingresses/:id", server.deleteIngressRule)
	server.ECHO_SERVER.GET("/ingresses/restricted-ports", server.getRestrictedPorts)
}

// REST API functions

// GET /ingresses
func (server *Server) getIngressRules(c echo.Context) error {
	var ingressRules []IngressRule
	tx := server.DB_CLIENT.Find(&ingressRules)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to fetch ingress rules from database",
		})
	}
	return c.JSON(200, ingressRules)
}

// GET /ingresses/:id
func (server *Server) getIngressRule(c echo.Context) error {
	id := c.Param("id")
	var ingressRule IngressRule
	tx := server.DB_CLIENT.Where("id = ?", id).First(&ingressRule)
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "ingress rule not found",
		})
	}
	return c.JSON(200, ingressRule)
}

// POST /ingresses
func (server *Server) createIngressRule(c echo.Context) error {
	// TODO: check service name before creating ingress rule
	// TODO: check domain name
	var ingressRule IngressRule
	err := c.Bind(&ingressRule)
	if err != nil {
		return c.JSON(400, map[string]interface{}{
			"error":   err.Error(),
			"message": "Failed to decode request body",
		})
	}
	ingressRule.ID = 0
	if ingressRule.Protocol == TCPProtcol {
		ingressRule.DomainName = ""
	}
	// check if restricted port
	if ingressRule.Protocol == TCPProtcol {
		if HAPROXY_MANAGER.IsPortRestrictedForManualConfig(int(ingressRule.Port), server.RESTRICTED_PORTS) {
			return c.JSON(400, map[string]interface{}{
				"message": "Port " + strconv.Itoa(int(ingressRule.Port)) + " is restricted",
			})
		}
	}
	// Verify if using different for https 
	if ingressRule.Protocol == HTTPSProtcol && ingressRule.Port != 443 {
		return c.JSON(400, map[string]interface{}{
			"message": "HTTPS protocol must use port 443",
		})
	}
	// Check if conflicting ingress rule exists
	isConflictFound := false
	if (ingressRule.Protocol == HTTPProtcol && ingressRule.Port == 80) || 
		(ingressRule.Protocol == HTTPSProtcol && ingressRule.Port == 443) {
		// check for same domain, if found
		var ingressRuleInConflict IngressRule
		tx := server.DB_CLIENT.Where("domain_name = ? AND protocol = ? AND port = ?", ingressRule.DomainName, HTTPProtcol, 80).First(&ingressRuleInConflict)
		if tx.Error == nil {
			isConflictFound = true
		}
	}
	if ingressRule.Protocol == HTTPSProtcol && ingressRule.Port == 443 {
		// check for same domain, if found
		var ingressRuleInConflict IngressRule
		tx := server.DB_CLIENT.Where("domain_name = ? AND protocol = ? AND port = ?", ingressRule.DomainName, HTTPSProtcol, 443).First(&ingressRuleInConflict)
		if tx.Error == nil {
			isConflictFound = true
		}
	}
	if !isConflictFound {
		// check for same domain, protocol and port, if found
		var ingressRuleInConflict IngressRule
		if ingressRule.Protocol == HTTPProtcol {
			tx := server.DB_CLIENT.Where("domain_name = ? AND protocol = ? AND port = ? AND service_name = ? AND service_port = ?", ingressRule.DomainName, ingressRule.Protocol, ingressRule.Port, ingressRule.ServiceName, ingressRule.ServicePort).First(&ingressRuleInConflict)
			if tx.Error == nil {
				isConflictFound = true
			}
		} else {
			tx := server.DB_CLIENT.Where("port = ?", ingressRule.Port).First(&ingressRuleInConflict)
			if tx.Error == nil {
				isConflictFound = true
			}
		}
	}

	if isConflictFound {
		return c.JSON(409, map[string]interface{}{
			"error":   "conflict",
			"message": "ingress rule already exists",
		})
	}
	ingressRule.UpdatedAt = time.Now()
	ingressRule.Status = IngressRuleStatusPending
	tx := server.DB_CLIENT.Create(&ingressRule)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to create ingress rule",
		})
	}
	return c.JSON(200, ingressRule)
}

// DELETE /ingresses/:id
func (server *Server) deleteIngressRule(c echo.Context) error {
	id := c.Param("id")
	var ingressRule IngressRule
	tx := server.DB_CLIENT.Where("id = ?", id).First(&ingressRule)
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "ingress rule not found",
		})
	}
	if ingressRule.Status == IngressRuleStatusDeletePending {
		return c.JSON(409, map[string]interface{}{
			"error":   "conflict",
			"message": "ingress rule already marked for deletion",
		})
	}
	if ingressRule.Status == IngressRuleStatusPending {
		return c.JSON(409, map[string]interface{}{
			"error":   "conflict",
			"message": "ingress rule is not yet applied, wait to be applied first, then you can delete",
		})
	}
	ingressRule.Status = IngressRuleStatusDeletePending
	ingressRule.UpdatedAt = time.Now()
	tx = server.DB_CLIENT.Save(&ingressRule)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to delete ingress rule",
		})
	}
	return c.JSON(200, ingressRule)
}

// GET /ingress/restricted-ports
func (server *Server) getRestrictedPorts(c echo.Context) error {
	return c.JSON(200, server.RESTRICTED_PORTS)
}