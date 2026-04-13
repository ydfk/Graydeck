package model

type Subscription struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	URL               string `json:"url"`
	Enabled           bool   `json:"enabled"`
	AutoSync          bool   `json:"autoSync"`
	SyncInterval      string `json:"syncInterval"`
	LastSyncAt        string `json:"lastSyncAt"`
	LastSuccess       string `json:"lastSuccess"`
	LastFailureReason string `json:"lastFailureReason"`
	Status            string `json:"status"`
	PreviewAvailable  bool   `json:"previewAvailable"`
}

type SystemStatus struct {
	RuntimeStatus          string `json:"runtimeStatus"`
	RuntimeError           string `json:"runtimeError"`
	CurrentConfigName      string `json:"currentConfigName"`
	BaseConfigPath         string `json:"baseConfigPath"`
	RuntimeMixedPort       string `json:"runtimeMixedPort"`
	RuntimeControllerAddr  string `json:"runtimeControllerAddr"`
	RuntimeSecret          string `json:"runtimeSecret"`
	CoreVersion            string `json:"coreVersion"`
	CoreLatestVersion      string `json:"coreLatestVersion"`
	CoreIsLatest           bool   `json:"coreIsLatest"`
	CoreExecutableReady    bool   `json:"coreExecutableReady"`
	ZashboardVersion       string `json:"zashboardVersion"`
	ZashboardLatestVersion string `json:"zashboardLatestVersion"`
	ZashboardIsLatest      bool   `json:"zashboardIsLatest"`
	ZashboardReady         bool   `json:"zashboardReady"`
	ZashboardError         string `json:"zashboardError"`
}

type SubscriptionPreview struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type LogEntry struct {
	At      string `json:"at"`
	Message string `json:"message"`
}
