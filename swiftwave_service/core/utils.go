package core

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
	"time"
)

// SetPassword : set password for user
func (user *User) SetPassword(password string) error {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}
	user.PasswordHash = string(hashedPasswordBytes)
	return nil
}

// CheckPassword : check password for user
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// GenerateJWT : generate jwt token for user
func (user *User) GenerateJWT(jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
		"username": user.Username,
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(jwtSecret))
}

// ReplicaCount : get replica count
func (application *Application) ReplicaCount() uint {
	if application.IsSleeping {
		return 0
	}
	return application.Replicas
}

// IPV4Regex : regex for IPv4
const IPV4Regex = `^(?:\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b)$`

// IsIPv4 : check if the domain is IPv4
func (domain *Domain) IsIPv4() bool {
	regex := regexp.MustCompile(IPV4Regex)
	return regex.MatchString(domain.Name)
}

// IsLocalhost check if the domain is localhost
func (server *Server) IsLocalhost() bool {
	// if `localhost` or `127.0.0.1` or `0.0.0.0`
	return server.IP == "localhost" || server.IP == "127.0.0.1" || server.IP == "0.0.0.0"
}

func (d *DockerProxyConfig) Equal(other *DockerProxyConfig) bool {
	return d.Enabled == other.Enabled &&
		d.Permission.Ping == other.Permission.Ping &&
		d.Permission.Version == other.Permission.Version &&
		d.Permission.Info == other.Permission.Info &&
		d.Permission.Events == other.Permission.Events &&
		d.Permission.Auth == other.Permission.Auth &&
		d.Permission.Secrets == other.Permission.Secrets &&
		d.Permission.Build == other.Permission.Build &&
		d.Permission.Commit == other.Permission.Commit &&
		d.Permission.Configs == other.Permission.Configs &&
		d.Permission.Containers == other.Permission.Containers &&
		d.Permission.Distribution == other.Permission.Distribution &&
		d.Permission.Exec == other.Permission.Exec &&
		d.Permission.Grpc == other.Permission.Grpc &&
		d.Permission.Images == other.Permission.Images &&
		d.Permission.Networks == other.Permission.Networks &&
		d.Permission.Nodes == other.Permission.Nodes &&
		d.Permission.Plugins == other.Permission.Plugins &&
		d.Permission.Services == other.Permission.Services &&
		d.Permission.Session == other.Permission.Session &&
		d.Permission.Swarm == other.Permission.Swarm &&
		d.Permission.System == other.Permission.System &&
		d.Permission.Tasks == other.Permission.Tasks &&
		d.Permission.Volumes == other.Permission.Volumes
}

func (c *ApplicationCustomHealthCheck) Equal(other *ApplicationCustomHealthCheck) bool {
	return c.Enabled == other.Enabled &&
		strings.Compare(c.TestCommand, other.TestCommand) == 0 &&
		c.IntervalSeconds == other.IntervalSeconds &&
		c.TimeoutSeconds == other.TimeoutSeconds &&
		c.StartPeriodSeconds == other.StartPeriodSeconds &&
		c.StartIntervalSeconds == other.StartIntervalSeconds &&
		c.Retries == other.Retries
}

func (application *Application) DockerProxyServiceName() string {
	return application.ID + "-dp"
}
