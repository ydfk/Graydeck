package manager

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mihomo-manager/internal/model"
)

type Config struct {
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

type Service struct {
	mu            sync.RWMutex
	cfg           Config
	httpClient    *http.Client
	status        model.SystemStatus
	subscriptions []model.Subscription
	coreCmd       *exec.Cmd
	logs          []model.LogEntry
}

func New(cfg Config) (*Service, error) {
	service := &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}

	if err := service.ensureLayout(); err != nil {
		return nil, err
	}

	if err := service.ensureBaseConfig(); err != nil {
		return nil, err
	}

	if err := service.ensureAppConfig(); err != nil {
		return nil, err
	}

	if err := service.loadSubscriptions(); err != nil {
		return nil, err
	}

	service.loadInstalledVersions()
	service.loadRuntimeConfigStatus()
	service.loadAppConfigStatus()
	service.appendLog("Graydeck 服务初始化完成")

	service.setRuntimeStatus("stopped", "未找到可用配置文件")

	if err := service.bootstrap(context.Background()); err != nil {
		service.setRuntimeStatus("error", err.Error())
	}

	return service, nil
}

func (s *Service) bootstrap(ctx context.Context) error {
	s.appendLog("开始执行启动初始化")

	if err := s.refreshCoreMetadata(ctx); err != nil {
		s.appendLogf("获取核心版本失败：%v", err)
		s.setRuntimeStatus("error", fmt.Sprintf("获取核心版本失败：%v", err))
	}

	if err := s.ensureCoreInstalled(ctx); err != nil {
		s.appendLogf("检查核心安装状态失败：%v", err)
		s.setRuntimeStatus("error", fmt.Sprintf("下载核心失败：%v", err))
	}

	if err := s.refreshZashboardMetadata(ctx); err != nil {
		s.appendLogf("获取 Zashboard 版本失败：%v", err)
		s.setZashboardStatusError(err.Error())
	}

	if err := s.ensureZashboardInstalled(ctx); err != nil {
		s.appendLogf("检查 Zashboard 安装状态失败：%v", err)
		s.setZashboardStatusError(err.Error())
	}

	if err := s.RefreshAll(ctx); err != nil {
		s.appendLogf("启动初始化失败：%v", err)
		return err
	}

	s.appendLog("启动初始化完成")
	return nil
}

func (s *Service) RefreshAll(ctx context.Context) error {
	s.appendLog("开始刷新系统状态")

	if err := s.refreshCoreMetadata(ctx); err != nil {
		s.appendLogf("刷新核心版本失败：%v", err)
		s.setRuntimeStatus("error", fmt.Sprintf("获取核心版本失败：%v", err))
	}

	if err := s.refreshZashboardMetadata(ctx); err != nil {
		s.appendLogf("刷新 Zashboard 版本失败：%v", err)
		s.setZashboardStatusError(err.Error())
	}

	if err := s.syncSubscriptions(ctx); err != nil {
		s.appendLogf("刷新配置文件状态时出现异常：%v", err)
	}

	if err := s.ensureRuntime(ctx); err != nil {
		s.appendLogf("刷新运行状态失败：%v", err)
		return err
	}

	s.appendLog("系统状态刷新完成")
	return nil
}

func (s *Service) Status() model.SystemStatus {
	s.loadRuntimeConfigStatus()
	s.loadAppConfigStatus()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.syncInstallStateLocked()
	return s.status
}

func (s *Service) Logs() []model.LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]model.LogEntry, len(s.logs))
	for index := range s.logs {
		items[index] = s.logs[len(s.logs)-1-index]
	}
	return items
}

func (s *Service) Subscriptions() []model.Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]model.Subscription, len(s.subscriptions))
	copy(items, s.subscriptions)
	return items
}

