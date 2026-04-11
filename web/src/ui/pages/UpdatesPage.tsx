import { useSystemStatus, useZashboardMode } from '@/api/queries'
import { useI18n } from '@/i18n/I18nProvider'
import { Panel } from '@/ui/components/Panel'

export function UpdatesPage() {
  const { t } = useI18n()
  const statusQuery = useSystemStatus()
  const zashboardModeQuery = useZashboardMode()

  return (
    <div className="page-grid page-grid-two-column">
      <Panel eyebrow={t('nav.updates')} title={t('updates.core.title')}>
        <div className="metric-stack">
          <div>
            <p className="mono-small">{t('status.coreVersion')}</p>
            <strong className="status-value">{statusQuery.data?.coreVersion ?? t('common.loading')}</strong>
          </div>
          <div>
            <p className="mono-small">{t('metric.channel')}</p>
            <strong className="status-value">Meta</strong>
          </div>
          <div>
            <p className="mono-small">{t('metric.lastResult')}</p>
            <strong className="status-value">{statusQuery.data?.lastUpdateResult ?? t('common.loading')}</strong>
          </div>
        </div>
      </Panel>
      <Panel eyebrow={t('panel.notes')} title={t('updates.rules.title')}>
        <ul className="stack-list">
          <li className="list-item">
            <div>
              <strong>{t('updates.rule.core')}</strong>
            </div>
          </li>
          <li className="list-item">
            <div>
              <strong>{t('updates.rule.apply')}</strong>
            </div>
          </li>
          <li className="list-item">
            <div>
              <strong>{t('updates.rule.hold')}</strong>
            </div>
          </li>
        </ul>
      </Panel>
      <Panel eyebrow={t('nav.zashboard')} title={t('updates.zashboard.title')}>
        <div className="metric-stack">
          <div>
            <p className="mono-small">{t('zashboard.mode')}</p>
            <strong className="status-value">{zashboardModeQuery.data?.mode ?? t('common.loading')}</strong>
          </div>
          <div>
            <p className="mono-small">{t('metric.allowedWriteScopes')}</p>
            <strong className="status-value">
              {zashboardModeQuery.data?.allowedWriteScopes.join(', ') ?? t('common.loading')}
            </strong>
          </div>
        </div>
      </Panel>
      <Panel eyebrow={t('panel.allowed')} title={t('updates.zashboard.title')}>
        <div className="page-grid page-grid-two-column compact-grid">
          <ul className="stack-list">
            {(zashboardModeQuery.data?.allowedFeatures ?? []).map((item) => (
              <li className="list-item" key={item}>
                <div>
                  <strong>{item}</strong>
                </div>
              </li>
            ))}
          </ul>
          <ul className="stack-list">
            {(zashboardModeQuery.data?.blockedFeatures ?? []).map((item) => (
              <li className="list-item" key={item}>
                <div>
                  <strong>{item}</strong>
                </div>
              </li>
            ))}
          </ul>
        </div>
      </Panel>
    </div>
  )
}
