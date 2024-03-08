package system_config

import "path/filepath"

var DefaultBindAddress = "0.0.0.0"
var DefaultBindPort = 3333
var DefaultSocketPathDirectory = "/var/run/swiftwave"
var DefaultDataDirectory = "/var/lib/swiftwave"
var DefaultNetworkName = "swiftwave_network"
var DefaultHAProxyServiceName = "swiftwave_haproxy"
var DefaultHAProxyUnixSocketPath = filepath.Join(DefaultSocketPathDirectory, "haproxy", "dataplaneapi.sock")
var DefaultHAProxyDataDirectoryPath = filepath.Join(DefaultDataDirectory, "haproxy")
var DefaultUDPProxyServiceName = "swiftwave_udpproxy"
var DefaultUDPProxyDataDirectoryPath = filepath.Join(DefaultSocketPathDirectory, "udpproxy", "api.sock")
