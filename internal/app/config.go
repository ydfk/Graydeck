package app

import "os"

type Config struct {
	ListenAddress string
	ZashboardMode string
}

func LoadConfigFromEnv() Config {
	listenAddress := os.Getenv("MGR_LISTEN")
	if listenAddress == "" {
		listenAddress = ":18080"
	}

	mode := os.Getenv("ZASHBOARD_MODE")
	if mode == "" {
		mode = "safe"
	}

	return Config{
		ListenAddress: listenAddress,
		ZashboardMode: mode,
	}
}
