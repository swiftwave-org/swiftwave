package core

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func (config *ServiceConfig) Load() {
	// Initialize Service Config
	serverPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("PORT environment variable is not set")
	}
	config.Port = serverPort
	config.CodeTarballDir = os.Getenv("CODE_TARBALL_DIR")
	config.SwarmNetwork = os.Getenv("SWARM_NETWORK")
	config.HaproxyService = os.Getenv("HAPROXY_SERVICE_NAME")
	config.Environment = os.Getenv("ENVIRONMENT")
	if config.Environment == "" {
		config.Environment = "production"
	}
	restrictedPortsStr := os.Getenv("RESTRICTED_PORTS")
	restrictedPortsStrSplit := strings.Split(restrictedPortsStr, ",")
	var restrictedPorts []int
	for _, port := range restrictedPortsStrSplit {
		portInt, err := strconv.Atoi(string(port))
		if err != nil {
			panic(err)
		}
		restrictedPorts = append(restrictedPorts, portInt)
	}
	config.RestrictedPorts = restrictedPorts
	tokenExpiryMinutes, err := strconv.Atoi(os.Getenv("SESSION_TOKEN_EXPIRY_MINUTES"))
	if err != nil {
		panic(err)
	}
	config.SessionTokens = make(map[string]time.Time)
	config.SessionTokenExpiryMinutes = tokenExpiryMinutes
}
