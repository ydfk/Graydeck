package app

import (
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ListenAddress     string
	DataDir           string
	CoreTargetOS      string
	CoreTargetArch    string
	ControllerAddr    string
	RuntimeMixedPort  string
	RuntimeSocksPort  string
	RuntimeRedirPort  string
	RuntimeTProxyPort string
	RuntimeSecret     string
	BaseConfigPath    string
	AppConfigPath     string
	WebRoot           string
}

func LoadConfigFromEnv() Config {
	listenAddress := ":18080"

	dataDir := defaultDataDir()
	configDir := defaultConfigDir()
	webRoot := defaultWebRoot()
	coreTargetOS := runtime.GOOS
	coreTargetArch := runtime.GOARCH
	controllerAddr := "127.0.0.1:19090"
	runtimeMixedPort := "7890"
	runtimeSocksPort := "7891"
	runtimeRedirPort := "7892"
	runtimeTProxyPort := "7893"

	runtimeSecret := os.Getenv("GRAYDECK_SECRET")
	if runtimeSecret == "" {
		runtimeSecret = "graydeck-secret"
	}

	baseConfigPath := os.Getenv("GRAYDECK_BASE_CONFIG")
	if baseConfigPath == "" {
		baseConfigPath = filepath.Join(configDir, "base.yaml")
	}

	appConfigPath := os.Getenv("GRAYDECK_APP_CONFIG")
	if appConfigPath == "" {
		appConfigPath = filepath.Join(configDir, "graydeck.yaml")
	}

	return Config{
		ListenAddress:     listenAddress,
		DataDir:           dataDir,
		CoreTargetOS:      coreTargetOS,
		CoreTargetArch:    coreTargetArch,
		ControllerAddr:    controllerAddr,
		RuntimeMixedPort:  runtimeMixedPort,
		RuntimeSocksPort:  runtimeSocksPort,
		RuntimeRedirPort:  runtimeRedirPort,
		RuntimeTProxyPort: runtimeTProxyPort,
		RuntimeSecret:     runtimeSecret,
		BaseConfigPath:    baseConfigPath,
		AppConfigPath:     appConfigPath,
		WebRoot:           webRoot,
	}
}

func defaultDataDir() string {
	const containerDataDir = "/data"
	if info, err := os.Stat(containerDataDir); err == nil && info.IsDir() {
		return containerDataDir
	}

	return filepath.Join(".", "data")
}

func defaultWebRoot() string {
	const containerWebRoot = "/opt/graydeck/web"
	if info, err := os.Stat(containerWebRoot); err == nil && info.IsDir() {
		return containerWebRoot
	}

	return filepath.Join(".", "web", "dist")
}

func defaultConfigDir() string {
	const containerConfigDir = "/config"
	if info, err := os.Stat(containerConfigDir); err == nil && info.IsDir() {
		return containerConfigDir
	}

	return filepath.Join(".", "config")
}
