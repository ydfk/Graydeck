package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mihomo-manager/internal/model"
)

func TestCreateSubscriptionDoesNotSyncRemote(t *testing.T) {
	service := &Service{
		cfg: Config{
			DataDir: t.TempDir(),
		},
	}

	subscription, err := service.CreateSubscription(context.Background(), "测试配置", "http://127.0.0.1:1/not-found", "30m")
	if err != nil {
		t.Fatalf("新增配置不应触发远程同步：%v", err)
	}

	if !subscription.Enabled {
		t.Fatalf("首个配置应保持自动选中")
	}

	if subscription.Status != "pending" {
		t.Fatalf("新增配置应等待手动更新，当前状态：%s", subscription.Status)
	}
}

func TestUpdateSubscriptionDoesNotSyncRemote(t *testing.T) {
	service := &Service{
		cfg: Config{
			DataDir: t.TempDir(),
		},
		subscriptions: []model.Subscription{
			{
				ID:               "current",
				Name:             "旧配置",
				URL:              "https://example.com/old.yaml",
				SyncInterval:     "30m",
				Status:           "ready",
				PreviewAvailable: true,
			},
		},
	}

	subscription, err := service.UpdateSubscription(context.Background(), "current", "新配置", "http://127.0.0.1:1/not-found", "30m")
	if err != nil {
		t.Fatalf("保存配置不应触发远程同步：%v", err)
	}

	if subscription.Status != "pending" {
		t.Fatalf("修改订阅地址后应等待手动更新，当前状态：%s", subscription.Status)
	}

	if subscription.PreviewAvailable {
		t.Fatalf("修改订阅地址后不应继续标记旧预览可用")
	}
}

func TestActivateSubscriptionDoesNotSyncRemote(t *testing.T) {
	service := &Service{
		cfg: Config{
			DataDir:        t.TempDir(),
			BaseConfigPath: filepath.Join(t.TempDir(), "base.yaml"),
		},
		subscriptions: []model.Subscription{
			{
				ID:           "target",
				Name:         "待同步配置",
				URL:          "http://127.0.0.1:1/not-found",
				SyncInterval: "30m",
				Status:       "pending",
			},
		},
	}

	_, err := service.ActivateSubscription(context.Background(), "target")
	if err == nil {
		t.Fatalf("待同步配置不应直接切换")
	}

	if !strings.Contains(err.Error(), "手动更新") {
		t.Fatalf("错误信息应提示手动更新，当前：%v", err)
	}
}

func TestEnsureRuntimeUsesCurrentConfigWhenEnabledSubscriptionUpdateFailed(t *testing.T) {
	dataDir := t.TempDir()
	service := &Service{
		cfg: Config{
			DataDir: dataDir,
		},
		subscriptions: []model.Subscription{
			{
				ID:                "current",
				Name:              "当前配置",
				Enabled:           true,
				Status:            "fetch_failed",
				LastFailureReason: "订阅拉取失败",
			},
		},
	}

	if err := os.MkdirAll(filepath.Dir(service.currentConfigPath()), 0o755); err != nil {
		t.Fatalf("创建运行配置目录失败：%v", err)
	}

	if err := os.WriteFile(service.currentConfigPath(), []byte("mixed-port: 7890\n"), 0o644); err != nil {
		t.Fatalf("写入运行配置失败：%v", err)
	}

	if err := service.ensureRuntime(context.Background()); err != nil {
		t.Fatalf("使用已有运行配置启动失败：%v", err)
	}

	status := service.Status()
	if status.CurrentConfigName != "当前配置" {
		t.Fatalf("保底启动未保留当前配置名：%q", status.CurrentConfigName)
	}

	if status.RuntimeError != "未找到可执行核心文件" {
		t.Fatalf("未进入已有运行配置分支：%q", status.RuntimeError)
	}
}

func TestValidateSyncIntervalSupportsFreeFormDurations(t *testing.T) {
	tests := map[string]time.Duration{
		"15m":   15 * time.Minute,
		"2h30m": 2*time.Hour + 30*time.Minute,
		"3d":    72 * time.Hour,
		"1.5w":  252 * time.Hour,
	}

	for input, expected := range tests {
		normalized, err := validateSyncInterval(input)
		if err != nil {
			t.Fatalf("自由同步间隔 %s 不应校验失败：%v", input, err)
		}

		actual, err := parseInterval(normalized)
		if err != nil || actual != expected {
			t.Fatalf("同步间隔 %s 解析错误：得到 %v，期望 %v，错误 %v", input, actual, expected, err)
		}
	}
}

