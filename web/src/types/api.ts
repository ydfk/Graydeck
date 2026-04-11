export type SystemStatus = {
  coreStatus: string
  coreVersion: string
  configSource: string
  lastAppliedAt: string
  lastUpdateResult: string
  lastValidationResult: string
  zashboardMode: string
}

export type Subscription = {
  id: string
  name: string
  url: string
  enabled: boolean
  autoSync: boolean
  syncInterval: string
  applyPolicy: string
  lastSyncAt: string
  lastSuccess: string
  lastFailureReason: string
  candidateVersion: string
}

export type SubscriptionListResponse = {
  items: Subscription[]
}

export type ZashboardMode = {
  mode: string
  allowedFeatures: string[]
  blockedFeatures: string[]
  urlFlags: string[]
  allowedWriteScopes: string[]
}
