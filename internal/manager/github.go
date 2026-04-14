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

	"mihomo-manager/internal/model"
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
	s.appendLog("开始检查 mihomo 核心最新版本")
	release, err := s.fetchLatestRelease(ctx, "MetaCubeX", "mihomo")
	if err != nil {
		return err
	}

	if _, err := s.selectCoreAsset(release); err != nil {
		return err
	}

	s.mu.Lock()
	s.status.CoreLatestVersion = release.TagName
	s.syncInstallStateLocked()
	s.mu.Unlock()

	s.appendLogf("mihomo 核心最新版本：%s", release.TagName)
	return nil
}

func (s *Service) ensureCoreInstalled(ctx context.Context) error {
	if _, err := os.Stat(s.coreExecutablePath()); err == nil {
		version := s.detectCoreVersionFromBinary()
		if version != "" {
			s.appendLog("已检测到可用的本地 mihomo 核心")
			s.mu.Lock()
			s.status.CoreVersion = version
			s.status.CoreExecutableReady = true
			s.syncInstallStateLocked()
			s.mu.Unlock()
			return nil
		}

		s.appendLog("检测到本地 mihomo 核心不可执行，准备重新安装")
		_ = os.Remove(s.coreExecutablePath())
	}

	return s.installCoreFromAuto(ctx)
}

func (s *Service) installCoreFromAuto(ctx context.Context) error {
	s.appendLog("开始自动安装 mihomo 核心")
	release, err := s.fetchLatestRelease(ctx, "MetaCubeX", "mihomo")
	if err != nil {
		return err
	}

	asset, err := s.selectCoreAsset(release)
	if err != nil {
		return err
	}

	packagePath := filepath.Join(s.coreDir(), safeBaseName(asset.Name, "mihomo-package"))
	if err := s.downloadReleaseAsset(ctx, asset, packagePath); err != nil {
		return err
	}

	s.appendLogf("mihomo 核心下载完成：%s", asset.Name)
	return s.installCorePackage(packagePath, asset.Name, release.TagName)
}