func (s *Service) CreateSubscription(ctx context.Context, name, rawURL, syncInterval string) (model.Subscription, error) {
	subscription := model.Subscription{
		ID:           fmt.Sprintf("cfg-%d", time.Now().UnixNano()),
		Name:         strings.TrimSpace(name),
		URL:          strings.TrimSpace(rawURL),
		SyncInterval: strings.TrimSpace(syncInterval),
		AutoSync:     true,
		Status:       "pending",
	}

	if subscription.Name == "" || subscription.URL == "" || subscription.SyncInterval == "" {
		return model.Subscription{}, errors.New("名称、地址和同步频率不能为空")
	}

	s.appendLogf("开始新增配置文件：%s", subscription.Name)

	s.mu.Lock()
	shouldEnable := len(s.subscriptions) == 0
	if shouldEnable {
		subscription.Enabled = true
	}
	s.subscriptions = append(s.subscriptions, subscription)
	s.mu.Unlock()

	if err := s.saveSubscriptions(); err != nil {
		s.appendLogf("保存配置文件失败：%s，%v", subscription.Name, err)
		return model.Subscription{}, err
	}

	if _, err := s.syncSubscription(ctx, subscription.ID); err != nil {
		s.appendLogf("新增配置文件后首次更新失败：%s，%v", subscription.Name, err)
		return model.Subscription{}, err
	}

	if shouldEnable {
		if _, err := s.ActivateSubscription(ctx, subscription.ID); err != nil {
			s.appendLogf("新增配置文件后自动启用失败：%s，%v", subscription.Name, err)
			return model.Subscription{}, err
		}
	}

	s.appendLogf("新增配置文件完成：%s", subscription.Name)
	return s.findSubscription(subscription.ID)
}

func (s *Service) UpdateSubscription(ctx context.Context, id, name, rawURL, syncInterval string) (model.Subscription, error) {
	s.appendLogf("开始保存配置文件：%s", strings.TrimSpace(name))
	s.mu.Lock()
	found := false

	for index := range s.subscriptions {
		item := &s.subscriptions[index]
		if item.ID != id {
			continue
		}

		item.Name = strings.TrimSpace(name)
		item.URL = strings.TrimSpace(rawURL)
		item.SyncInterval = strings.TrimSpace(syncInterval)
		item.Status = "pending"
		item.LastFailureReason = ""

		if item.Name == "" || item.URL == "" || item.SyncInterval == "" {
			s.mu.Unlock()
			return model.Subscription{}, errors.New("名称、地址和同步频率不能为空")
		}

		found = true
		break
	}

	if !found {
		s.mu.Unlock()
		return model.Subscription{}, errors.New("配置文件不存在")
	}

	if err := s.saveSubscriptionsLocked(); err != nil {
		s.mu.Unlock()
		s.appendLogf("保存配置文件失败：%s，%v", strings.TrimSpace(name), err)
		return model.Subscription{}, err
	}
	s.mu.Unlock()

	if _, err := s.syncSubscription(ctx, id); err != nil {
		s.appendLogf("保存配置文件后更新失败：%s，%v", strings.TrimSpace(name), err)
		return model.Subscription{}, err
	}

	subscription, err := s.findSubscription(id)
	if err != nil {
		return model.Subscription{}, err
	}

	if subscription.Enabled {
		if err := s.ensureRuntime(ctx); err != nil {
			s.appendLogf("保存配置文件后应用失败：%s，%v", subscription.Name, err)
			return model.Subscription{}, err
		}
	}

	s.appendLogf("保存配置文件完成：%s", subscription.Name)
	return subscription, nil
}

func (s *Service) ActivateSubscription(ctx context.Context, id string) (model.Subscription, error) {
	target, _ := s.findSubscription(id)
	s.appendLogf("开始切换当前配置：%s", target.Name)

	subscription, err := s.syncSubscription(ctx, id)
	if err != nil {
		s.appendLogf("切换配置前更新失败：%s，%v", target.Name, err)
		return model.Subscription{}, err
	}

	if subscription.Status != "ready" && subscription.Status != "active" {
		s.appendLogf("切换配置失败：%s 当前状态不可用", subscription.Name)
		return model.Subscription{}, errors.New("当前配置文件不可用，无法切换")
	}

	s.mu.Lock()
	for index := range s.subscriptions {
		s.subscriptions[index].Enabled = s.subscriptions[index].ID == id
		if s.subscriptions[index].Enabled {
			s.subscriptions[index].Status = "active"
		} else if s.subscriptions[index].Status == "active" {
			s.subscriptions[index].Status = "ready"
		}
	}
	if err := s.saveSubscriptionsLocked(); err != nil {
		s.mu.Unlock()
		return model.Subscription{}, err
	}
	s.mu.Unlock()

	if err := s.ensureRuntime(ctx); err != nil {
		s.appendLogf("切换配置后应用失败：%s，%v", subscription.Name, err)
		return model.Subscription{}, err
	}

	s.appendLogf("当前配置已切换：%s", subscription.Name)
	return s.findSubscription(id)
}

