import { Navigate, createBrowserRouter, useLocation } from "react-router-dom";
import { useAuthStatus } from "@/api/queries";
import { useI18n } from "@/i18n/I18nProvider";
import { AppLayout } from "@/ui/AppLayout";
import { LogsPage } from "@/ui/pages/LogsPage";
import { LoginPage } from "@/ui/pages/LoginPage";
import { OverviewPage } from "@/ui/pages/OverviewPage";
import { ZashboardPage } from "@/ui/pages/ZashboardPage";

function AuthLoading() {
  const { t } = useI18n();
  return <div className="auth-loading">{t("common.loading")}</div>;
}

function AuthGate() {
  const authQuery = useAuthStatus();
  const location = useLocation();

  if (authQuery.isLoading) {
    return <AuthLoading />;
  }

  if (!authQuery.data?.authenticated) {
    return <Navigate replace state={{ from: location.pathname }} to="/login" />;
  }

  return <AppLayout />;
}

function LoginRoute() {
  const authQuery = useAuthStatus();

  if (authQuery.isLoading) {
    return <AuthLoading />;
  }

  if (authQuery.data?.authenticated) {
    return <Navigate replace to="/" />;
  }

  return <LoginPage />;
}

export const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginRoute />,
  },
  {
    path: "/",
    element: <AuthGate />,
    children: [
      { index: true, element: <OverviewPage /> },
      { path: "zashboard", element: <ZashboardPage /> },
      { path: "logs", element: <LogsPage /> },
      { path: "*", element: <Navigate replace to="/" /> },
    ],
  },
  {
    path: "*",
    element: <Navigate replace to="/" />,
  },
]);