func (s *Service) InstallCoreFromURL(ctx context.Context, rawURL string) (model.SystemStatus, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return model.SystemStatus{}, errors.New("下载地址不能为空")
	}

	s.appendLogf("开始通过地址安装 mihomo 核心：%s", trimmedURL)
	packageName := safeBaseName(trimmedURL, "mihomo-package")
	packagePath := filepath.Join(s.coreDir(), packageName)
	if err := s.downloadDirectFile(ctx, trimmedURL, packagePath); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.installCorePackage(packagePath, packageName, detectVersionFromName(packageName)); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.ensureRuntime(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) InstallCoreFromUpload(ctx context.Context, sourcePath, sourceName string) (model.SystemStatus, error) {
	s.appendLogf("开始通过上传文件安装 mihomo 核心：%s", sourceName)
	if err := s.installCorePackage(sourcePath, sourceName, detectVersionFromName(sourceName)); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.ensureRuntime(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) installCorePackage(packagePath, packageName, fallbackVersion string) error {
	s.appendLogf("开始安装 mihomo 核心包：%s", packageName)
	if isUnsupportedCorePackage(packageName) {
		return fmt.Errorf("暂不支持直接安装 %s，请使用 .gz、.zip 或可执行文件", filepath.Ext(packageName))
	}

	switch {
	case strings.HasSuffix(strings.ToLower(packageName), ".gz"):
		if err := extractGzip(packagePath, s.coreExecutablePath()); err != nil {
			return err
		}
	case strings.HasSuffix(strings.ToLower(packageName), ".zip"):
		if err := extractZipExecutable(packagePath, s.coreExecutablePath()); err != nil {
			return err
		}
	default:
		if err := copyFile(packagePath, s.coreExecutablePath()); err != nil {
			return err
		}
	}

	if err := os.Chmod(s.coreExecutablePath(), 0o755); err != nil {
		return err
	}

	version := strings.TrimSpace(fallbackVersion)
	if version == "" {
		version = s.detectCoreVersionFromBinary()
	}
	if version == "" {
		version = "unknown"
	}

	s.mu.Lock()
	s.status.CoreVersion = version
	s.status.CoreExecutableReady = true
	s.syncInstallStateLocked()
	s.mu.Unlock()

	s.appendLogf("mihomo 核心安装完成，当前版本：%s", version)
	return os.WriteFile(filepath.Join(s.coreDir(), "version.txt"), []byte(version+"\n"), 0o644)
}

func (s *Service) refreshZashboardMetadata(ctx context.Context) error {
	s.appendLog("开始检查 Zashboard 最新版本")
	release, err := s.fetchLatestRelease(ctx, "Zephyruso", "zashboard")
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.status.ZashboardLatestVersion = release.TagName
	s.status.ZashboardError = ""
	s.syncInstallStateLocked()
	s.mu.Unlock()
	s.appendLogf("Zashboard 最新版本：%s", release.TagName)
	return nil
}

func (s *Service) ensureZashboardInstalled(ctx context.Context) error {
	if _, err := os.Stat(filepath.Join(s.zashboardRoot(), "index.html")); err == nil {
		installedVersion := s.detectInstalledZashboardVersion()
		s.appendLog("已检测到本地 Zashboard 资源")

		s.mu.Lock()
		s.status.ZashboardVersion = installedVersion
		s.status.ZashboardReady = true
		s.status.ZashboardError = ""
		s.syncInstallStateLocked()
		s.mu.Unlock()
		return nil
	}

	return s.installZashboardFromAuto(ctx)
}

func (s *Service) installZashboardFromAuto(ctx context.Context) error {
	s.appendLog("开始自动安装 Zashboard")
	release, err := s.fetchLatestRelease(ctx, "Zephyruso", "zashboard")
	if err != nil {
		return err
	}

	asset, err := s.selectZashboardAsset(release)
	if err != nil {
		return err
	}

	archivePath := filepath.Join(s.zashboardDir(), safeBaseName(asset.Name, "zashboard.zip"))
	if err := s.downloadReleaseAsset(ctx, asset, archivePath); err != nil {
		return err
	}

	s.appendLogf("Zashboard 资源下载完成：%s", asset.Name)
	return s.installZashboardArchive(archivePath, asset.Name, release.TagName)
}

func (s *Service) InstallZashboardFromURL(ctx context.Context, rawURL string) (model.SystemStatus, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return model.SystemStatus{}, errors.New("下载地址不能为空")
	}

	s.appendLogf("开始通过地址安装 Zashboard：%s", trimmedURL)
	archiveName := safeBaseName(trimmedURL, "zashboard.zip")
	archivePath := filepath.Join(s.zashboardDir(), archiveName)
	if err := s.downloadDirectFile(ctx, trimmedURL, archivePath); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.installZashboardArchive(archivePath, archiveName, detectVersionFromName(archiveName)); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) InstallZashboardFromUpload(sourcePath, sourceName string) (model.SystemStatus, error) {
	s.appendLogf("开始通过上传文件安装 Zashboard：%s", sourceName)
	if err := s.installZashboardArchive(sourcePath, sourceName, detectVersionFromName(sourceName)); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) installZashboardArchive(sourcePath, sourceName, fallbackVersion string) error {
	s.appendLogf("开始安装 Zashboard 资源包：%s", sourceName)
	archiveName := safeBaseName(sourceName, "zashboard.zip")
	archivePath := filepath.Join(s.zashboardDir(), archiveName)

	if filepath.Clean(sourcePath) != filepath.Clean(archivePath) {
		if err := copyFile(sourcePath, archivePath); err != nil {
			return err
		}
	}

	if err := clearDirExcept(s.zashboardDir(), archiveName); err != nil {
		return err
	}

	if err := extractZipToDir(archivePath, s.zashboardDir()); err != nil {
		return err
	}

	if !fileExists(filepath.Join(s.zashboardRoot(), "index.html")) {
		return errors.New("zashboard 资源校验失败：缺少 index.html")
	}

	version := strings.TrimSpace(fallbackVersion)
	if detectedVersion := s.detectInstalledZashboardVersion(); detectedVersion != "" {
		version = detectedVersion
	}
	if version == "" {
		version = "unknown"
	}

	s.mu.Lock()
	s.status.ZashboardVersion = version
	s.status.ZashboardReady = true
	s.status.ZashboardError = ""
	s.syncInstallStateLocked()
	s.mu.Unlock()

	s.appendLogf("Zashboard 安装完成，当前版本：%s", version)
	return os.WriteFile(filepath.Join(s.zashboardDir(), "version.txt"), []byte(version+"\n"), 0o644)
}

