package local_config

import "path/filepath"

var defaultBindAddress = "0.0.0.0"
var defaultBindPort = 3333
var defaultSocketPathDirectory = "/var/run/swiftwave"
var defaultDataDirectory = "/var/lib/swiftwave"
var defaultNetworkName = "swiftwave_network"
var defaultHAProxyServiceName = "swiftwave_haproxy"
var defaultHAProxyUnixSocketPath = filepath.Join(defaultSocketPathDirectory, "haproxy", "dataplaneapi.sock")
var defaultHAProxyDataDirectoryPath = filepath.Join(defaultDataDirectory, "haproxy")
var defaultUDPProxyServiceName = "swiftwave_udpproxy"
var defaultUDPProxyDataDirectoryPath = filepath.Join(defaultSocketPathDirectory, "udpproxy", "api.sock")
var defaultSSLCertDirectoryPath = filepath.Join(defaultDataDirectory, "certs")
var LocalConfigPath = filepath.Join(defaultDataDirectory, "config.yml")
