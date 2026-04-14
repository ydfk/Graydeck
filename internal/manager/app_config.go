package manager

import (
	"crypto/subtle"
	"os"
	"path/filepath"
	"strings"
)

type appSettings struct {
	authEnabled           bool
	authUsername          string
	authPassword          string
	updatePreferProxy     bool
	updateProxyURL        string
	zashboardHideSettings bool
}

func (s *Service) ensureAppConfig() error {
	if fileExists(s.cfg.AppConfigPath) {
		return nil
	}

	if err := s.migrateLegacyAppConfig(); err != nil {
		return err
	}

	if fileExists(s.cfg.AppConfigPath) {
		return nil
	}

	return os.WriteFile(s.cfg.AppConfigPath, []byte(defaultAppConfigContent()), 0o644)
}

func (s *Service) migrateLegacyAppConfig() error {
	legacyPath := filepath.Join(s.cfg.DataDir, "config", "graydeck.yaml")
	if legacyPath == s.cfg.AppConfigPath || !fileExists(legacyPath) {
		return nil
	}

	content, err := os.ReadFile(legacyPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.cfg.AppConfigPath, content, 0o644); err != nil {
		return err
	}

	return os.Remove(legacyPath)
}

func (s *Service) loadAppConfigStatus() {
	settings := s.loadAppSettings()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.ZashboardHideSettings = settings.zashboardHideSettings
}

func (s *Service) AuthEnabled() bool {
	return s.loadAppSettings().authEnabled
}

func (s *Service) UpdateProxyURL() string {
	return strings.TrimSpace(s.loadAppSettings().updateProxyURL)
}

func (s *Service) PreferProxyForUpdate() bool {
	return s.loadAppSettings().updatePreferProxy
}

func (s *Service) ValidateCredentials(username, password string) bool {
	settings := s.loadAppSettings()
	if !settings.authEnabled {
		return true
	}

	expectedUsername := strings.TrimSpace(settings.authUsername)
	actualUsername := strings.TrimSpace(username)

	if subtle.ConstantTimeCompare([]byte(actualUsername), []byte(expectedUsername)) != 1 {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(password), []byte(settings.authPassword)) == 1
}

func (s *Service) loadAppSettings() appSettings {
	settings := appSettings{
		authEnabled:           true,
		authUsername:          "admin",
		authPassword:          "admin123",
		updatePreferProxy:     true,
		updateProxyURL:        "https://ghfast.top/",
		zashboardHideSettings: true,
	}

	content, err := os.ReadFile(s.cfg.AppConfigPath)
	if err != nil {
		return settings
	}

	return parseAppSettings(string(content), settings)
}

func defaultAppConfigContent() string {
	return strings.TrimSpace(`
# Graydeck 服务配置
auth:
  enabled: true
  username: admin
  password: admin123
update:
  prefer-proxy: true
  proxy-url: https://ghfast.top/
zashboard:
  hide-settings: true
`) + "\n"
}

func parseAppSettings(content string, fallback appSettings) appSettings {
	settings := fallback
	section := ""

	for _, rawLine := range strings.Split(content, "\n") {
		if strings.TrimSpace(rawLine) == "" || strings.HasPrefix(strings.TrimSpace(rawLine), "#") {
			continue
		}

		trimmed := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(rawLine, " ") && strings.HasSuffix(trimmed, ":") {
			section = strings.TrimSuffix(trimmed, ":")
			continue
		}

		switch section {
		case "auth":
			switch {
			case strings.HasPrefix(trimmed, "enabled:"):
				value := strings.ToLower(parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "enabled:"))))
				if value == "true" {
					settings.authEnabled = true
				}
				if value == "false" {
					settings.authEnabled = false
				}
			case strings.HasPrefix(trimmed, "username:"):
				settings.authUsername = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "username:")))
			case strings.HasPrefix(trimmed, "password:"):
				settings.authPassword = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "password:")))
			}
		case "update":
			switch {
			case strings.HasPrefix(trimmed, "prefer-proxy:"):
				value := strings.ToLower(parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "prefer-proxy:"))))
				if value == "true" {
					settings.updatePreferProxy = true
				}
				if value == "false" {
					settings.updatePreferProxy = false
				}
			case strings.HasPrefix(trimmed, "proxy-url:"):
				settings.updateProxyURL = parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "proxy-url:")))
			}
		case "zashboard":
			if !strings.HasPrefix(trimmed, "hide-settings:") {
				continue
			}

			value := strings.ToLower(parseYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "hide-settings:"))))
			if value == "true" {
				settings.zashboardHideSettings = true
			}
			if value == "false" {
				settings.zashboardHideSettings = false
			}
		}
	}

	return settings
}
