package app

import (
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ListenAddress   string
	DataDir         string
	CoreTargetOS    string
	CoreTargetArch  string
	ControllerAddr  string
	RuntimeMixedPort string
	RuntimeSecret   string
	BaseConfigPath  string
}

func LoadConfigFromEnv() Config {
	listenAddress := os.Getenv("MGR_LISTEN")
	if listenAddress == "" {
		listenAddress = ":18080"
	}

	dataDir := os.Getenv("GRAYDECK_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(".", "data")
	}

	coreTargetOS := os.Getenv("GRAYDECK_CORE_OS")
	if coreTargetOS == "" {
		coreTargetOS = runtime.GOOS
	}

	coreTargetArch := os.Getenv("GRAYDECK_CORE_ARCH")
	if coreTargetArch == "" {
		coreTargetArch = runtime.GOARCH
	}

	controllerAddr := os.Getenv("GRAYDECK_CONTROLLER_ADDR")
	if controllerAddr == "" {
		controllerAddr = "127.0.0.1:19090"
	}

	runtimeMixedPort := os.Getenv("GRAYDECK_MIXED_PORT")
	if runtimeMixedPort == "" {
		runtimeMixedPort = "7890"
	}

	runtimeSecret := os.Getenv("GRAYDECK_SECRET")
	if runtimeSecret == "" {
		runtimeSecret = "graydeck-secret"
	}

	baseConfigPath := os.Getenv("GRAYDECK_BASE_CONFIG")
	if baseConfigPath == "" {
		baseConfigPath = filepath.Join(dataDir, "runtime", "base.yaml")
	}

	return Config{
		ListenAddress:    listenAddress,
		DataDir:          dataDir,
		CoreTargetOS:     coreTargetOS,
		CoreTargetArch:   coreTargetArch,
		ControllerAddr:   controllerAddr,
		RuntimeMixedPort: runtimeMixedPort,
		RuntimeSecret:    runtimeSecret,
		BaseConfigPath:   baseConfigPath,
	}
}
