import { useQuery } from '@tanstack/react-query'
import { apiGet } from './client'
import type { SubscriptionListResponse, SystemStatus, ZashboardMode } from '@/types/api'

export function useSystemStatus() {
  return useQuery({
    queryKey: ['system-status'],
    queryFn: () => apiGet<SystemStatus>('/api/system/status'),
  })
}

export function useSubscriptions() {
  return useQuery({
    queryKey: ['subscriptions'],
    queryFn: () => apiGet<SubscriptionListResponse>('/api/subscriptions'),
  })
}

export function useZashboardMode() {
  return useQuery({
    queryKey: ['zashboard-mode'],
    queryFn: () => apiGet<ZashboardMode>('/api/zashboard/mode'),
  })
}
