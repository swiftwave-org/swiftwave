package system_config

import "fmt"

func (p PostgresqlConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=disable", p.Host, p.Port, p.User, p.Password, p.Database, p.TimeZone)
}