func (s *Service) PreviewSubscription(id string) (model.SubscriptionPreview, error) {
	subscription, err := s.findSubscription(id)
	if err != nil {
		return model.SubscriptionPreview{}, err
	}

	content, err := os.ReadFile(s.subscriptionPreviewPath(id))
	if err != nil {
		return model.SubscriptionPreview{}, err
	}

	return model.SubscriptionPreview{
		ID:      subscription.ID,
		Name:    subscription.Name,
		Content: string(content),
	}, nil
}

func (s *Service) RuntimeConfigPreview() (model.RuntimeConfigPreview, error) {
	if enabled, ok := s.enabledSubscription(); !ok || strings.TrimSpace(enabled.Name) == "" {
		return model.RuntimeConfigPreview{}, errors.New("当前没有可查看的运行配置")
	}

	content, err := os.ReadFile(s.currentConfigPath())
	if err != nil {
		return model.RuntimeConfigPreview{}, err
	}

	s.mu.RLock()
	name := strings.TrimSpace(s.status.CurrentConfigName)
	s.mu.RUnlock()

	if name == "" {
		name = filepath.Base(s.currentConfigPath())
	}

	return model.RuntimeConfigPreview{
		Name:    name,
		Path:    s.currentConfigPath(),
		Content: string(content),
	}, nil
}

func (s *Service) DefaultConfig() model.DefaultRuntimeConfig {
	return s.DefaultRuntimeConfig()
}

