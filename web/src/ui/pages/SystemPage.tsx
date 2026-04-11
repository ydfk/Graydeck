import { useSystemStatus } from '@/api/queries'
import { useI18n } from '@/i18n/I18nProvider'
import { Panel } from '@/ui/components/Panel'

export function SystemPage() {
  const { t } = useI18n()
  const statusQuery = useSystemStatus()

  return (
    <div className="page-grid page-grid-two-column">
      <Panel
        eyebrow={t('nav.system')}
        title={t('system.runtime.title')}
      >
        <div className="metric-stack">
          <div>
            <p className="mono-small">{t('status.coreStatus')}</p>
            <strong className="status-value">{statusQuery.data?.coreStatus ?? t('common.loading')}</strong>
          </div>
          <div>
            <p className="mono-small">{t('status.lastApplied')}</p>
            <strong className="status-value">{statusQuery.data?.lastAppliedAt ?? t('common.loading')}</strong>
          </div>
        </div>
      </Panel>
      <Panel
        eyebrow={t('panel.notes')}
        title={t('system.boundary.title')}
      >
        <ul className="stack-list">
          <li className="list-item">
            <div>
              <strong>{t('system.boundary.controller')}</strong>
            </div>
          </li>
          <li className="list-item">
            <div>
              <strong>{t('system.boundary.config')}</strong>
            </div>
          </li>
          <li className="list-item">
            <div>
              <strong>{t('system.boundary.audit')}</strong>
            </div>
          </li>
        </ul>
      </Panel>
      <Panel
        eyebrow={t('panel.quickActions')}
        title={t('system.logs.title')}
      >
        <div className="metric-stack">
          <div>
            <p className="mono-small">{t('status.validation')}</p>
            <strong className="status-value">{statusQuery.data?.lastValidationResult ?? t('common.loading')}</strong>
          </div>
          <div>
            <p className="mono-small">{t('status.update')}</p>
            <strong className="status-value">{statusQuery.data?.lastUpdateResult ?? t('common.loading')}</strong>
          </div>
          <div>
            <p className="mono-small">{t('status.configSource')}</p>
            <strong className="status-value">{statusQuery.data?.configSource ?? t('common.loading')}</strong>
          </div>
        </div>
      </Panel>
    </div>
  )
}
