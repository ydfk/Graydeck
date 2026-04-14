package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mihomo-manager/internal/model"
)

type managedRuntimeValues struct {
	mixedPort  string
	socksPort  string
	redirPort  string
	tproxyPort string
	bindAddr   string
	allowLAN   string
	mode       string
	logLevel   string
	controller string
	secret     string
}

func (s *Service) ensureBaseConfig() error {
	if fileExists(s.cfg.BaseConfigPath) {
		return nil
	}

	if err := s.migrateLegacyBaseConfig(); err != nil {
		return err
	}

	if fileExists(s.cfg.BaseConfigPath) {
		return nil
	}

	return os.WriteFile(s.cfg.BaseConfigPath, []byte(s.defaultBaseConfigContent()), 0o644)
}

func (s *Service) migrateLegacyBaseConfig() error {
	legacyPaths := []string{
		filepath.Join(s.cfg.DataDir, "config", "base.yaml"),
		filepath.Join(s.runtimeDir(), "base.yaml"),
	}

	for _, legacyPath := range legacyPaths {
		if legacyPath == s.cfg.BaseConfigPath || !fileExists(legacyPath) {
			continue
		}

		content, err := os.ReadFile(legacyPath)
		if err != nil {
			return err
		}

		if err := os.WriteFile(s.cfg.BaseConfigPath, content, 0o644); err != nil {
			return err
		}

		return os.Remove(legacyPath)
	}

	return nil
}

func (s *Service) defaultBaseConfigContent() string {
	return strings.TrimSpace(fmt.Sprintf(`
# Graydeck 自动注入的基础配置
mixed-port: %s
socks-port: %s
redir-port: %s
tproxy-port: %s
bind-address: %s
allow-lan: %s
external-controller: %s
secret: %s
mode: %s
log-level: %s
`, yamlScalar(s.cfg.RuntimeMixedPort), yamlScalar(s.cfg.RuntimeSocksPort), yamlScalar(s.cfg.RuntimeRedirPort), yamlScalar(s.cfg.RuntimeTProxyPort), yamlString("0.0.0.0"), yamlBool("true"), yamlString(s.cfg.ControllerAddr), yamlString(s.cfg.RuntimeSecret), yamlString("rule"), yamlString("info"))) + "\n"
}

func (s *Service) writeRuntimeConfig(sourcePath, targetPath string) error {
	baseConfig, err := os.ReadFile(s.cfg.BaseConfigPath)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	managedKeys := []string{
		"mixed-port",
		"port",
		"socks-port",
		"redir-port",
		"tproxy-port",
		"bind-address",
		"allow-lan",
		"external-controller",
		"secret",
		"mode",
		"log-level",
	}

	values := parseManagedRuntimeValues(string(baseConfig), managedRuntimeValues{
		mixedPort:  s.cfg.RuntimeMixedPort,
		socksPort:  s.cfg.RuntimeSocksPort,
		redirPort:  s.cfg.RuntimeRedirPort,
		tproxyPort: s.cfg.RuntimeTProxyPort,
		bindAddr:   "0.0.0.0",
		allowLAN:   "true",
		mode:       "rule",
		logLevel:   "info",
		controller: s.cfg.ControllerAddr,
		secret:     s.cfg.RuntimeSecret,
	})

	managedSection := strings.TrimSpace(fmt.Sprintf(`
mixed-port: %s
socks-port: %s
redir-port: %s
tproxy-port: %s
bind-address: %s
allow-lan: %s
external-controller: %s
secret: %s
mode: %s
log-level: %s
`, yamlScalar(values.mixedPort), yamlScalar(values.socksPort), yamlScalar(values.redirPort), yamlScalar(values.tproxyPort), yamlString(values.bindAddr), yamlBool(values.allowLAN), yamlString(values.controller), yamlString(values.secret), yamlString(values.mode), yamlString(values.logLevel)))

	baseExtraSection := strings.TrimSpace(stripTopLevelKeys(string(baseConfig), managedKeys...))
	subscriptionSection := strings.TrimSpace(stripTopLevelKeys(string(content), managedKeys...))
	mergedSections := []string{managedSection}
	if baseExtraSection != "" {
		mergedSections = append(mergedSections, baseExtraSection)
	}
	if subscriptionSection != "" {
		mergedSections = append(mergedSections, subscriptionSection)
	}

	merged := strings.Join(mergedSections, "\n\n") + "\n"
	return os.WriteFile(targetPath, []byte(merged), 0o644)
}

func (s *Service) buildValidationConfig(id string) (string, func(), error) {
	tempFile, err := os.CreateTemp(s.runtimeDir(), "validate-*.yaml")
	if err != nil {
		return "", nil, err
	}

	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", nil, err
	}

	if err := s.writeRuntimeConfig(s.subscriptionPreviewPath(id), tempPath); err != nil {
		_ = os.Remove(tempPath)
		return "", nil, err
	}

	return tempPath, func() {
		_ = os.Remove(tempPath)
	}, nil
}

func (s *Service) DefaultRuntimeConfig() model.DefaultRuntimeConfig {
	values := s.loadManagedRuntimeValues()
	return model.DefaultRuntimeConfig{
		Path:       s.cfg.BaseConfigPath,
		MixedPort:  values.mixedPort,
		SocksPort:  values.socksPort,
		RedirPort:  values.redirPort,
		TProxyPort: values.tproxyPort,
	}
}

