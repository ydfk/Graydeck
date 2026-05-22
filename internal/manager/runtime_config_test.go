package manager

import (
	"strings"
	"testing"
)

func TestStripTopLevelKeysRemovesYAMLBlock(t *testing.T) {
	content := strings.TrimSpace(`
dns:
  enable: false
  nameserver:
    - 1.1.1.1
tun:
  enable: false
proxies:
  - name: direct
`)

	stripped := stripTopLevelKeys(content, "dns", "tun")

	if strings.Contains(stripped, "dns:") || strings.Contains(stripped, "tun:") {
		t.Fatalf("同名顶层配置块未移除：%s", stripped)
	}

	if !strings.Contains(stripped, "proxies:") {
		t.Fatalf("后续配置块被误删：%s", stripped)
	}
}

func TestMergeBaseRuntimeSectionsKeepsMissingNestedFields(t *testing.T) {
	baseConfig := strings.TrimSpace(`
dns:
  enable: true
  listen: 0.0.0.0:53
tun:
  enable: true
  auto-route: true
`)

	subscriptionConfig := strings.TrimSpace(`
dns:
  enable: false
  nameserver:
    - 1.1.1.1
tun:
  enable: false
  stack: gvisor
`)

	section, keys, err := mergeBaseRuntimeSections(baseConfig, subscriptionConfig, "dns", "tun")
	if err != nil {
		t.Fatalf("合并覆盖配置失败：%v", err)
	}

	if len(keys) != 2 {
		t.Fatalf("覆盖配置块识别错误：%v", keys)
	}

	for _, expected := range []string{
		"enable: true",
		"listen: 0.0.0.0:53",
		"nameserver:",
		"1.1.1.1",
		"auto-route: true",
		"stack: gvisor",
	} {
		if !strings.Contains(section, expected) {
			t.Fatalf("合并结果缺少 %q：%s", expected, section)
		}
	}
}
