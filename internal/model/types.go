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
	RuntimeMixedPort       string `json:"runtimeMixedPort"`
	RuntimeSocksPort       string `json:"runtimeSocksPort"`
	RuntimeRedirPort       string `json:"runtimeRedirPort"`
	RuntimeTProxyPort      string `json:"runtimeTProxyPort"`
	CoreVersion            string `json:"coreVersion"`
	CoreLatestVersion      string `json:"coreLatestVersion"`
	CoreIsLatest           bool   `json:"coreIsLatest"`
	CoreExecutableReady    bool   `json:"coreExecutableReady"`
	ZashboardVersion       string `json:"zashboardVersion"`
	ZashboardLatestVersion string `json:"zashboardLatestVersion"`
	ZashboardIsLatest      bool   `json:"zashboardIsLatest"`
	ZashboardReady         bool   `json:"zashboardReady"`
	ZashboardError         string `json:"zashboardError"`
	ZashboardHideSettings  bool   `json:"zashboardHideSettings"`
}

type SubscriptionPreview struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type RuntimeConfigPreview struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type DefaultRuntimeConfig struct {
	Path       string `json:"path"`
	MixedPort  string `json:"mixedPort"`
	SocksPort  string `json:"socksPort"`
	RedirPort  string `json:"redirPort"`
	TProxyPort string `json:"tproxyPort"`
}

type LogEntry struct {
	At      string `json:"at"`
	Message string `json:"message"`
}
