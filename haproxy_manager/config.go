package haproxymanager

func isPortRestrictedForManualConfig(port int) bool {
	return port == 80 || port == 443
}