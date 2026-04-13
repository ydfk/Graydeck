import { useQuery } from '@tanstack/react-query'
import { apiGet } from './client'
import type { LogListResponse, SubscriptionListResponse, SystemStatus } from '@/types/api'

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

export function useLogs() {
  return useQuery({
    queryKey: ['logs'],
    queryFn: () => apiGet<LogListResponse>('/api/logs'),
    refetchInterval: 2000,
  })
}
