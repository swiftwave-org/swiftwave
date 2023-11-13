package system_config

<<<<<<< HEAD
import (
	"fmt"
	"strings"
)
=======
import "fmt"
>>>>>>> 5f6e33e0fb2a7d5fd0d52314aef4a850df72ec56

func (p PostgresqlConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=disable", p.Host, p.Port, p.User, p.Password, p.Database, p.TimeZone)
}

func (a AMQPConfig) URI() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d", a.Protocol, a.User, a.Password, a.Host)
}
<<<<<<< HEAD

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
=======
>>>>>>> 5f6e33e0fb2a7d5fd0d52314aef4a850df72ec56
