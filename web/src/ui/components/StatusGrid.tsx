import { useI18n } from '@/i18n/I18nProvider'
import type { SystemStatus } from '@/types/api'

type StatusGridProps = {
  status: SystemStatus
}

export function StatusGrid({ status }: StatusGridProps) {
  const { t } = useI18n()

  const items = [
    { label: t('status.coreStatus'), value: status.coreStatus },
    { label: t('status.coreVersion'), value: status.coreVersion },
    { label: t('status.configSource'), value: status.configSource },
    { label: t('status.validation'), value: status.lastValidationResult },
    { label: t('status.update'), value: status.lastUpdateResult },
    { label: t('status.zashboardMode'), value: status.zashboardMode },
  ]

  return (
    <div className="status-grid">
      {items.map((item) => (
        <article className="status-card" key={item.label}>
          <p className="mono-small">{item.label}</p>
          <strong className="status-value">{item.value}</strong>
        </article>
      ))}
    </div>
  )
}
