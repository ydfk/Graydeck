import { useEffect, useState } from 'react'
import { useI18n } from '@/i18n/I18nProvider'
import type { Subscription } from '@/types/api'

type EditableFields = {
  name: string
  url: string
  syncInterval: string
}

type SubscriptionTableProps = {
  subscriptions: Subscription[]
  switchingId: string | null
  savingId: string | null
  syncingId: string | null
  previewingId: string | null
  creating: boolean
  showCreateForm: boolean
  onCreate: (fields: EditableFields) => void
  onUpdate: (id: string, fields: EditableFields) => void
  onSync: (id: string) => void
  onActivate: (id: string) => void
  onPreview: (id: string) => void
  onHideCreateForm: () => void
}

const emptyCreateForm: EditableFields = {
  name: '',
  url: '',
  syncInterval: '30m',
}

export function SubscriptionTable({
  subscriptions,
  switchingId,
  savingId,
  syncingId,
  previewingId,
  creating,
  showCreateForm,
  onCreate,
  onUpdate,
  onSync,
  onActivate,
  onPreview,
  onHideCreateForm,
}: SubscriptionTableProps) {
  const { t } = useI18n()
  const [rows, setRows] = useState(subscriptions)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [createForm, setCreateForm] = useState<EditableFields>(emptyCreateForm)

  useEffect(() => {
    setRows(subscriptions)
  }, [subscriptions])

  useEffect(() => {
    if (!showCreateForm) {
      clearCreateForm()
    }
  }, [showCreateForm])

  function updateRow(id: string, field: keyof EditableFields, value: string) {
    setRows((current) =>
      current.map((item) => (item.id === id ? { ...item, [field]: value } : item)),
    )
  }

  function updateCreateForm(field: keyof EditableFields, value: string) {
    setCreateForm((current) => ({ ...current, [field]: value }))
  }

  function clearCreateForm() {
    setCreateForm(emptyCreateForm)
  }

  function cancelEdit() {
    setRows(subscriptions)
    setEditingId(null)
  }

  function getStatusLabel(status: string) {
    const mapping: Record<string, string> = {
      active: t('subscription.active'),
      ready: t('subscription.ready'),
      fetch_failed: t('subscription.fetchFailed'),
      validation_failed: t('subscription.validationFailed'),
      pending: t('status.pending'),
    }

    return mapping[status] ?? status
  }

  function getStatusBadgeClass(status: string, enabled: boolean) {
    if (enabled || status === 'active') {
      return 'badge active'
    }

    if (status === 'fetch_failed' || status === 'validation_failed') {
      return 'badge danger'
    }

    if (status === 'pending') {
      return 'badge warning'
    }

    return 'badge'
  }

  return (
    <div className="table-stack">
      {(switchingId || savingId || syncingId || previewingId || creating) && (
        <div className="status-notice">{t('common.loading')}</div>
      )}
      {showCreateForm ? (
        <div className="create-card">
          <div className="create-grid">
            <input
              className="table-input"
              onChange={(event) => updateCreateForm('name', event.target.value)}
              placeholder={t('config.add.name')}
              value={createForm.name}
            />
            <input
              className="table-input"
              onChange={(event) => updateCreateForm('url', event.target.value)}
              placeholder={t('config.add.url')}
              value={createForm.url}
            />
            <input
              className="table-input"
              onChange={(event) => updateCreateForm('syncInterval', event.target.value)}
              placeholder={t('config.add.syncInterval')}
              value={createForm.syncInterval}
            />
          </div>
          <div className="table-actions">
            <button className="primary-pill table-action-button" disabled={creating} onClick={() => onCreate(createForm)} type="button">
              {creating ? t('common.loading') : t('common.confirm')}
            </button>
            <button
              className="secondary-pill table-action-button"
              onClick={() => {
                clearCreateForm()
                onHideCreateForm()
              }}
              type="button"
            >
              {t('subscription.cancel')}
            </button>
          </div>
        </div>
      ) : null}
      <div className="table-shell">
        <table className="table">
          <thead>
            <tr>
              <th>{t('config.name')}</th>
              <th>{t('config.url')}</th>
              <th>{t('config.syncInterval')}</th>
              <th>{t('config.lastSuccess')}</th>
              <th>{t('config.status')}</th>
              <th>{t('config.error')}</th>
              <th>{t('config.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr>
                <td className="table-empty" colSpan={7}>
                  {t('common.none')}
                </td>
              </tr>
            ) : null}
            {rows.map((subscription) => (
              <tr key={subscription.id}>
                <td>
                  {editingId === subscription.id ? (
                    <input
                      className="table-input"
                      onChange={(event) => updateRow(subscription.id, 'name', event.target.value)}
                      value={subscription.name}
                    />
                  ) : (
                    <div className="table-primary">{subscription.name}</div>
                  )}
                </td>
                <td>
                  {editingId === subscription.id ? (
                    <input
                      className="table-input"
                      onChange={(event) => updateRow(subscription.id, 'url', event.target.value)}
                      value={subscription.url}
                    />
                  ) : (
                    <div className="table-secondary">{subscription.url}</div>
                  )}
                </td>
                <td>
                  {editingId === subscription.id ? (
                    <input
                      className="table-input table-input-compact"
                      onChange={(event) => updateRow(subscription.id, 'syncInterval', event.target.value)}
                      value={subscription.syncInterval}
                    />
                  ) : (
                    <div className="table-primary">{subscription.syncInterval}</div>
                  )}
                </td>
                <td>{subscription.lastSuccess || t('common.empty')}</td>
                <td>
                  <span className={getStatusBadgeClass(subscription.status, subscription.enabled)}>
                    {getStatusLabel(subscription.status)}
                  </span>
                </td>
                <td className={subscription.lastFailureReason ? 'table-secondary danger' : 'table-secondary'}>
                  {subscription.lastFailureReason || t('common.empty')}
                </td>
                <td>
                  <div className="table-actions">
                    {editingId === subscription.id ? (
                      <>
                        <button
                          className="primary-pill table-action-button"
                          disabled={savingId === subscription.id}
                          onClick={() =>
                            onUpdate(subscription.id, {
                              name: subscription.name,
                              url: subscription.url,
                              syncInterval: subscription.syncInterval,
                            })
                          }
                          type="button"
                        >
                          {savingId === subscription.id ? t('common.loading') : t('subscription.save')}
                        </button>
                        <button className="secondary-pill table-action-button" onClick={cancelEdit} type="button">
                          {t('subscription.cancel')}
                        </button>
                      </>
                    ) : (
                      <>
                        <button
                          className="secondary-pill table-action-button"
                          onClick={() => setEditingId(subscription.id)}
                          type="button"
                        >
                          {t('subscription.edit')}
                        </button>
                        <button
                          className="secondary-pill table-action-button"
                          disabled={syncingId === subscription.id}
                          onClick={() => onSync(subscription.id)}
                          type="button"
                        >
                          {syncingId === subscription.id ? t('common.loading') : t('subscription.sync')}
                        </button>
                        <button
                          className="secondary-pill table-action-button"
                          disabled={previewingId === subscription.id || !subscription.previewAvailable}
                          onClick={() => onPreview(subscription.id)}
                          type="button"
                        >
                          {previewingId === subscription.id ? t('common.loading') : t('subscription.preview')}
                        </button>
                        {subscription.enabled ? (
                          <button className="primary-pill table-action-button" disabled type="button">
                            {t('subscription.current')}
                          </button>
                        ) : (
                          <button
                            className="primary-pill table-action-button"
                            disabled={switchingId === subscription.id}
                            onClick={() => onActivate(subscription.id)}
                            type="button"
                          >
                            {switchingId === subscription.id ? t('common.loading') : t('subscription.switch')}
                          </button>
                        )}
                      </>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
