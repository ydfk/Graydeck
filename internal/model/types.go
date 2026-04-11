package model

type Subscription struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	URL               string `json:"url"`
	Enabled           bool   `json:"enabled"`
	AutoSync          bool   `json:"autoSync"`
	SyncInterval      string `json:"syncInterval"`
	ApplyPolicy       string `json:"applyPolicy"`
	LastSyncAt        string `json:"lastSyncAt"`
	LastSuccess       string `json:"lastSuccess"`
	LastFailureReason string `json:"lastFailureReason"`
	CandidateVersion  string `json:"candidateVersion"`
}

type SystemStatus struct {
	CoreStatus           string `json:"coreStatus"`
	CoreVersion          string `json:"coreVersion"`
	ConfigSource         string `json:"configSource"`
	LastAppliedAt        string `json:"lastAppliedAt"`
	LastUpdateResult     string `json:"lastUpdateResult"`
	LastValidationResult string `json:"lastValidationResult"`
	ZashboardMode        string `json:"zashboardMode"`
}

type ZashboardMode struct {
	Mode               string   `json:"mode"`
	AllowedFeatures    []string `json:"allowedFeatures"`
	BlockedFeatures    []string `json:"blockedFeatures"`
	URLFlags           []string `json:"urlFlags"`
	AllowedWriteScopes []string `json:"allowedWriteScopes"`
}
