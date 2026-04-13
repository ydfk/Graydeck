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
	DataDir          string
	CoreTargetOS     string
	CoreTargetArch   string
	ControllerAddr   string
	RuntimeMixedPort string
	RuntimeSecret    string
	BaseConfigPath   string
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

	if err := service.loadSubscriptions(); err != nil {
		return nil, err
	}

	service.loadInstalledVersions()
	service.loadRuntimeConfigStatus()

	service.setRuntimeStatus("stopped", "未找到可用配置文件")

	if err := service.bootstrap(context.Background()); err != nil {
		service.setRuntimeStatus("error", err.Error())
	}

	return service, nil
}

func (s *Service) bootstrap(ctx context.Context) error {
	if err := s.refreshCoreMetadata(ctx); err != nil {
		s.setRuntimeStatus("error", fmt.Sprintf("获取核心版本失败：%v", err))
	}

	if err := s.ensureCoreInstalled(ctx); err != nil {
		s.setRuntimeStatus("error", fmt.Sprintf("下载核心失败：%v", err))
	}

	if err := s.refreshZashboardMetadata(ctx); err != nil {
		s.setZashboardStatusError(err.Error())
	}

	if err := s.ensureZashboardInstalled(ctx); err != nil {
		s.setZashboardStatusError(err.Error())
	}

	if err := s.RefreshAll(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) RefreshAll(ctx context.Context) error {
	if err := s.refreshCoreMetadata(ctx); err != nil {
		s.setRuntimeStatus("error", fmt.Sprintf("获取核心版本失败：%v", err))
	}

	if err := s.refreshZashboardMetadata(ctx); err != nil {
		s.setZashboardStatusError(err.Error())
	}

	_ = s.syncSubscriptions(ctx)

	return s.ensureRuntime(ctx)
}

func (s *Service) Status() model.SystemStatus {
	s.loadRuntimeConfigStatus()

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

	s.mu.Lock()
	shouldEnable := len(s.subscriptions) == 0
	if shouldEnable {
		subscription.Enabled = true
	}
	s.subscriptions = append(s.subscriptions, subscription)
	s.mu.Unlock()

	if err := s.saveSubscriptions(); err != nil {
		return model.Subscription{}, err
	}

	if _, err := s.syncSubscription(ctx, subscription.ID); err != nil {
		return model.Subscription{}, err
	}

	if shouldEnable {
		if _, err := s.ActivateSubscription(ctx, subscription.ID); err != nil {
			return model.Subscription{}, err
		}
	}

	return s.findSubscription(subscription.ID)
}

func (s *Service) UpdateSubscription(ctx context.Context, id, name, rawURL, syncInterval string) (model.Subscription, error) {
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
		return model.Subscription{}, err
	}
	s.mu.Unlock()

	if _, err := s.syncSubscription(ctx, id); err != nil {
		return model.Subscription{}, err
	}

	subscription, err := s.findSubscription(id)
	if err != nil {
		return model.Subscription{}, err
	}

	if subscription.Enabled {
		if err := s.ensureRuntime(ctx); err != nil {
			return model.Subscription{}, err
		}
	}

	return subscription, nil
}

func (s *Service) ActivateSubscription(ctx context.Context, id string) (model.Subscription, error) {
	subscription, err := s.syncSubscription(ctx, id)
	if err != nil {
		return model.Subscription{}, err
	}

	if subscription.Status != "ready" && subscription.Status != "active" {
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
		return model.Subscription{}, err
	}

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

func (s *Service) UpdateCore(ctx context.Context) (model.SystemStatus, error) {
	if err := s.refreshCoreMetadata(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.downloadCore(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.ensureRuntime(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) UpdateZashboard(ctx context.Context) (model.SystemStatus, error) {
	if err := s.refreshZashboardMetadata(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	if err := s.downloadZashboard(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) StartRuntime(ctx context.Context) (model.SystemStatus, error) {
	if err := s.ensureRuntime(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) RestartRuntime(ctx context.Context) (model.SystemStatus, error) {
	s.stopCore()

	if err := s.ensureRuntime(ctx); err != nil {
		return model.SystemStatus{}, err
	}

	return s.Status(), nil
}

func (s *Service) StopRuntime() model.SystemStatus {
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

	if subscription.URL == "" {
		return model.Subscription{}, errors.New("配置文件地址不能为空")
	}

	content, err := s.fetchText(ctx, subscription.URL)
	if err != nil {
		s.updateSubscriptionStatus(id, "fetch_failed", fmt.Sprintf("订阅拉取失败：%v", err), false)
		return model.Subscription{}, err
	}

	if err := os.WriteFile(s.subscriptionPreviewPath(id), []byte(content), 0o644); err != nil {
		return model.Subscription{}, err
	}

	validatePath, cleanup, err := s.buildValidationConfig(id)
	if err != nil {
		s.updateSubscriptionStatus(id, "validation_failed", fmt.Sprintf("生成运行配置失败：%v", err), true)
		return model.Subscription{}, err
	}
	defer cleanup()

	if err := s.validateConfigFile(validatePath); err != nil {
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

	return s.findSubscription(id)
}

func (s *Service) ensureRuntime(ctx context.Context) error {
	enabled, ok := s.enabledSubscription()
	if !ok {
		s.stopCore()
		s.setRuntimeStatus("stopped", "未选择当前配置文件")
		return nil
	}

	if enabled.Status != "active" && enabled.Status != "ready" {
		s.stopCore()
		s.setRuntimeStatus("error", enabled.LastFailureReason)
		return nil
	}

	if err := s.writeRuntimeConfig(s.subscriptionPreviewPath(enabled.ID), s.currentConfigPath()); err != nil {
		s.stopCore()
		s.setRuntimeStatus("error", fmt.Sprintf("应用配置失败：%v", err))
		return err
	}

	if !s.isCoreReady() {
		s.stopCore()
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
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) loadInstalledVersions() {
	if version, err := os.ReadFile(filepath.Join(s.coreDir(), "version.txt")); err == nil {
		s.status.CoreVersion = strings.TrimSpace(string(version))
	}

	if version, err := os.ReadFile(filepath.Join(s.zashboardDir(), "version.txt")); err == nil {
		s.status.ZashboardVersion = strings.TrimSpace(string(version))
	}

	s.syncInstallStateLocked()
}

func (s *Service) loadRuntimeConfigStatus() {
	values := s.loadManagedRuntimeValues()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.BaseConfigPath = s.cfg.BaseConfigPath
	s.status.RuntimeMixedPort = values.mixedPort
	s.status.RuntimeControllerAddr = values.controller
	s.status.RuntimeSecret = values.secret
}

func (s *Service) setRuntimeStatus(status, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.RuntimeStatus = status
	s.status.RuntimeError = reason
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

	if s.status.CoreVersion == "" || s.status.CoreLatestVersion == "" {
		s.status.CoreIsLatest = false
	} else {
		s.status.CoreIsLatest = s.status.CoreVersion == s.status.CoreLatestVersion
	}

	if s.status.ZashboardVersion == "" || s.status.ZashboardLatestVersion == "" {
		s.status.ZashboardIsLatest = false
	} else {
		s.status.ZashboardIsLatest = s.status.ZashboardVersion == s.status.ZashboardLatestVersion
	}

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
			controller: s.cfg.ControllerAddr,
			secret:     s.cfg.RuntimeSecret,
		}
	}

	return parseManagedRuntimeValues(string(content), managedRuntimeValues{
		mixedPort:  s.cfg.RuntimeMixedPort,
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
