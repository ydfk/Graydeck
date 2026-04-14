import { useMutation, useQueryClient } from "@tanstack/react-query";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { apiPost } from "@/api/client";
import { useAuthStatus } from "@/api/queries";
import { useI18n } from "@/i18n/I18nProvider";

export function AppLayout() {
  const { locale, localeLabels, setLocale, t } = useI18n();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const authQuery = useAuthStatus();

  const logoutMutation = useMutation({
    mutationFn: () => apiPost("/api/auth/logout"),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["auth-status"] });
      navigate("/login", { replace: true });
    },
  });

  const navItems = [
    { to: "/", label: t("nav.overview"), end: true },
    { to: "/zashboard", label: t("nav.zashboard") },
    { to: "/logs", label: t("nav.logs") },
  ];

  return (
    <div className="dashboard-shell">
      <header className="topbar">
        <div className="topbar-brand">
          <div>
            <h1 className="app-title">{t("layout.title")}</h1>
          </div>
          <p className="app-subtitle">{t("layout.subtitle")}</p>
        </div>
        <div className="topbar-actions">
          <div className="segmented-control" aria-label="Language switch">
            {(Object.keys(localeLabels) as Array<keyof typeof localeLabels>).map((item) => (
              <button
                key={item}
                className={item === locale ? "segmented-button active" : "segmented-button"}
                onClick={() => setLocale(item)}
                type="button"
              >
                {localeLabels[item]}
              </button>
            ))}
          </div>
          <button className="secondary-pill" disabled={logoutMutation.isPending} onClick={() => logoutMutation.mutate()} type="button">
            {logoutMutation.isPending ? t("common.loading") : t("auth.logout")}
          </button>
          {authQuery.data?.username ? <span className="auth-user">{authQuery.data.username}</span> : null}
        </div>
      </header>
      <div className="workspace-shell">
        <nav aria-label="Primary" className="nav-row">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              className={({ isActive }) => (isActive ? "nav-pill active" : "nav-pill")}
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
  );
}
