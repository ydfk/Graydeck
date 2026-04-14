package manager

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var zashboardAssetVersionPattern = regexp.MustCompile("=L\\((?:\"|'|`)(v?\\d+\\.\\d+\\.\\d+(?:[-+][0-9A-Za-z.-]+)?)(?:\"|'|`)\\)")
var semanticVersionPattern = regexp.MustCompile(`v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?`)

func (s *Service) detectInstalledZashboardVersion() string {
	if version := detectZashboardVersionFromAssets(s.zashboardRoot()); version != "" {
		return version
	}

	candidates := []string{
		filepath.Join(s.zashboardRoot(), "version.txt"),
		filepath.Join(s.zashboardDir(), "version.txt"),
	}

	for _, path := range candidates {
		if version := readTrimmedFile(path); version != "" {
			return version
		}
	}

	return ""
}

func detectZashboardVersionFromAssets(root string) string {
	files, err := filepath.Glob(filepath.Join(root, "assets", "*.js"))
	if err != nil {
		return ""
	}

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		match := zashboardAssetVersionPattern.FindSubmatch(content)
		if len(match) == 2 {
			return strings.TrimSpace(string(match[1]))
		}
	}

	return ""
}

func readTrimmedFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}

func sameVersion(current, latest string) bool {
	normalizedCurrent := normalizeVersion(current)
	if normalizedCurrent == "" {
		return false
	}

	return normalizedCurrent == normalizeVersion(latest)
}

func normalizeVersion(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	return strings.TrimPrefix(trimmed, "v")
}

func detectVersionFromText(value string) string {
	match := semanticVersionPattern.FindString(strings.TrimSpace(value))
	return strings.TrimSpace(match)
}

func detectVersionFromName(value string) string {
	return detectVersionFromText(filepath.Base(value))
}

func (s *Service) detectInstalledCoreVersion() string {
	if version := s.detectCoreVersionFromBinary(); version != "" {
		return version
	}

	if version := readTrimmedFile(filepath.Join(s.coreDir(), "version.txt")); version != "" && fileExists(s.coreExecutablePath()) {
		return version
	}

	return ""
}

func (s *Service) detectCoreVersionFromBinary() string {
	if !fileExists(s.coreExecutablePath()) {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, s.coreExecutablePath(), "-v")
	output, err := command.CombinedOutput()
	if err != nil && len(output) == 0 {
		return ""
	}

	return detectVersionFromText(string(output))
}
