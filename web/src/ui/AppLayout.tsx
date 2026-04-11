import { NavLink, Outlet } from 'react-router-dom'
import { useSubscriptions, useSystemStatus } from '@/api/queries'
import { useI18n } from '@/i18n/I18nProvider'

export function AppLayout() {
  const { locale, localeLabels, setLocale, t } = useI18n()
  const statusQuery = useSystemStatus()
  const subscriptionsQuery = useSubscriptions()

  const navItems = [
    { to: '/', label: t('nav.overview'), end: true },
    { to: '/config', label: t('nav.config') },
    { to: '/updates', label: t('nav.updates') },
    { to: '/system', label: t('nav.system') },
    { to: '/zashboard', label: t('nav.zashboard') },
  ]

  const activeCount = subscriptionsQuery.data?.items.filter((item) => item.enabled).length ?? 0

  return (
    <div className="shell">
      <header className="app-header">
        <div className="brand-lockup">
          <p className="mono-label">{t('app.label')}</p>
          <h1 className="app-title">{t('layout.title')}</h1>
          <div className="header-meta">
            <span>{t('layout.lastApplied')}: {statusQuery.data?.lastAppliedAt ?? t('common.loading')}</span>
            <span>{t('layout.configSource')}: {statusQuery.data?.configSource ?? t('common.loading')}</span>
          </div>
        </div>
        <div className="header-tools">
          <div className="toolbar-group">
            <span className="badge active">{t('layout.safeMode')}</span>
            <span className="badge">{statusQuery.data?.coreVersion ?? t('common.loading')}</span>
            <span className="badge">{t('layout.activeSubscription', { count: activeCount })}</span>
          </div>
          <div className="segmented-control" aria-label="Language switch">
            {(Object.keys(localeLabels) as Array<keyof typeof localeLabels>).map((item) => (
              <button
                key={item}
                className={item === locale ? 'segmented-button active' : 'segmented-button'}
                onClick={() => setLocale(item)}
                type="button"
              >
                {localeLabels[item]}
              </button>
            ))}
          </div>
        </div>
      </header>
      <div className="nav-frame">
        <nav className="nav-tabs" aria-label="Primary">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              className={({ isActive }) => (isActive ? 'nav-pill active' : 'nav-pill')}
              end={item.end}
              to={item.to}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}
