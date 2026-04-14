package manager

import (
	"os"
	"path/filepath"
	"strings"
)

type appSettings struct {
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

func (s *Service) loadAppSettings() appSettings {
	settings := appSettings{
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
zashboard:
  hide-settings: true
`) + "\n"
}

func parseAppSettings(content string, fallback appSettings) appSettings {
	settings := fallback

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "hide-settings:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "hide-settings:"))
			switch strings.ToLower(value) {
			case "true":
				settings.zashboardHideSettings = true
			case "false":
				settings.zashboardHideSettings = false
			}
		}
	}

	return settings
}
