package manager

import (
	"fmt"
	"runtime"
	"testing"
)

func TestSelectGraydeckAssetMatchesCurrentPlatform(t *testing.T) {
	expectedName := fmt.Sprintf("graydeck-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedName += ".exe"
	}

	service := &Service{}
	asset, err := service.selectGraydeckAsset(releaseInfo{
		Assets: []releaseAsset{
			{Name: "graydeck-linux-arm64"},
			{Name: expectedName},
		},
	})
	if err != nil {
		t.Fatalf("匹配 Graydeck 资源失败：%v", err)
	}

	if asset.Name != expectedName {
		t.Fatalf("匹配到了错误的 Graydeck 资源：%s", asset.Name)
	}
}

func TestParseSystemdServiceInfo(t *testing.T) {
	info := parseSystemdServiceInfo("LoadState=loaded\nActiveState=active\nMainPID=87\n")

	if info.LoadState != "loaded" || info.ActiveState != "active" || info.MainPID != 87 {
		t.Fatalf("解析 systemd 状态失败：%+v", info)
	}
}
