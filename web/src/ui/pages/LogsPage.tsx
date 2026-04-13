import { useLogs } from '@/api/queries'
import { useI18n } from '@/i18n/I18nProvider'
import { Panel } from '@/ui/components/Panel'

export function LogsPage() {
  const { t } = useI18n()
  const logsQuery = useLogs()
  const items = logsQuery.data?.items ?? []

  return (
    <div className="page-grid">
      <Panel title={t('logs.title')}>
        {items.length > 0 ? (
          <div className="log-shell">
            {items.map((item, index) => (
              <div key={`${item.at}-${index}`} className="log-line">
                <span className="log-time">{item.at}</span>
                <span className="log-message">{item.message}</span>
              </div>
            ))}
          </div>
        ) : (
          <p className="body-copy">{t('logs.empty')}</p>
        )}
      </Panel>
    </div>
  )
}
