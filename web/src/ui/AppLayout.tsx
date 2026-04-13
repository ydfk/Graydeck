import { NavLink, Outlet } from 'react-router-dom'
import { useI18n } from '@/i18n/I18nProvider'

export function AppLayout() {
  const { locale, localeLabels, setLocale, t } = useI18n()

  const navItems = [
    { to: '/', label: t('nav.overview'), end: true },
    { to: '/zashboard', label: t('nav.zashboard') },
    { to: '/logs', label: t('nav.logs') },
  ]

  return (
    <div className="dashboard-shell">
      <header className="topbar">
        <div className="topbar-brand">
          <div>
            <h1 className="app-title">{t('layout.title')}</h1>
          </div>
          <p className="app-subtitle">{t('layout.subtitle')}</p>
        </div>
        <div className="topbar-actions">
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
      <div className="workspace-shell">
        <nav className="nav-row" aria-label="Primary">
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
        <main className="workspace-content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
