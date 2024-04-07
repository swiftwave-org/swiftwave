package local_config

import "path/filepath"

var defaultBindAddress = "0.0.0.0"
var defaultBindPort = 3333
var defaultImageRegistryPort = 3334
var defaultSocketPathDirectory = "/var/run/swiftwave"
var defaultDataDirectory = "/var/lib/swiftwave"
var defaultNetworkName = "swiftwave_network"
var defaultHAProxyServiceName = "swiftwave_haproxy"
var defaultHAProxyUnixSocketDirectory = filepath.Join(defaultSocketPathDirectory, "haproxy")
var defaultHAProxyUnixSocketPath = filepath.Join(defaultHAProxyUnixSocketDirectory, "dataplaneapi.sock")
var defaultHAProxyDataDirectoryPath = filepath.Join(defaultDataDirectory, "haproxy")
var defaultUDPProxyServiceName = "swiftwave_udpproxy"
var defaultUDPProxyDataDirectoryPath = filepath.Join(defaultDataDirectory, "udpproxy")
var defaultUDPProxyUnixSocketDirectory = filepath.Join(defaultSocketPathDirectory, "udpproxy")
var defaultUDPProxyUnixSocketPath = filepath.Join(defaultUDPProxyUnixSocketDirectory, "api.sock")
var defaultSSLCertDirectoryPath = filepath.Join(defaultDataDirectory, "cert")
var defaultLocalImageRegistryDirectoryPath = filepath.Join(defaultDataDirectory, "registry")
var defaultLocalImageRegistryDataDirectoryPath = filepath.Join(defaultLocalImageRegistryDirectoryPath, "data")
var defaultLocalImageRegistryAuthDirectoryPath = filepath.Join(defaultLocalImageRegistryDirectoryPath, "auth")
var defaultLocalImageRegistryCertDirectoryPath = defaultSSLCertDirectoryPath
var defaultPVBackupDirectoryPath = filepath.Join(defaultDataDirectory, "pvbackup")
var defaultPVRestoreDirectoryPath = filepath.Join(defaultDataDirectory, "pvrestore")
var defaultTarballDirectoryPath = filepath.Join(defaultDataDirectory, "tarball")
var defaultLocalPostgresDataDirectory = filepath.Join(defaultDataDirectory, "postgres")
var LocalConfigPath = filepath.Join(defaultDataDirectory, "config.yml")
var LogDirectoryPath = "/var/log/swiftwave"
var InfoLogFilePath = filepath.Join(LogDirectoryPath, "swiftwave.log")
var ErrorLogFilePath = filepath.Join(LogDirectoryPath, "swiftwave.error.log")
var defaultSSHTimeout = 10
