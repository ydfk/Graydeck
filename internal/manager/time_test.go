package manager

import (
	"testing"
	"time"
)

func TestFormatLocalTimeUsesConfiguredTimezone(t *testing.T) {
	t.Setenv("GRAYDECK_TIMEZONE", "Asia/Shanghai")

	value := time.Date(2026, 7, 4, 0, 30, 0, 0, time.UTC)
	if actual := formatLocalTime(value); actual != "2026-07-04 08:30:00" {
		t.Fatalf("本地时间格式化错误：%s", actual)
	}
}
