package store

import "mihomo-manager/internal/model"

type MemoryStore struct {
	systemStatus  model.SystemStatus
	subscriptions []model.Subscription
	zashboardMode model.ZashboardMode
}

func NewMemoryStore(mode string) *MemoryStore {
	return &MemoryStore{
		systemStatus: model.SystemStatus{
			CoreStatus:           "running",
			CoreVersion:          "Mihomo Meta v1.19.0",
			ConfigSource:         "remote subscription",
			LastAppliedAt:        "2026-04-11 20:30:00",
			LastUpdateResult:     "Core is up to date",
			LastValidationResult: "Validation passed",
			ZashboardMode:        mode,
		},
		subscriptions: []model.Subscription{
			{
				ID:               "primary",
				Name:             "Primary Remote",
				URL:              "https://example.com/subscriptions/primary.yaml",
				Enabled:          true,
				AutoSync:         true,
				SyncInterval:     "30m",
				ApplyPolicy:      "auto_apply_if_valid",
				LastSyncAt:       "2026-04-11 20:25:00",
				LastSuccess:      "2026-04-11 20:25:00",
				CandidateVersion: "active",
			},
			{
				ID:                "backup",
				Name:              "Backup Candidate",
				URL:               "https://example.com/subscriptions/backup.yaml",
				Enabled:           false,
				AutoSync:          true,
				SyncInterval:      "2h",
				ApplyPolicy:       "candidate_only",
				LastSyncAt:        "2026-04-11 18:00:00",
				LastSuccess:       "2026-04-11 18:00:00",
				LastFailureReason: "",
				CandidateVersion:  "2026-04-11-18-00",
			},
		},
		zashboardMode: model.ZashboardMode{
			Mode: "safe",
			AllowedFeatures: []string{
				"overview",
				"proxies_view",
				"proxy_group_switch",
				"latency_test",
				"traffic_view",
				"connections_view",
				"logs_view",
				"rules_view",
			},
			BlockedFeatures: []string{
				"setup_editor",
				"core_upgrade",
				"tun_toggle",
				"config_mutation",
				"sensitive_settings",
			},
			URLFlags: []string{
				"disableUpgradeCore=1",
				"disableTunMode=1",
			},
			AllowedWriteScopes: []string{
				"proxy_group_switch",
			},
		},
	}
}

func (s *MemoryStore) GetSystemStatus() model.SystemStatus {
	return s.systemStatus
}

func (s *MemoryStore) GetSubscriptions() []model.Subscription {
	return s.subscriptions
}

func (s *MemoryStore) GetZashboardMode() model.ZashboardMode {
	return s.zashboardMode
}
