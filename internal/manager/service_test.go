package manager

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