func (s *Service) UpdateDefaultConfig(ctx context.Context, config model.DefaultRuntimeConfig) (model.SystemStatus, error) {
	if err := s.UpdateDefaultRuntimeConfig(config); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.ensureRuntime(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) UpdateCore(ctx context.Context) (model.SystemStatus, error) {
	s.appendLog("开始更新 mihomo 核心")
	if err := s.installCoreFromAuto(ctx); err != nil {
		s.appendLogf("更新 mihomo 核心失败：%v", err)
		return model.SystemStatus{}, err
	}

	if err := s.ensureRuntime(ctx); err != nil {
		s.appendLogf("更新 mihomo 核心后重载失败：%v", err)
		return model.SystemStatus{}, err
	}

	s.appendLog("mihomo 核心更新完成")
	return s.Status(), nil
}

func (s *Service) UpdateZashboard(ctx context.Context) (model.SystemStatus, error) {
	s.appendLog("开始更新 Zashboard")
	if err := s.installZashboardFromAuto(ctx); err != nil {
		s.appendLogf("更新 Zashboard 失败：%v", err)
		return model.SystemStatus{}, err
	}

	s.appendLog("Zashboard 更新完成")
	return s.Status(), nil
}

func (s *Service) StartRuntime(ctx context.Context) (model.SystemStatus, error) {
	s.appendLog("收到启动核心请求")
	if err := s.ensureRuntime(ctx); err != nil {
		s.appendLogf("启动核心失败：%v", err)
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) RestartRuntime(ctx context.Context) (model.SystemStatus, error) {
	s.appendLog("收到重启核心请求")
	s.stopCore()

	if err := s.ensureRuntime(ctx); err != nil {
		s.appendLogf("重启核心失败：%v", err)
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) StopRuntime() model.SystemStatus {
	s.appendLog("收到停止核心请求")
	s.stopCore()
	s.setRuntimeStatus("stopped", "")
	s.appendLog("核心已停止")
	return s.Status()
}

func (s *Service) ZashboardRoot() string {
	return s.zashboardRoot()
}

func (s *Service) BaseConfigPath() string {
	return s.cfg.BaseConfigPath
}

func (s *Service) ControllerURL() string {
	controllerAddr := s.runtimeControllerAddr()
	if strings.HasPrefix(controllerAddr, "http://") || strings.HasPrefix(controllerAddr, "https://") {
		return controllerAddr
	}

	return "http://" + controllerAddr
}

func (s *Service) syncSubscriptions(ctx context.Context) error {
	items := s.Subscriptions()
	var syncErr error

	for _, item := range items {
		if _, err := s.syncSubscription(ctx, item.ID); err != nil {
			syncErr = err
		}
	}

	return syncErr
}

func (s *Service) syncSubscription(ctx context.Context, id string) (model.Subscription, error) {
	subscription, err := s.findSubscription(id)
	if err != nil {
		return model.Subscription{}, err
	}

	s.appendLogf("开始更新配置文件：%s", subscription.Name)

	if subscription.URL == "" {
		s.appendLogf("更新配置文件失败：%s，订阅地址为空", subscription.Name)
		return model.Subscription{}, errors.New("配置文件地址不能为空")
	}

	content, err := s.fetchText(ctx, subscription.URL)
	if err != nil {
		s.appendLogf("拉取配置文件失败：%s，%v", subscription.Name, err)
		s.updateSubscriptionStatus(id, "fetch_failed", fmt.Sprintf("订阅拉取失败：%v", err), false)
		return model.Subscription{}, err
	}

	if err := os.WriteFile(s.subscriptionPreviewPath(id), []byte(content), 0o644); err != nil {
		s.appendLogf("写入配置预览失败：%s，%v", subscription.Name, err)
		return model.Subscription{}, err
	}

	validatePath, cleanup, err := s.buildValidationConfig(id)
	if err != nil {
		s.appendLogf("生成校验配置失败：%s，%v", subscription.Name, err)
		s.updateSubscriptionStatus(id, "validation_failed", fmt.Sprintf("生成运行配置失败：%v", err), true)
		return model.Subscription{}, err
	}
	defer cleanup()

	if err := s.validateConfigFile(validatePath); err != nil {
		s.appendLogf("配置校验失败：%s，%v", subscription.Name, err)
		s.updateSubscriptionStatus(id, "validation_failed", err.Error(), true)
		return model.Subscription{}, err
	}

	s.updateSubscriptionStatus(id, "ready", "", true)
	updated, err := s.findSubscription(id)
	if err != nil {
		return model.Subscription{}, err
	}

	if updated.Enabled {
		s.mu.Lock()
		for index := range s.subscriptions {
			if s.subscriptions[index].ID == id {
				s.subscriptions[index].Status = "active"
				break
			}
		}
		if err := s.saveSubscriptionsLocked(); err != nil {
			s.mu.Unlock()
			return model.Subscription{}, err
		}
		s.mu.Unlock()
	}

	s.appendLogf("配置文件可用：%s", subscription.Name)
	return s.findSubscription(id)
}

func (s *Service) ensureRuntime(ctx context.Context) error {
	enabled, ok := s.enabledSubscription()
	if !ok {
		s.stopCore()
		s.setCurrentConfigName("")
		s.appendLog("当前没有可用配置文件，核心不会启动")
		s.setRuntimeStatus("stopped", "未选择当前配置文件")
		return nil
	}

	if enabled.Status != "active" && enabled.Status != "ready" {
		s.stopCore()
		s.setCurrentConfigName("")
		s.appendLogf("当前配置不可用，核心不会启动：%s，%s", enabled.Name, enabled.LastFailureReason)
		s.setRuntimeStatus("error", enabled.LastFailureReason)
		return nil
	}

	if err := s.writeRuntimeConfig(s.subscriptionPreviewPath(enabled.ID), s.currentConfigPath()); err != nil {
		s.stopCore()
		s.setCurrentConfigName("")
		s.appendLogf("生成运行配置失败：%s，%v", enabled.Name, err)
		s.setRuntimeStatus("error", fmt.Sprintf("应用配置失败：%v", err))
		return err
	}

	if !s.isCoreReady() {
		s.stopCore()
		s.setCurrentConfigName(enabled.Name)
		s.appendLog("mihomo 核心未安装，无法启动")
		s.setRuntimeStatus("error", "未找到可执行核心文件")
		return nil
	}

	s.stopCore()

	command := exec.Command(s.coreExecutablePath(), "-f", s.currentConfigPath())
	stdout, err := command.StdoutPipe()
	if err != nil {
		s.setRuntimeStatus("error", fmt.Sprintf("获取核心输出失败：%v", err))
		return err
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		s.setRuntimeStatus("error", fmt.Sprintf("获取核心错误输出失败：%v", err))
		return err
	}

	if err := command.Start(); err != nil {
		s.appendLogf("启动 mihomo 核心进程失败：%v", err)
		s.setRuntimeStatus("error", fmt.Sprintf("启动核心失败：%v", err))
		return err
	}

	s.mu.Lock()
	s.coreCmd = command
	s.status.RuntimeStatus = "running"
	s.status.RuntimeError = ""
	s.status.CurrentConfigName = enabled.Name
	s.mu.Unlock()

	s.appendLog(fmt.Sprintf("核心已启动，当前配置：%s", enabled.Name))
	go s.captureLogs(stdout)
	go s.captureLogs(stderr)
	go s.waitCore(command)
	return nil
}

func (s *Service) waitCore(command *exec.Cmd) {
	err := command.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.coreCmd != command {
		return
	}

	s.coreCmd = nil

	if err != nil {
		s.status.RuntimeStatus = "error"
		s.status.RuntimeError = fmt.Sprintf("核心进程退出：%v", err)
		s.logs = appendLogEntry(s.logs, fmt.Sprintf("核心进程异常退出：%v", err))
		return
	}

	s.status.RuntimeStatus = "stopped"
	s.status.RuntimeError = "核心进程已退出"
	s.logs = appendLogEntry(s.logs, "核心进程已退出")
}

func (s *Service) stopCore() {
	s.mu.Lock()
	command := s.coreCmd
	s.coreCmd = nil
	s.mu.Unlock()

	if command == nil || command.Process == nil {
		return
	}

	_ = command.Process.Kill()
	_, _ = command.Process.Wait()
}

func (s *Service) captureLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)

	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}
		s.appendLog(message)
	}
}

