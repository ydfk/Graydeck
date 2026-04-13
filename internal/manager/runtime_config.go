package manager

import (
	"fmt"
	"os"
	"strings"
)

type managedRuntimeValues struct {
	mixedPort     string
	controller    string
	secret        string
}

func (s *Service) ensureBaseConfig() error {
	if fileExists(s.cfg.BaseConfigPath) {
		return nil
	}

	return os.WriteFile(s.cfg.BaseConfigPath, []byte(s.defaultBaseConfigContent()), 0o644)
}

func (s *Service) defaultBaseConfigContent() string {
	return strings.TrimSpace(fmt.Sprintf(`
# Graydeck 自动注入的基础配置
mixed-port: %s
external-controller: %s
secret: %s
allow-lan: false
mode: rule
log-level: info
`, yamlScalar(s.cfg.RuntimeMixedPort), yamlString(s.cfg.ControllerAddr), yamlString(s.cfg.RuntimeSecret))) + "\n"
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
		"external-controller",
		"secret",
		"allow-lan",
		"mode",
		"log-level",
	}

	values := parseManagedRuntimeValues(string(baseConfig), managedRuntimeValues{
		mixedPort:  s.cfg.RuntimeMixedPort,
		controller: s.cfg.ControllerAddr,
		secret:     s.cfg.RuntimeSecret,
	})

	managedSection := strings.TrimSpace(fmt.Sprintf(`
mixed-port: %s
external-controller: %s
secret: %s
`, yamlScalar(values.mixedPort), yamlString(values.controller), yamlString(values.secret)))

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

func isDigitsOnly(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
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
		case strings.HasPrefix(trimmed, "external-controller:"):
			values.controller = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "external-controller:")))
		case strings.HasPrefix(trimmed, "secret:"):
			values.secret = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "secret:")))
		}
	}

	return values
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
