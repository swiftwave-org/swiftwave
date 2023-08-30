package server

import (
	"time"

	"github.com/labstack/echo/v4"
)

// Init functions
func (server *Server) InitRedirectRestAPI() {
	server.ECHO_SERVER.GET("/mapping/redirects", server.getRedirectRules)
	server.ECHO_SERVER.GET("/mapping/redirects/:id", server.getRedirectRule)
	server.ECHO_SERVER.POST("/mapping/redirects", server.createRedirectRule)
	server.ECHO_SERVER.DELETE("/mapping/redirects/:id", server.deleteRedirectRule)
}

// REST API functions

// GET /mapping/redirects
func (server *Server) getRedirectRules(c echo.Context) error {
	var redirectRules []RedirectRule
	tx := server.DB_CLIENT.Find(&redirectRules)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to fetch redirect rules from database",
		})
	}
	return c.JSON(200, redirectRules)
}

// GET /mapping/redirects/:id
func (server *Server) getRedirectRule(c echo.Context) error {
	id := c.Param("id")
	var redirectRule RedirectRule
	tx := server.DB_CLIENT.Where("id = ?", id).First(&redirectRule)
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "redirect rule not found",
		})
	}
	return c.JSON(200, redirectRule)
}

// POST /mapping/redirects
func (server *Server) createRedirectRule(c echo.Context) error {
	var redirectRule RedirectRule
	err := c.Bind(&redirectRule)
	if err != nil {
		return c.JSON(400, map[string]interface{}{
			"error":   err.Error(),
			"message": "Failed to decode request body",
		})
	}
	// verify params
	if redirectRule.DomainName == "" {
		return c.JSON(400, map[string]interface{}{
			"error":   "DomainName is required",
			"message": "Failed to create redirect rule",
		})
	}
	// verify it should have port 80
	if redirectRule.Port != 80 {
		return c.JSON(400, map[string]interface{}{
			"error":   "Port should be 80",
			"message": "Failed to create redirect rule",
		})
	}
	// check if redirect url is valid
	if redirectRule.RedirectURL == "" {
		return c.JSON(400, map[string]interface{}{
			"error":   "RedirectURL is required",
			"message": "Failed to create redirect rule",
		})
	}
	redirectRule.ID = 0
	// check if already exists
	var redirectRuleDB RedirectRule
	tx := server.DB_CLIENT.Where("domain_name = ?", redirectRule.DomainName).First(&redirectRuleDB)
	if tx.Error == nil {
		return c.JSON(400, map[string]interface{}{
			"error":   "DomainName already exists",
			"message": "Failed to create redirect rule",
		})
	}
	// check for same domain, if found
	var ingressRuleInConflict IngressRule
	tx2 := server.DB_CLIENT.Where("domain_name = ? AND protocol = ? AND port = ?", redirectRule.DomainName, HTTPProtcol, redirectRule.Port).First(&ingressRuleInConflict)
	if tx2.Error == nil {
		return c.JSON(400, map[string]interface{}{
			"error":   "DomainName already exists as ingress rule",
			"message": "Failed to create redirect rule",
		})
	}
	// create redirect rule
	redirectRule.Status = RedirectRuleStatusPending
	redirectRule.UpdatedAt = time.Now()
	tx = server.DB_CLIENT.Create(&redirectRule)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to create redirect rule",
		})
	}
	// return success
	return c.JSON(200, redirectRule)
}

// DELETE /mapping/redirects/:id
func (server *Server) deleteRedirectRule(c echo.Context) error {
	id := c.Param("id")
	var redirectRule RedirectRule
	tx := server.DB_CLIENT.Where("id = ?", id).First(&redirectRule)
	if tx.Error != nil {
		return c.JSON(404, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "redirect rule not found",
		})
	}
	if redirectRule.Status == RedirectRuleStatusDeletePending {
		return c.JSON(400, map[string]interface{}{
			"error":   "redirect rule already queued for deletion",
			"message": "Failed to delete redirect rule",
		})
	}
	// update status
	redirectRule.Status = RedirectRuleStatusDeletePending
	redirectRule.UpdatedAt = time.Now()
	tx = server.DB_CLIENT.Save(&redirectRule)
	if tx.Error != nil {
		return c.JSON(500, map[string]interface{}{
			"error":   tx.Error.Error(),
			"message": "Failed to delete redirect rule",
		})
	}
	// return success
	return c.JSON(200, redirectRule)
}
