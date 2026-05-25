package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"mihomo-manager/internal/model"
)

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
