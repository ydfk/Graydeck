import { useEffect, useMemo, useState, type SyntheticEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiPost } from "@/api/client";
import { useSystemStatus } from "@/api/queries";
import { useI18n } from "@/i18n/I18nProvider";
import type { SystemStatus } from "@/types/api";
import { Panel } from "@/ui/components/Panel";

export function ZashboardPage() {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const statusQuery = useSystemStatus();
  const [isFullscreen, setIsFullscreen] = useState(false);

  const dashboardBackend = useMemo(() => {
    const port = window.location.port || (window.location.protocol === "https:" ? "443" : "80");
    return {
      disableTunMode: true,
      disableUpgradeCore: true,
      host: window.location.hostname,
      label: "Graydeck",
      password: "",
      port,
      protocol: window.location.protocol.replace(":", ""),
      secondaryPath: "/api/clash",
      uuid: "graydeck-embedded",
    };
  }, []);

  const dashboardUrl = "/zashboard-ui/";

  useEffect(() => {
    localStorage.setItem("setup/api-list", JSON.stringify([dashboardBackend]));
    localStorage.setItem("setup/active-uuid", dashboardBackend.uuid);
  }, [dashboardBackend]);

  const updateMutation = useMutation({
    mutationFn: () => apiPost<SystemStatus>("/api/zashboard/update"),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
    },
  });

  const status = statusQuery.data;
  function formatVersion(value: string, fallback: string) {
    if (!value) {
      return fallback;
    }

    if (value === "installed" || value === "unknown") {
      return t("core.unknown");
    }

    return value;
  }

  const versionStatus = useMemo(() => {
    if (!status) {
      return t("common.loading");
    }

    if (!status.zashboardReady) {
      return t("zashboard.notReady");
    }

    if (status.zashboardIsLatest) {
      return t("zashboard.isLatest");
    }

    if (status.zashboardVersion === "") {
      return t("zashboard.ready");
    }

    return t("zashboard.hasUpdate");
  }, [status, t]);

  function getZashboardBadgeClass() {
    if (!status?.zashboardReady) {
      return "badge danger";
    }

    if (status.zashboardIsLatest) {
      return "badge active";
    }

    return "badge warning";
  }

  function handleFrameLoad(event: SyntheticEvent<HTMLIFrameElement>) {
    const doc = event.currentTarget.contentDocument;
    if (!doc || doc.getElementById("graydeck-zashboard-style")) {
      return;
    }

    const style = doc.createElement("style");
    style.id = "graydeck-zashboard-style";
    style.textContent = `
      a[href*="settings"],
      button[aria-label*="Settings"],
      button[aria-label*="settings"],
      button[aria-label*="设置"],
      [data-page="settings"],
      [data-route="settings"] {
        display: none !important;
      }
    `;
    doc.head?.appendChild(style);
  }

  const title = status ? (
    <span className="panel-title-inline">
      <span>{t("zashboard.title")}</span>
      <span className="panel-title-meta">
        {t("zashboard.currentVersion")} {formatVersion(status.zashboardVersion, t("core.notReady"))}
      </span>
      <span className="panel-title-meta">
        {t("zashboard.latestVersion")} {formatVersion(status.zashboardLatestVersion, t("core.unknown"))}
      </span>
      <span className={getZashboardBadgeClass()}>{versionStatus}</span>
    </span>
  ) : (
    t("zashboard.title")
  );

  return (
    <div className="page-grid">
      <Panel
        actions={
          <>
            <button
              className="secondary-pill"
              disabled={updateMutation.isPending || (status?.zashboardIsLatest ?? false)}
              onClick={() => updateMutation.mutate()}
              type="button"
            >
              {updateMutation.isPending ? t("common.loading") : t("zashboard.updateNow")}
            </button>
            <button className="primary-pill" onClick={() => setIsFullscreen(true)} type="button">
              {t("common.fullscreen")}
            </button>
          </>
        }
        title={title}
      >
        {status ? (
          <div className="page-grid compact-grid">
            {!status.zashboardReady && status.zashboardError ? (
              <div className="status-notice status-notice-error">{status.zashboardError}</div>
            ) : null}
            <div className="zashboard-frame-shell">
              <iframe className="zashboard-iframe" onLoad={handleFrameLoad} src={dashboardUrl} title="Zashboard" />
            </div>
          </div>
        ) : (
          <p className="body-copy">{t("common.loading")}</p>
        )}

        {isFullscreen ? (
          <div className="fullscreen-backdrop" onClick={() => setIsFullscreen(false)} role="presentation">
            <div aria-modal="true" className="fullscreen-card" onClick={(event) => event.stopPropagation()} role="dialog">
              <div className="fullscreen-header">
                <span className="panel-title-inline">
                  <span>{t("zashboard.title")}</span>
                  <span className={getZashboardBadgeClass()}>{versionStatus}</span>
                </span>
                <button className="secondary-pill" onClick={() => setIsFullscreen(false)} type="button">
                  {t("common.close")}
                </button>
              </div>
              <iframe
                className="zashboard-iframe zashboard-iframe-fullscreen"
                onLoad={handleFrameLoad}
                src={dashboardUrl}
                title="Zashboard fullscreen"
              />
            </div>
          </div>
        ) : null}
      </Panel>
    </div>
  );
}
