package manager

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mihomo-manager/internal/model"
)

func (s *Service) autoSyncLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.syncStopCh:
			return
		case <-ticker.C:
			s.tickAutoSync()
		}
	}
}

func (s *Service) tickAutoSync() {
	s.mu.RLock()
	candidates := make([]model.Subscription, len(s.subscriptions))
	copy(candidates, s.subscriptions)
	s.mu.RUnlock()

	for _, subscription := range candidates {
		if automaticSyncDue(subscription, time.Now()) {
			s.syncIfDue(subscription)
		}
	}
}

func automaticSyncDue(subscription model.Subscription, now time.Time) bool {
	if !subscription.AutoSync || subscription.SyncInterval == "" {
		return false
	}

	interval, err := parseInterval(subscription.SyncInterval)
	if err != nil || interval <= 0 {
		return false
	}
	if subscription.LastSyncAt == "" {
		return true
	}

	lastSync, err := time.ParseInLocation("2006-01-02 15:04:05", subscription.LastSyncAt, time.Local)
	if err != nil {
		return true
	}

	return now.Sub(lastSync) >= interval
}

func (s *Service) syncIfDue(subscription model.Subscription) {
	s.appendLogf("自动同步配置文件：%s", subscription.Name)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if _, err := s.syncSubscriptionWithTrigger(ctx, subscription.ID, "auto"); err != nil {
		s.appendLogf("自动同步失败：%s，%v", subscription.Name, err)
		return
	}

	updated, err := s.findSubscription(subscription.ID)
	if err == nil && updated.Enabled {
		if err := s.ensureRuntime(ctx); err != nil {
			s.appendLogf("自动同步后应用失败：%s，%v", subscription.Name, err)
		}
	}
}

func parseInterval(raw string) (time.Duration, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" || trimmed == "disabled" || trimmed == "0" || trimmed == "off" {
		return 0, nil
	}

	for suffix, hours := range map[string]float64{"d": 24, "w": 24 * 7} {
		if !strings.HasSuffix(trimmed, suffix) {
			continue
		}

		value, err := strconv.ParseFloat(strings.TrimSuffix(trimmed, suffix), 64)
		if err != nil || value <= 0 {
			return 0, fmt.Errorf("无效的同步频率：%s", raw)
		}
		return time.Duration(value * hours * float64(time.Hour)), nil
	}

	return time.ParseDuration(trimmed)
}

func validateSyncInterval(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	interval, err := parseInterval(trimmed)
	if err != nil {
		return "", errors.New("同步频率格式无效，请使用 disabled、15m、2h、3d 等格式")
	}
	if interval < 0 {
		return "", errors.New("同步频率超出支持范围")
	}
	if interval > 0 && interval < time.Minute {
		return "", errors.New("自动同步间隔不能小于 1 分钟")
	}
	if interval == 0 {
		return "disabled", nil
	}

	return strings.ToLower(trimmed), nil
}