func (s *Service) fetchLatestRelease(ctx context.Context, owner, repo string) (releaseInfo, error) {
	rawURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	var lastErr error

	for _, candidateURL := range s.preferredUpdateURLs(rawURL, true) {
		s.appendLogf("尝试获取发布信息：%s", candidateURL)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, candidateURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		request.Header.Set("User-Agent", "Graydeck")
		response, err := s.httpClient.Do(request)
		if err != nil {
			lastErr = err
			continue
		}

		if response.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("获取发布信息失败：%s", response.Status)
			response.Body.Close()
			continue
		}

		var release releaseInfo
		decodeErr := json.NewDecoder(response.Body).Decode(&release)
		response.Body.Close()
		if decodeErr != nil {
			lastErr = decodeErr
			continue
		}

		return release, nil
	}

	if lastErr == nil {
		lastErr = errors.New("获取发布信息失败")
	}

	return releaseInfo{}, lastErr
}

func (s *Service) selectCoreAsset(release releaseInfo) (releaseAsset, error) {
	prefix := fmt.Sprintf("mihomo-%s-%s-", s.cfg.CoreTargetOS, s.cfg.CoreTargetArch)

	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, prefix) && isPreferredCoreAsset(asset.Name) && !strings.Contains(asset.Name, "-go") && !strings.Contains(asset.Name, "compatible") {
			return asset, nil
		}
	}

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, s.cfg.CoreTargetOS) && strings.Contains(asset.Name, s.cfg.CoreTargetArch) && isPreferredCoreAsset(asset.Name) {
			return asset, nil
		}
	}

	return releaseAsset{}, errors.New("未找到匹配当前平台的核心文件")
}

func isPreferredCoreAsset(name string) bool {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	if lowerName == "" || isUnsupportedCorePackage(lowerName) {
		return false
	}

	return strings.HasSuffix(lowerName, ".gz") || strings.HasSuffix(lowerName, ".zip") || strings.HasSuffix(lowerName, ".exe")
}

func isUnsupportedCorePackage(name string) bool {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	return strings.HasSuffix(lowerName, ".deb") || strings.HasSuffix(lowerName, ".rpm") || strings.HasSuffix(lowerName, ".apk")
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

func (s *Service) downloadReleaseAsset(ctx context.Context, asset releaseAsset, outputPath string) error {
	return s.downloadURL(ctx, asset.BrowserDownloadURL, outputPath, strings.TrimPrefix(asset.Digest, "sha256:"), true)
}

func (s *Service) downloadDirectFile(ctx context.Context, rawURL, outputPath string) error {
	return s.downloadURL(ctx, rawURL, outputPath, "", false)
}

func (s *Service) downloadURL(ctx context.Context, rawURL, outputPath, expectedDigest string, preferProxy bool) error {
	var lastErr error

	for _, candidateURL := range s.preferredUpdateURLs(rawURL, preferProxy) {
		if err := s.downloadSingleURL(ctx, candidateURL, outputPath, expectedDigest); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	if lastErr == nil {
		lastErr = errors.New("下载失败")
	}

	return lastErr
}

func (s *Service) downloadSingleURL(ctx context.Context, rawURL, outputPath, expectedDigest string) error {
	s.appendLogf("开始下载资源：%s", rawURL)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
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

	if expectedDigest != "" {
		actualDigest := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(expectedDigest, actualDigest) {
			return fmt.Errorf("校验失败：期望 %s，实际 %s", expectedDigest, actualDigest)
		}
	}

	s.appendLogf("资源下载完成：%s", outputPath)
	return nil
}

func (s *Service) preferredUpdateURLs(rawURL string, preferProxy bool) []string {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return nil
	}

	urls := make([]string, 0, 2)
	if preferProxy && s.PreferProxyForUpdate() {
		proxyURL := strings.TrimRight(s.UpdateProxyURL(), "/")
		if proxyURL != "" {
			urls = append(urls, proxyURL+"/"+trimmedURL)
		}
	}

	urls = append(urls, trimmedURL)
	return urls
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
