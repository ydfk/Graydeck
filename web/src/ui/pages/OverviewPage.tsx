import { useSubscriptions, useSystemStatus } from '@/api/queries'
import { useI18n } from '@/i18n/I18nProvider'
import { Panel } from '@/ui/components/Panel'
import { StatusGrid } from '@/ui/components/StatusGrid'
import { SubscriptionTable } from '@/ui/components/SubscriptionTable'

export function OverviewPage() {
  const { t } = useI18n()
  const statusQuery = useSystemStatus()
  const subscriptionsQuery = useSubscriptions()

  return (
    <div className="page-grid">
      <Panel eyebrow={t('nav.overview')} title={t('overview.runtime.title')}>
        {statusQuery.data ? <StatusGrid status={statusQuery.data} /> : <p className="body-copy">{t('common.loading')}</p>}
      </Panel>
      <div className="page-grid page-grid-two-column">
        <Panel eyebrow={t('panel.notes')} title={t('overview.rules.title')}>
          <ul className="stack-list">
            <li className="list-item">
              <div>
                <strong>{t('overview.rule.active')}</strong>
              </div>
            </li>
            <li className="list-item">
              <div>
                <strong>{t('overview.rule.candidate')}</strong>
              </div>
            </li>
            <li className="list-item">
              <div>
                <strong>{t('overview.rule.safe')}</strong>
              </div>
            </li>
          </ul>
        </Panel>
        <Panel eyebrow={t('panel.quickActions')} title={t('config.sources.title')}>
          {subscriptionsQuery.data ? (
            <ul className="stack-list">
              {subscriptionsQuery.data.items.map((subscription) => (
                <li className="list-item" key={subscription.id}>
                  <div>
                    <strong>{subscription.name}</strong>
                    <p className="body-copy">{subscription.url}</p>
                    <p className="table-secondary">
                      {t('subscription.interval')}: {subscription.syncInterval}
                    </p>
                  </div>
                  <span className={subscription.enabled ? 'badge active' : 'badge'}>
                    {subscription.enabled ? t('common.active') : t('common.candidate')}
                  </span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="body-copy">{t('common.loading')}</p>
          )}
        </Panel>
      </div>
      <Panel eyebrow={t('nav.config')} title={t('overview.subscriptions.title')}>
        {subscriptionsQuery.data ? (
          <SubscriptionTable subscriptions={subscriptionsQuery.data.items} />
        ) : (
          <p className="body-copy">{t('common.loading')}</p>
        )}
      </Panel>
    </div>
  )
}