func TestValidateSyncIntervalRejectsTooShortInterval(t *testing.T) {
	if _, err := validateSyncInterval("30s"); err == nil {
		t.Fatalf("小于一分钟的自动同步间隔应被拒绝")
	}
}

func TestUpdateSubscriptionStoresAutomaticSyncState(t *testing.T) {
	service := &Service{
		cfg: Config{DataDir: t.TempDir()},
		subscriptions: []model.Subscription{{
			ID:           "current",
			Name:         "配置",
			URL:          "https://example.com/config.yaml",
			SyncInterval: "30m",
			AutoSync:     true,
		}},
	}

	disabled, err := service.UpdateSubscription(context.Background(), "current", "配置", "https://example.com/config.yaml", "disabled")
	if err != nil {
		t.Fatalf("关闭自动同步失败：%v", err)
	}
	if disabled.AutoSync || disabled.SyncInterval != "disabled" {
		t.Fatalf("关闭自动同步后状态不正确：%+v", disabled)
	}

	enabled, err := service.UpdateSubscription(context.Background(), "current", "配置", "https://example.com/config.yaml", "3d")
	if err != nil {
		t.Fatalf("启用自由间隔失败：%v", err)
	}
	if !enabled.AutoSync || enabled.SyncInterval != "3d" {
		t.Fatalf("启用自动同步后状态不正确：%+v", enabled)
	}
}

func TestTickAutoSyncFetchesDueSubscriptionAndRecordsAttempt(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		_, _ = w.Write([]byte("proxies: []\n"))
	}))
	defer server.Close()

	dataDir := t.TempDir()
	baseConfigPath := filepath.Join(dataDir, "base.yaml")
	service := &Service{
		cfg: Config{
			DataDir:        dataDir,
			BaseConfigPath: baseConfigPath,
		},
		httpClient: server.Client(),
		subscriptions: []model.Subscription{{
			ID:           "auto",
			Name:         "自动配置",
			URL:          server.URL,
			AutoSync:     true,
			SyncInterval: "1m",
			LastSyncAt:   formatLocalTime(localNow().Add(-2 * time.Minute)),
		}},
	}

	if err := os.MkdirAll(service.subscriptionDir(), 0o755); err != nil {
		t.Fatalf("创建订阅目录失败：%v", err)
	}
	if err := os.MkdirAll(service.runtimeDir(), 0o755); err != nil {
		t.Fatalf("创建运行目录失败：%v", err)
	}
	if err := os.WriteFile(baseConfigPath, []byte("mixed-port: 7890\n"), 0o644); err != nil {
		t.Fatalf("写入基础配置失败：%v", err)
	}

	service.tickAutoSync()

	updated, err := service.findSubscription("auto")
	if err != nil {
		t.Fatalf("读取自动同步结果失败：%v", err)
	}
	if requestCount != 1 {
		t.Fatalf("到期配置应发起一次自动拉取，实际 %d 次", requestCount)
	}
	if updated.LastSyncAt == "" || updated.LastSyncTrigger != "auto" {
		t.Fatalf("自动同步尝试未被记录：%+v", updated)
	}
}

func TestStopCoreWaitsForProcessExit(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=TestCoreProcessHelper")
	command.Env = append(os.Environ(), "GRAYDECK_CORE_HELPER=1")
	if err := command.Start(); err != nil {
		t.Fatalf("启动测试核心进程失败：%v", err)
	}

	done := make(chan struct{})
	service := &Service{coreCmd: command, coreDone: done}
	go service.waitCore(command, done)
	service.stopCore()

	if command.ProcessState == nil || !command.ProcessState.Exited() {
		t.Fatalf("停止核心返回前进程应已退出")
	}
}

func TestCoreProcessHelper(t *testing.T) {
	if os.Getenv("GRAYDECK_CORE_HELPER") != "1" {
		return
	}

	time.Sleep(30 * time.Second)
	os.Exit(0)
}