func (s *Service) validateConfigFile(path string) error {
	if !s.isCoreReady() {
		return errors.New("核心尚未就绪，无法校验配置文件")
	}

	command := exec.Command(s.coreExecutablePath(), "-t", "-f", path)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("格式校验失败：%s", message)
	}

	s.appendLogf("配置校验通过：%s", filepath.Base(path))
	return nil
}

func (s *Service) updateSubscriptionStatus(id, status, reason string, previewAvailable bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Format("2006-01-02 15:04:05")

	for index := range s.subscriptions {
		item := &s.subscriptions[index]
		if item.ID != id {
			continue
		}

		item.Status = status
		item.PreviewAvailable = previewAvailable
		item.LastSyncAt = now
		item.LastFailureReason = reason
		if reason == "" {
			item.LastSuccess = now
		}
		break
	}

	_ = s.saveSubscriptionsLocked()
}

func (s *Service) enabledSubscription() (model.Subscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.subscriptions {
		if item.Enabled {
			return item, true
		}
	}

	return model.Subscription{}, false
}

func (s *Service) findSubscription(id string) (model.Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.subscriptions {
		if item.ID == id {
			return item, nil
		}
	}

	return model.Subscription{}, errors.New("配置文件不存在")
}

func (s *Service) loadSubscriptions() error {
	path := s.subscriptionsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(path, []byte("[]\n"), 0o644)
		}
		return err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		data = []byte("[]")
	}

	var items []model.Subscription
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = items
	s.mu.Unlock()
	return nil
}

func (s *Service) saveSubscriptions() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveSubscriptionsLocked()
}

func (s *Service) saveSubscriptionsLocked() error {
	data, err := json.MarshalIndent(s.subscriptions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.subscriptionsFilePath(), append(data, '\n'), 0o644)
}

