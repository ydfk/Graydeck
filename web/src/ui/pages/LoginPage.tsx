import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useLocation, useNavigate } from "react-router-dom";
import { ApiError, apiPost } from "@/api/client";
import { useAuthStatus } from "@/api/queries";
import { useI18n } from "@/i18n/I18nProvider";
import type { AuthStatus } from "@/types/api";

export function LoginPage() {
  const { locale, localeLabels, setLocale, t } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const authQuery = useAuthStatus();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const loginMutation = useMutation({
    mutationFn: (payload: { username: string; password: string }) => apiPost<AuthStatus>("/api/auth/login", payload),
    onSuccess: async () => {
      setErrorMessage("");
      await queryClient.invalidateQueries({ queryKey: ["auth-status"] });
      const redirectPath =
        typeof location.state === "object" &&
        location.state !== null &&
        "from" in location.state &&
        typeof location.state.from === "string"
          ? location.state.from
          : "/";
      navigate(redirectPath, { replace: true });
    },
    onError: (error) => {
      if (error instanceof ApiError) {
        setErrorMessage(error.message);
        return;
      }

      setErrorMessage(t("auth.loginFailed"));
    },
  });

  useEffect(() => {
    if (authQuery.data?.authenticated) {
      navigate("/", { replace: true });
    }
  }, [authQuery.data?.authenticated, navigate]);

  return (
    <div className="login-shell">
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
        </div>
      </header>

      <main className="login-main">
        <section className="login-card">
          <div className="login-copy">
            <h2 className="panel-title">{t("auth.title")}</h2>
          </div>

          {errorMessage ? <div className="status-notice status-notice-error">{errorMessage}</div> : null}

          <div className="login-form">
            <label className="login-field">
              <span className="summary-label">{t("auth.username")}</span>
              <input
                autoComplete="username"
                className="table-input"
                onChange={(event) => setUsername(event.target.value)}
                value={username}
              />
            </label>
            <label className="login-field">
              <span className="summary-label">{t("auth.password")}</span>
              <input
                autoComplete="current-password"
                className="table-input"
                onChange={(event) => setPassword(event.target.value)}
                type="password"
                value={password}
              />
            </label>
          </div>

          <button
            className="primary-pill login-submit"
            disabled={loginMutation.isPending || authQuery.isLoading}
            onClick={() => loginMutation.mutate({ username, password })}
            type="button"
          >
            {loginMutation.isPending ? t("common.loading") : t("auth.submit")}
          </button>
        </section>
      </main>
    </div>
  );
}
