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
	WebRoot           string
}

func LoadConfigFromEnv() Config {
	listenAddress := ":18080"

	dataDir := defaultDataDir()
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
		baseConfigPath = filepath.Join(dataDir, "runtime", "base.yaml")
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