func (s *Service) UpdateDefaultRuntimeConfig(values model.DefaultRuntimeConfig) error {
	updated := managedRuntimeValues{
		mixedPort:  strings.TrimSpace(values.MixedPort),
		socksPort:  strings.TrimSpace(values.SocksPort),
		redirPort:  strings.TrimSpace(values.RedirPort),
		tproxyPort: strings.TrimSpace(values.TProxyPort),
	}

	if err := validatePortValue(updated.mixedPort, true); err != nil {
		return fmt.Errorf("mixed-port %v", err)
	}
	if err := validatePortValue(updated.socksPort, false); err != nil {
		return fmt.Errorf("socks-port %v", err)
	}
	if err := validatePortValue(updated.redirPort, false); err != nil {
		return fmt.Errorf("redir-port %v", err)
	}
	if err := validatePortValue(updated.tproxyPort, false); err != nil {
		return fmt.Errorf("tproxy-port %v", err)
	}

	current := s.loadManagedRuntimeValues()
	updated.bindAddr = current.bindAddr
	updated.allowLAN = current.allowLAN
	updated.mode = current.mode
	updated.logLevel = current.logLevel
	updated.controller = current.controller
	updated.secret = current.secret

	return s.saveManagedRuntimeValues(updated)
}

func stripTopLevelKeys(content string, keys ...string) string {
	lines := strings.Split(content, "\n")
	filtered := make([]string, 0, len(lines))

	for _, line := range lines {
		if shouldSkipTopLevelLine(line, keys) {
			continue
		}
		filtered = append(filtered, line)
	}

	return strings.Join(filtered, "\n")
}

func shouldSkipTopLevelLine(line string, keys []string) bool {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return false
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}

	for _, key := range keys {
		if strings.HasPrefix(trimmed, key+":") {
			return true
		}
	}

	return false
}

func yamlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func yamlScalar(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "0"
	}

	if isDigitsOnly(trimmed) {
		return trimmed
	}

	return yamlString(trimmed)
}

func yamlBool(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return "true"
	case "false":
		return "false"
	default:
		return "false"
	}
}

func isDigitsOnly(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func validatePortValue(value string, required bool) error {
	if value == "" {
		if required {
			return fmt.Errorf("不能为空")
		}
		return nil
	}

	if !isDigitsOnly(value) {
		return fmt.Errorf("必须为数字")
	}

	return nil
}

func parseManagedRuntimeValues(content string, fallback managedRuntimeValues) managedRuntimeValues {
	values := fallback

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "mixed-port:"):
			values.mixedPort = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "mixed-port:")))
		case strings.HasPrefix(trimmed, "socks-port:"):
			values.socksPort = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "socks-port:")))
		case strings.HasPrefix(trimmed, "redir-port:"):
			values.redirPort = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "redir-port:")))
		case strings.HasPrefix(trimmed, "tproxy-port:"):
			values.tproxyPort = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "tproxy-port:")))
		case strings.HasPrefix(trimmed, "bind-address:"):
			values.bindAddr = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "bind-address:")))
		case strings.HasPrefix(trimmed, "allow-lan:"):
			values.allowLAN = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "allow-lan:")))
		case strings.HasPrefix(trimmed, "external-controller:"):
			values.controller = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "external-controller:")))
		case strings.HasPrefix(trimmed, "secret:"):
			values.secret = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "secret:")))
		case strings.HasPrefix(trimmed, "mode:"):
			values.mode = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "mode:")))
		case strings.HasPrefix(trimmed, "log-level:"):
			values.logLevel = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "log-level:")))
		}
	}

	return values
}

func (s *Service) saveManagedRuntimeValues(values managedRuntimeValues) error {
	currentContent, err := os.ReadFile(s.cfg.BaseConfigPath)
	if err != nil {
		return err
	}

	managedKeys := []string{
		"mixed-port",
		"socks-port",
		"redir-port",
		"tproxy-port",
		"bind-address",
		"allow-lan",
		"external-controller",
		"secret",
		"mode",
		"log-level",
	}

	managedSection := strings.TrimSpace(fmt.Sprintf(`
mixed-port: %s
socks-port: %s
redir-port: %s
tproxy-port: %s
bind-address: %s
allow-lan: %s
external-controller: %s
secret: %s
mode: %s
log-level: %s
`, yamlScalar(values.mixedPort), yamlScalar(values.socksPort), yamlScalar(values.redirPort), yamlScalar(values.tproxyPort), yamlString(values.bindAddr), yamlBool(values.allowLAN), yamlString(values.controller), yamlString(values.secret), yamlString(values.mode), yamlString(values.logLevel)))

	extraSection := strings.TrimSpace(stripTopLevelKeys(string(currentContent), managedKeys...))
	sections := []string{managedSection}
	if extraSection != "" {
		sections = append(sections, extraSection)
	}

	if err := os.WriteFile(s.cfg.BaseConfigPath, []byte(strings.Join(sections, "\n\n")+"\n"), 0o644); err != nil {
		return err
	}

	s.loadRuntimeConfigStatus()
	return nil
}

func parseYAMLScalar(value string) string {
	if len(value) >= 2 {
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			return strings.ReplaceAll(value[1:len(value)-1], "''", "'")
		}

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			return value[1 : len(value)-1]
		}
	}

	return value
}
