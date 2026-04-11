import { useI18n } from '@/i18n/I18nProvider'
import type { Subscription } from '@/types/api'

type SubscriptionTableProps = {
  subscriptions: Subscription[]
}

export function SubscriptionTable({ subscriptions }: SubscriptionTableProps) {
  const { t } = useI18n()

  function getRoleLabel(enabled: boolean) {
    return enabled ? t('subscription.role.active') : t('subscription.role.candidate')
  }

  function getPolicyLabel(policy: string) {
    return policy.includes('auto-apply') ? t('subscription.policy.autoApply') : t('subscription.policy.manual')
  }

  return (
    <div className="table-shell">
      <table className="table">
        <thead>
          <tr>
            <th>{t('subscription.name')}</th>
            <th>{t('subscription.role')}</th>
            <th>{t('subscription.sync')}</th>
            <th>{t('subscription.applyPolicy')}</th>
            <th>{t('subscription.lastSuccess')}</th>
            <th>{t('subscription.candidateVersion')}</th>
          </tr>
        </thead>
        <tbody>
          {subscriptions.map((subscription) => (
            <tr key={subscription.id}>
              <td>
                <div className="table-primary">{subscription.name}</div>
                <div className="table-secondary">{subscription.url}</div>
              </td>
              <td>
                <span className={subscription.enabled ? 'badge active' : 'badge'}>{getRoleLabel(subscription.enabled)}</span>
              </td>
              <td>
                <div className="table-primary">{subscription.syncInterval}</div>
                <div className="table-secondary">
                  {subscription.autoSync ? t('common.autoSyncOn') : t('common.autoSyncOff')}
                </div>
              </td>
              <td>{getPolicyLabel(subscription.applyPolicy)}</td>
              <td>{subscription.lastSuccess}</td>
              <td>{subscription.candidateVersion}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
