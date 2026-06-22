package manager

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceExecutableReplacesExistingFile(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "mihomo")
	sourcePath := filepath.Join(dir, "mihomo-new")

	if err := os.WriteFile(targetPath, []byte("old"), 0o755); err != nil {
		t.Fatalf("写入旧核心失败：%v", err)
	}
	if err := os.WriteFile(sourcePath, []byte("new"), 0o755); err != nil {
		t.Fatalf("写入新核心失败：%v", err)
	}

	if err := replaceExecutable(sourcePath, targetPath); err != nil {
		t.Fatalf("替换核心失败：%v", err)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("读取替换后的核心失败：%v", err)
	}
	if string(content) != "new" {
		t.Fatalf("核心未被正确替换：%q", content)
	}
}

func TestZashboardInstallationPersistsInDataDir(t *testing.T) {
	dataDir := t.TempDir()
	service := &Service{cfg: Config{DataDir: dataDir}}
	if err := os.MkdirAll(service.zashboardDir(), 0o755); err != nil {
		t.Fatalf("创建 Zashboard 目录失败：%v", err)
	}

	archivePath := filepath.Join(t.TempDir(), "zashboard.zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("创建测试压缩包失败：%v", err)
	}
	writer := zip.NewWriter(archive)
	indexFile, err := writer.Create("dist/index.html")
	if err != nil {
		t.Fatalf("创建 index.html 失败：%v", err)
	}
	_, _ = indexFile.Write([]byte("new dashboard"))
	versionFile, err := writer.Create("dist/version.txt")
	if err != nil {
		t.Fatalf("创建 version.txt 失败：%v", err)
	}
	_, _ = versionFile.Write([]byte("v9.9.9\n"))
	if err := writer.Close(); err != nil {
		t.Fatalf("关闭测试压缩包失败：%v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("关闭测试文件失败：%v", err)
	}

	if err := service.installZashboardArchive(archivePath, "zashboard.zip", "v9.9.9"); err != nil {
		t.Fatalf("安装 Zashboard 失败：%v", err)
	}

	reloaded := &Service{cfg: Config{DataDir: dataDir}}
	reloaded.loadInstalledVersions()
	status := reloaded.Status()
	if !status.ZashboardReady || status.ZashboardVersion != "v9.9.9" {
		t.Fatalf("重载后未保留 Zashboard 更新：%+v", status)
	}
}
