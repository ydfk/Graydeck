package manager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"mihomo-manager/internal/buildinfo"
	"mihomo-manager/internal/model"
)

type systemdServiceInfo struct {
	LoadState   string
	ActiveState string
	MainPID     int
}

func (s *Service) UpdateGraydeck(ctx context.Context) (model.SystemStatus, error) {
	if supported, reason := s.graydeckSelfUpdateSupported(); !supported {
		return model.SystemStatus{}, errors.New(reason)
	}

	s.appendLog("开始更新 Graydeck")
	release, err := s.fetchLatestRelease(ctx, buildinfo.RepositoryOwner, buildinfo.RepositoryName)
	if err != nil {
		s.appendLogf("更新 Graydeck 失败：%v", err)
		return model.SystemStatus{}, err
	}

	asset, err := s.selectGraydeckAsset(release)
	if err != nil {
		s.appendLogf("更新 Graydeck 失败：%v", err)
		return model.SystemStatus{}, err
	}

	currentPath, err := os.Executable()
	if err != nil {
		return model.SystemStatus{}, err
	}

	packagePath := filepath.Join(filepath.Dir(currentPath), ".graydeck-download-"+safeBaseName(asset.Name, "graydeck"))
	if err := s.downloadReleaseAsset(ctx, asset, packagePath); err != nil {
		_ = os.Remove(packagePath)
		s.appendLogf("更新 Graydeck 失败：%v", err)
		return model.SystemStatus{}, err
	}
	defer os.Remove(packagePath)

	if err := installGraydeckBinary(packagePath, currentPath); err != nil {
		s.appendLogf("更新 Graydeck 失败：%v", err)
		return model.SystemStatus{}, err
	}

	s.mu.Lock()
	s.status.GraydeckVersion = release.TagName
	s.status.GraydeckLatestVersion = release.TagName
	s.syncInstallStateLocked()
	s.mu.Unlock()

	if err := scheduleGraydeckServiceRestart(graydeckServiceName()); err != nil {
		s.appendLogf("Graydeck 已更新，但自动重启失败：%v", err)
		return model.SystemStatus{}, err
	}

	s.appendLog("Graydeck 更新完成，服务即将重启")
	return s.Status(), nil
}

func (s *Service) graydeckSelfUpdateSupported() (bool, string) {
	if strings.EqualFold(strings.TrimSpace(s.cfg.DeploymentMode), "docker") {
		return false, "Docker 模式请通过更新镜像升级 Graydeck"
	}

	if runtime.GOOS != "linux" {
		return false, "Graydeck 自更新目前仅支持 Linux systemd 服务"
	}

	if _, err := exec.LookPath("systemctl"); err != nil {
		return false, "未找到 systemctl，无法自动重启 Graydeck 服务"
	}

	unitName := graydeckServiceName()
	info, err := queryGraydeckServiceInfo(unitName)
	if err != nil {
		return false, err.Error()
	}

	if info.LoadState != "loaded" {
		return false, fmt.Sprintf("%s 未安装", unitName)
	}

	if info.ActiveState != "active" || info.MainPID != os.Getpid() {
		return false, fmt.Sprintf("当前进程不是 %s 管理的主进程", unitName)
	}

	return true, ""
}

func (s *Service) selectGraydeckAsset(release releaseInfo) (releaseAsset, error) {
	expectedName := fmt.Sprintf("graydeck-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedName += ".exe"
	}

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset, nil
		}
	}

	return releaseAsset{}, fmt.Errorf("未找到匹配当前平台的 Graydeck 文件：%s", expectedName)
}

func installGraydeckBinary(sourcePath, currentPath string) error {
	tempFile, err := os.CreateTemp(filepath.Dir(currentPath), ".graydeck-install-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	defer os.Remove(tempPath)

	if err := copyFile(sourcePath, tempPath); err != nil {
		return err
	}

	if err := os.Chmod(tempPath, 0o755); err != nil {
		return err
	}

	return replaceExecutable(tempPath, currentPath)
}

func queryGraydeckServiceInfo(unitName string) (systemdServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output, err := exec.CommandContext(
		ctx,
		"systemctl",
		"show",
		unitName,
		"--property=LoadState",
		"--property=ActiveState",
		"--property=MainPID",
		"--no-pager",
	).Output()
	if err != nil {
		return systemdServiceInfo{}, fmt.Errorf("读取 %s 状态失败：%w", unitName, err)
	}

	return parseSystemdServiceInfo(string(output)), nil
}

func parseSystemdServiceInfo(output string) systemdServiceInfo {
	info := systemdServiceInfo{}

	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}

		switch key {
		case "LoadState":
			info.LoadState = value
		case "ActiveState":
			info.ActiveState = value
		case "MainPID":
			pid, _ := strconv.Atoi(value)
			info.MainPID = pid
		}
	}

	return info
}

func scheduleGraydeckServiceRestart(unitName string) error {
	systemctlPath, err := exec.LookPath("systemctl")
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("systemd-run"); err == nil {
		return exec.Command(
			"systemd-run",
			fmt.Sprintf("--unit=graydeck-self-restart-%d", os.Getpid()),
			"--description=Restart Graydeck after self-update",
			"--on-active=1s",
			systemctlPath,
			"restart",
			unitName,
		).Start()
	}

	go func() {
		time.Sleep(time.Second)
		_ = exec.Command(systemctlPath, "restart", unitName).Run()
	}()

	return nil
}

func graydeckServiceName() string {
	name := strings.TrimSpace(os.Getenv("GRAYDECK_SERVICE_NAME"))
	if name == "" {
		return "graydeck.service"
	}

	if !strings.HasSuffix(name, ".service") {
		return name + ".service"
	}

	return name
}
