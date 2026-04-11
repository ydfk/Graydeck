import { useForm } from 'react-hook-form'
import { useSubscriptions } from '@/api/queries'
import { useI18n } from '@/i18n/I18nProvider'
import { Panel } from '@/ui/components/Panel'

type ConfigFormValues = {
  syncInterval: string
  draft: string
}

export function ConfigPage() {
  const { t } = useI18n()
  const subscriptionsQuery = useSubscriptions()
  const form = useForm<ConfigFormValues>({
    defaultValues: {
      syncInterval: '30m',
      draft: '',
    },
  })

  return (
    <div className="page-grid">
      <Panel eyebrow={t('nav.config')} title={t('config.sources.title')}>
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
                  <p className="table-secondary">
                    {t('subscription.lastFailure')}: {subscription.lastFailureReason || t('subscription.emptyFailure')}
                  </p>
                </div>
                <span className={subscription.enabled ? 'badge active' : 'badge'}>
                  {subscription.enabled ? t('common.enabled') : t('common.candidate')}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="body-copy">{t('common.loading')}</p>
        )}
      </Panel>
      <div className="page-grid page-grid-two-column">
        <Panel eyebrow={t('panel.notes')} title={t('config.sync.title')}>
          <form className="editor-form">
            <label className="field">
              <span className="field-label">{t('config.field.interval')}</span>
              <input className="text-input" {...form.register('syncInterval')} />
            </label>
            <div className="button-row">
              <button className="primary-pill" type="button">
                {t('config.button.validate')}
              </button>
              <button className="secondary-pill" type="button">
                {t('config.button.save')}
              </button>
            </div>
          </form>
        </Panel>
        <Panel eyebrow={t('panel.quickActions')} title={t('config.draft.title')}>
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
          </ul>
        </Panel>
      </div>
      <Panel eyebrow={t('common.saved')} title={t('config.draft.title')}>
        <form className="editor-form">
          <label className="field">
            <span className="field-label">{t('config.field.draft')}</span>
            <textarea className="text-area" rows={14} {...form.register('draft')} />
          </label>
          <div className="button-row">
            <button className="primary-pill" type="button">
              {t('config.button.validate')}
            </button>
            <button className="secondary-pill" type="button">
              {t('config.button.save')}
            </button>
          </div>
        </form>
      </Panel>
    </div>
  )
}
