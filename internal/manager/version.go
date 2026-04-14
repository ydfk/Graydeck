package manager

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var zashboardAssetVersionPattern = regexp.MustCompile("=L\\((?:\"|'|`)(v?\\d+\\.\\d+\\.\\d+(?:[-+][0-9A-Za-z.-]+)?)(?:\"|'|`)\\)")

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
