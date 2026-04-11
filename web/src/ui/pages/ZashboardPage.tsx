import { useZashboardMode } from '@/api/queries'
import { useI18n } from '@/i18n/I18nProvider'
import { useUiStore } from '@/store/ui-store'
import { Panel } from '@/ui/components/Panel'

export function ZashboardPage() {
  const { t } = useI18n()
  const modeQuery = useZashboardMode()
  const frameMode = useUiStore((state) => state.zashboardFrameMode)
  const setFrameMode = useUiStore((state) => state.setZashboardFrameMode)

  return (
    <div className="page-grid">
      <Panel
        eyebrow={t('nav.zashboard')}
        title={t('zashboard.shell.title')}
        actions={
          <div className="button-row">
            <button className="secondary-pill" onClick={() => setFrameMode('embedded')} type="button">
              {t('common.embedded')}
            </button>
            <button className="primary-pill" onClick={() => setFrameMode('fullscreen')} type="button">
              {t('common.fullscreen')}
            </button>
          </div>
        }
      >
        <div className="metric-stack">
          <div>
            <p className="mono-small">{t('zashboard.mode')}</p>
            <strong className="status-value">{modeQuery.data?.mode ?? t('common.loading')}</strong>
          </div>
          <div>
            <p className="mono-small">{t('zashboard.flags')}</p>
            <strong className="status-value">{modeQuery.data?.urlFlags.join(' · ') ?? t('common.loading')}</strong>
          </div>
          <div>
            <p className="mono-small">{t('metric.frame')}</p>
            <strong className="status-value">{frameMode}</strong>
          </div>
        </div>
        <div className="zashboard-preview">
          <div className="metric-stack">
            <div>
              <p className="mono-small">{t('zashboard.entry')}</p>
              <strong className="status-value">/zashboard/</strong>
            </div>
            <div>
              <p className="mono-small">{t('metric.frame')}</p>
              <strong className="status-value">{frameMode}</strong>
            </div>
          </div>
        </div>
      </Panel>
      <Panel
        eyebrow={t('panel.notes')}
        title={t('zashboard.restrictions.title')}
      >
        <div className="page-grid page-grid-two-column compact-grid">
          <div>
            <p className="mono-small">{t('panel.allowed')}</p>
            <ul className="stack-list">
              {(modeQuery.data?.allowedFeatures ?? []).map((item) => (
                <li className="list-item" key={item}>
                  <div>
                    <strong>{item}</strong>
                  </div>
                </li>
              ))}
            </ul>
          </div>
          <div>
            <p className="mono-small">{t('panel.blocked')}</p>
            <ul className="stack-list">
              {(modeQuery.data?.blockedFeatures ?? []).map((item) => (
                <li className="list-item" key={item}>
                  <div>
                    <strong>{item}</strong>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </Panel>
    </div>
  )
}