func (s *Service) ensureLayout() error {
	paths := []string{
		s.cfg.DataDir,
		s.coreDir(),
		s.runtimeDir(),
		s.zashboardDir(),
		s.subscriptionDir(),
		filepath.Dir(s.cfg.BaseConfigPath),
		filepath.Dir(s.cfg.AppConfigPath),
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) loadInstalledVersions() {
	coreVersion := s.detectInstalledCoreVersion()
	zashboardVersion := s.detectInstalledZashboardVersion()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.CoreVersion = coreVersion
	s.status.ZashboardVersion = zashboardVersion
	s.syncInstallStateLocked()
}

func (s *Service) loadRuntimeConfigStatus() {
	values := s.loadManagedRuntimeValues()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.RuntimeMixedPort = values.mixedPort
	s.status.RuntimeSocksPort = values.socksPort
	s.status.RuntimeRedirPort = values.redirPort
	s.status.RuntimeTProxyPort = values.tproxyPort
}

func (s *Service) setRuntimeStatus(status, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.RuntimeStatus = status
	s.status.RuntimeError = reason
}

func (s *Service) setCurrentConfigName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.CurrentConfigName = strings.TrimSpace(name)
}

func (s *Service) setZashboardStatusError(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.ZashboardError = reason
	s.syncInstallStateLocked()
}

func (s *Service) isCoreReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.CoreExecutableReady
}

func (s *Service) subscriptionsFilePath() string {
	return filepath.Join(s.cfg.DataDir, "subscriptions.json")
}

func (s *Service) subscriptionDir() string {
	return filepath.Join(s.cfg.DataDir, "subscriptions")
}

func (s *Service) subscriptionPreviewPath(id string) string {
	return filepath.Join(s.subscriptionDir(), id+".yaml")
}

func (s *Service) runtimeDir() string {
	return filepath.Join(s.cfg.DataDir, "runtime")
}

func (s *Service) currentConfigPath() string {
	return filepath.Join(s.runtimeDir(), "current.yaml")
}

func (s *Service) coreDir() string {
	return filepath.Join(s.cfg.DataDir, "core")
}

func (s *Service) coreExecutablePath() string {
	extension := ""
	if s.cfg.CoreTargetOS == "windows" {
		extension = ".exe"
	}
	return filepath.Join(s.coreDir(), "mihomo"+extension)
}

func (s *Service) zashboardDir() string {
	return filepath.Join(s.cfg.DataDir, "zashboard")
}

func (s *Service) zashboardRoot() string {
	distDir := filepath.Join(s.zashboardDir(), "dist")
	if dirExists(distDir) && fileExists(filepath.Join(distDir, "index.html")) {
		return distDir
	}

	return s.zashboardDir()
}

func (s *Service) appendLog(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = appendLogEntry(s.logs, message)
}

func (s *Service) syncInstallStateLocked() {
	s.status.CoreExecutableReady = fileExists(s.coreExecutablePath())
	s.status.ZashboardReady = fileExists(filepath.Join(s.zashboardRoot(), "index.html"))

	s.status.CoreIsLatest = sameVersion(s.status.CoreVersion, s.status.CoreLatestVersion)
	s.status.ZashboardIsLatest = sameVersion(s.status.ZashboardVersion, s.status.ZashboardLatestVersion)

	if s.status.ZashboardReady && s.status.ZashboardVersion == "" {
		if s.status.ZashboardLatestVersion != "" {
			s.status.ZashboardVersion = s.status.ZashboardLatestVersion
		} else {
			s.status.ZashboardVersion = "installed"
		}
	}

	if s.status.ZashboardReady {
		s.status.ZashboardError = ""
	}
}

func (s *Service) loadManagedRuntimeValues() managedRuntimeValues {
	content, err := os.ReadFile(s.cfg.BaseConfigPath)
	if err != nil {
		return managedRuntimeValues{
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
		}
	}

	return parseManagedRuntimeValues(string(content), managedRuntimeValues{
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
}

func (s *Service) runtimeControllerAddr() string {
	return s.loadManagedRuntimeValues().controller
}

func (s *Service) RuntimeSecret() string {
	return s.loadManagedRuntimeValues().secret
}

func (s *Service) WebRoot() string {
	return s.cfg.WebRoot
}

func appendLogEntry(current []model.LogEntry, message string) []model.LogEntry {
	current = append(current, model.LogEntry{
		At:      time.Now().Format("2006-01-02 15:04:05"),
		Message: message,
	})

	if len(current) > 400 {
		current = append([]model.LogEntry(nil), current[len(current)-400:]...)
	}

	return current
}
