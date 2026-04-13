export type SystemStatus = {
  runtimeStatus: string;
  runtimeError: string;
  currentConfigName: string;
  baseConfigPath: string;
  runtimeMixedPort: string;
  runtimeControllerAddr: string;
  runtimeSecret: string;
  coreVersion: string;
  coreLatestVersion: string;
  coreIsLatest: boolean;
  coreExecutableReady: boolean;
  zashboardVersion: string;
  zashboardLatestVersion: string;
  zashboardIsLatest: boolean;
  zashboardReady: boolean;
  zashboardError: string;
};

export type Subscription = {
  id: string;
  name: string;
  url: string;
  enabled: boolean;
  autoSync: boolean;
  syncInterval: string;
  lastSyncAt: string;
  lastSuccess: string;
  lastFailureReason: string;
  status: string;
  previewAvailable: boolean;
};

export type SubscriptionListResponse = {
  items: Subscription[];
};

export type SubscriptionPreview = {
  id: string;
  name: string;
  content: string;
};

export type RuntimeConfigPreview = {
  name: string;
  path: string;
  content: string;
};

export type LogEntry = {
  at: string;
  message: string;
};

export type LogListResponse = {
  items: LogEntry[];
};
