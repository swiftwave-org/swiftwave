package system_config

import (
	"fmt"
	"strings"
)

func (p PostgresqlConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=%s", p.Host, p.Port, p.User, p.Password, p.Database, p.TimeZone, p.SSLMode)
}

func (a AMQPConfig) URI() string {
	return fmt.Sprintf("%s://%s:%s@%s", a.Protocol, a.User, a.Password, a.Host)
}

func (c ServiceConfig) IsAllDomainsAllowed() bool {
	if len(c.WhiteListedDomains) == 0 {
		return true
	}
	for _, domain := range c.WhiteListedDomains {
		if strings.Trim(domain, " ") == "*" {
			return true
		}
	}
	return false
}
