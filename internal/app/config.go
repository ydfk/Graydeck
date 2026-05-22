package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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
		ListenAddress:     loadListenAddress(appConfigPath),
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

func loadListenAddress(appConfigPath string) string {
	const fallbackPort = "8080"

	content, err := os.ReadFile(appConfigPath)
	if err != nil {
		return ":" + fallbackPort
	}

	if port := parseServerPort(string(content)); port != "" {
		return ":" + port
	}

	return ":" + fallbackPort
}

func parseServerPort(content string) string {
	section := ""

	for _, rawLine := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if !strings.HasPrefix(rawLine, " ") && strings.HasSuffix(trimmed, ":") {
			section = strings.TrimSuffix(trimmed, ":")
			continue
		}

		if section != "server" || !strings.HasPrefix(trimmed, "port:") {
			continue
		}

		value := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "port:")), `"'`)
		if validPort(value) {
			return value
		}
	}

	return ""
}

func validPort(value string) bool {
	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return false
	}

	return port > 0
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
