package manager

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

func (s *Service) refreshCoreMetadata(ctx context.Context) error {
	release, err := s.fetchLatestRelease(ctx, "MetaCubeX", "mihomo")
	if err != nil {
		return err
	}

	asset, err := s.selectCoreAsset(release)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.status.CoreLatestVersion = release.TagName
	s.syncInstallStateLocked()
	s.mu.Unlock()

	_ = asset
	return nil
}

func (s *Service) ensureCoreInstalled(ctx context.Context) error {
	if _, err := os.Stat(s.coreExecutablePath()); err == nil {
		s.mu.Lock()
		s.status.CoreExecutableReady = true
		s.mu.Unlock()
		return nil
	}

	return s.downloadCore(ctx)
}

func (s *Service) downloadCore(ctx context.Context) error {
	release, err := s.fetchLatestRelease(ctx, "MetaCubeX", "mihomo")
	if err != nil {
		return err
	}

	asset, err := s.selectCoreAsset(release)
	if err != nil {
		return err
	}

	downloadPath := filepath.Join(s.coreDir(), asset.Name)
	if err := s.downloadFile(ctx, asset, downloadPath); err != nil {
		return err
	}

	if strings.HasSuffix(asset.Name, ".gz") {
		if err := extractGzip(downloadPath, s.coreExecutablePath()); err != nil {
			return err
		}
	} else if strings.HasSuffix(asset.Name, ".zip") {
		if err := extractZipExecutable(downloadPath, s.coreExecutablePath()); err != nil {
			return err
		}
	} else {
		if err := copyFile(downloadPath, s.coreExecutablePath()); err != nil {
			return err
		}
	}

	if err := os.Chmod(s.coreExecutablePath(), 0o755); err != nil {
		return err
	}

	s.mu.Lock()
	s.status.CoreVersion = release.TagName
	s.status.CoreLatestVersion = release.TagName
	s.syncInstallStateLocked()
	s.mu.Unlock()

	if err := os.WriteFile(filepath.Join(s.coreDir(), "version.txt"), []byte(release.TagName+"\n"), 0o644); err != nil {
		return err
	}

	return nil
}

func (s *Service) refreshZashboardMetadata(ctx context.Context) error {
	release, err := s.fetchLatestRelease(ctx, "Zephyruso", "zashboard")
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.status.ZashboardLatestVersion = release.TagName
	s.status.ZashboardError = ""
	s.syncInstallStateLocked()
	s.mu.Unlock()
	return nil
}

func (s *Service) ensureZashboardInstalled(ctx context.Context) error {
	if _, err := os.Stat(filepath.Join(s.zashboardRoot(), "index.html")); err == nil {
		installedVersion := s.detectInstalledZashboardVersion()

		s.mu.Lock()
		s.status.ZashboardVersion = installedVersion
		s.status.ZashboardReady = true
		s.status.ZashboardError = ""
		s.syncInstallStateLocked()
		s.mu.Unlock()
		return nil
	}

	return s.downloadZashboard(ctx)
}

func (s *Service) downloadZashboard(ctx context.Context) error {
	release, err := s.fetchLatestRelease(ctx, "Zephyruso", "zashboard")
	if err != nil {
		return err
	}

	asset, err := s.selectZashboardAsset(release)
	if err != nil {
		return err
	}

	archivePath := filepath.Join(s.zashboardDir(), asset.Name)
	if err := s.downloadFile(ctx, asset, archivePath); err != nil {
		return err
	}

	if err := clearDirExcept(s.zashboardDir(), asset.Name); err != nil {
		return err
	}

	if err := extractZipToDir(archivePath, s.zashboardDir()); err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(s.zashboardRoot(), "index.html")); err != nil {
		return errors.New("zashboard 资源校验失败：缺少 index.html")
	}

	installedVersion := s.detectInstalledZashboardVersion()
	if installedVersion == "" {
		installedVersion = release.TagName
	}

	s.mu.Lock()
	s.status.ZashboardVersion = installedVersion
	s.status.ZashboardLatestVersion = release.TagName
	s.status.ZashboardError = ""
	s.syncInstallStateLocked()
	s.mu.Unlock()

	if err := os.WriteFile(filepath.Join(s.zashboardDir(), "version.txt"), []byte(installedVersion+"\n"), 0o644); err != nil {
		return err
	}

	return nil
}

func (s *Service) fetchLatestRelease(ctx context.Context, owner, repo string) (releaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return releaseInfo{}, err
	}

	request.Header.Set("User-Agent", "Graydeck")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return releaseInfo{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return releaseInfo{}, fmt.Errorf("获取发布信息失败：%s", response.Status)
	}

	var release releaseInfo
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return releaseInfo{}, err
	}

	return release, nil
}

func (s *Service) selectCoreAsset(release releaseInfo) (releaseAsset, error) {
	prefix := fmt.Sprintf("mihomo-%s-%s-", s.cfg.CoreTargetOS, s.cfg.CoreTargetArch)

	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, prefix) && !strings.Contains(asset.Name, "-go") && !strings.Contains(asset.Name, "compatible") {
			return asset, nil
		}
	}

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, s.cfg.CoreTargetOS) && strings.Contains(asset.Name, s.cfg.CoreTargetArch) {
			return asset, nil
		}
	}

	return releaseAsset{}, errors.New("未找到匹配当前平台的核心文件")
}

func (s *Service) selectZashboardAsset(release releaseInfo) (releaseAsset, error) {
	for _, asset := range release.Assets {
		if asset.Name == "dist-no-fonts.zip" {
			return asset, nil
		}
	}

	for _, asset := range release.Assets {
		if asset.Name == "dist.zip" {
			return asset, nil
		}
	}

	return releaseAsset{}, errors.New("未找到 zashboard 静态资源包")
}

func (s *Service) downloadFile(ctx context.Context, asset releaseAsset, outputPath string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}

	request.Header.Set("User-Agent", "Graydeck")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败：%s", response.Status)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)
	if _, err := io.Copy(writer, response.Body); err != nil {
		return err
	}

	if strings.HasPrefix(asset.Digest, "sha256:") {
		expected := strings.TrimPrefix(asset.Digest, "sha256:")
		actual := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(expected, actual) {
			return fmt.Errorf("校验失败：期望 %s，实际 %s", expected, actual)
		}
	}

	return nil
}

func (s *Service) fetchText(ctx context.Context, rawURL string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("User-Agent", "Graydeck")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败：%s", response.Status)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func extractGzip(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	reader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	target, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, reader)
	return err
}

func extractZipExecutable(sourcePath, targetPath string) error {
	archive, err := zip.OpenReader(sourcePath)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		if file.FileInfo().IsDir() {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}

		target, err := os.Create(targetPath)
		if err != nil {
			reader.Close()
			return err
		}

		_, copyErr := io.Copy(target, reader)
		closeErr := reader.Close()
		targetErr := target.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if targetErr != nil {
			return targetErr
		}
		return nil
	}

	return errors.New("压缩包中未找到可执行文件")
}

func extractZipToDir(sourcePath, outputDir string) error {
	archive, err := zip.OpenReader(sourcePath)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		targetPath := filepath.Join(outputDir, file.Name)
		if !strings.HasPrefix(targetPath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			return errors.New("非法压缩包路径")
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}

		target, err := os.Create(targetPath)
		if err != nil {
			reader.Close()
			return err
		}

		if _, err := io.Copy(target, reader); err != nil {
			reader.Close()
			target.Close()
			return err
		}

		reader.Close()
		target.Close()
	}

	return nil
}
