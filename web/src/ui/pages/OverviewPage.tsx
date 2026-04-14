import { Fragment, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiGet, apiPost, apiUpload } from "@/api/client";
import { useSubscriptions, useSystemStatus } from "@/api/queries";
import { useI18n } from "@/i18n/I18nProvider";
import type { RuntimeConfigPreview, SubscriptionPreview, SystemStatus } from "@/types/api";
import { ArtifactUpdateDialog } from "@/ui/components/ArtifactUpdateDialog";
import { Panel } from "@/ui/components/Panel";
import { SubscriptionTable } from "@/ui/components/SubscriptionTable";

type CreatePayload = {
  name: string;
  url: string;
  syncInterval: string;
};

export function OverviewPage() {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const statusQuery = useSystemStatus();
  const subscriptionsQuery = useSubscriptions();
  const systemStatus = statusQuery.data;
  const subscriptions = subscriptionsQuery.data?.items ?? [];
  const [previewId, setPreviewId] = useState<string | null>(null);
  const [showRuntimeConfig, setShowRuntimeConfig] = useState(false);
  const [showCoreUpdateDialog, setShowCoreUpdateDialog] = useState(false);
  const [actionError, setActionError] = useState("");
  const [showCreateForm, setShowCreateForm] = useState(false);

  const previewQuery = useQuery({
    enabled: Boolean(previewId),
    queryKey: ["subscription-preview", previewId],
    queryFn: () => apiGet<SubscriptionPreview>(`/api/subscriptions/preview?id=${previewId}`),
  });

  const runtimeConfigQuery = useQuery({
    enabled: showRuntimeConfig && Boolean(systemStatus?.currentConfigName),
    queryKey: ["runtime-config-preview"],
    queryFn: () => apiGet<RuntimeConfigPreview>("/api/system/config/current"),
  });

  const refreshMutation = useMutation({
    mutationFn: () => apiPost<SystemStatus>("/api/system/refresh"),
    onSuccess: async () => {
      setActionError("");
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      if (previewId) {
        await queryClient.invalidateQueries({ queryKey: ["subscription-preview", previewId] });
      }
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const startRuntimeMutation = useMutation({
    mutationFn: () => apiPost<SystemStatus>("/api/system/start"),
    onSuccess: async () => {
      setActionError("");
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const restartRuntimeMutation = useMutation({
    mutationFn: () => apiPost<SystemStatus>("/api/system/restart"),
    onSuccess: async () => {
      setActionError("");
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const stopRuntimeMutation = useMutation({
    mutationFn: () => apiPost<SystemStatus>("/api/system/stop"),
    onSuccess: async () => {
      setActionError("");
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const updateCoreMutation = useMutation({
    mutationFn: (payload?: { source: "auto" | "url"; url?: string }) => apiPost<SystemStatus>("/api/system/core/update", payload),
    onSuccess: async () => {
      setActionError("");
      setShowCoreUpdateDialog(false);
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const uploadCoreMutation = useMutation({
    mutationFn: (file: File) => apiUpload<SystemStatus>("/api/system/core/upload", file),
    onSuccess: async () => {
      setActionError("");
      setShowCoreUpdateDialog(false);
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const createSubscriptionMutation = useMutation({
    mutationFn: (payload: CreatePayload) => apiPost("/api/subscriptions/create", payload),
    onSuccess: async () => {
      setActionError("");
      setShowCreateForm(false);
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const updateSubscriptionMutation = useMutation({
    mutationFn: (payload: CreatePayload & { id: string }) => apiPost("/api/subscriptions/update", payload),
    onSuccess: async (_data, variables) => {
      setActionError("");
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      if (previewId === variables.id) {
        await queryClient.invalidateQueries({ queryKey: ["subscription-preview", previewId] });
      }
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const syncSubscriptionMutation = useMutation({
    mutationFn: (id: string) => apiPost("/api/subscriptions/sync", { id }),
    onSuccess: async (_data, id) => {
      setActionError("");
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["logs"] });
      if (previewId === id) {
        await queryClient.invalidateQueries({ queryKey: ["subscription-preview", previewId] });
      }
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const activateSubscriptionMutation = useMutation({
    mutationFn: (id: string) => apiPost("/api/subscriptions/activate", { id }),
    onSuccess: async (_data, id) => {
      setActionError("");
      setPreviewId(id);
      await queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      await queryClient.invalidateQueries({ queryKey: ["system-status"] });
      await queryClient.invalidateQueries({ queryKey: ["subscription-preview", id] });
    },
    onError: (error) => setActionError(error instanceof Error ? error.message : t("status.error")),
  });

  const runtimeActionPending =
    startRuntimeMutation.isPending || restartRuntimeMutation.isPending || stopRuntimeMutation.isPending;

  const runtimePorts = useMemo(
    () =>
      systemStatus
        ? [
            { key: "mixed", label: t("config.mixedPort"), value: systemStatus.runtimeMixedPort },
            { key: "socks", label: t("config.socksPort"), value: systemStatus.runtimeSocksPort },
            { key: "redir", label: t("config.redirPort"), value: systemStatus.runtimeRedirPort },
            { key: "tproxy", label: t("config.tproxyPort"), value: systemStatus.runtimeTProxyPort },
          ]
        : [],
    [systemStatus, t],
  );

  const versionStatusLabel = useMemo(() => {
    if (!systemStatus) {
      return t("common.loading");
    }

    if (!systemStatus.coreExecutableReady) {
      return t("core.notReady");
    }

    return systemStatus.coreIsLatest ? t("core.isLatest") : t("core.hasUpdate");
  }, [systemStatus, t]);

  const coreActionLabel = systemStatus?.coreExecutableReady ? t("core.updateNow") : `${t("common.install")}${t("update.coreTitle")}`;
  const showCoreActionButton = !systemStatus?.coreExecutableReady || !systemStatus.coreIsLatest;

  function formatVersion(value: string, fallback: string) {
    if (!value) {
      return fallback;
    }

    if (value === "installed") {
      return t("common.installed");
    }

    if (value === "unknown") {
      return t("core.unknown");
    }

    return value;
  }

  function getRuntimeLabel(status: string) {
    const mapping: Record<string, string> = {
      running: t("status.running"),
      stopped: t("status.stopped"),
      error: t("status.error"),
      pending: t("status.pending"),
    };

    return mapping[status] ?? status;
  }

  function getRuntimeBadgeClass(status: string) {
    if (status === "running") {
      return "badge active";
    }

    if (status === "error") {
      return "badge danger";
    }

    if (status === "pending") {
      return "badge warning";
    }

    return "badge";
  }

  function getCoreVersionBadgeClass() {
    if (!systemStatus?.coreExecutableReady) {
      return "badge danger";
    }

    if (systemStatus.coreIsLatest) {
      return "badge active";
    }

    return "badge warning";
  }

  return (
    <div className="page-grid">
      {actionError ? <div className="status-notice status-notice-error">{actionError}</div> : null}

      <Panel
        actions={
          <button className="primary-pill" disabled={refreshMutation.isPending} onClick={() => refreshMutation.mutate()} type="button">
            {refreshMutation.isPending ? t("common.loading") : t("common.refresh")}
          </button>
        }
        title={t("system.status.title")}
      >
        {systemStatus ? (
          <div className="info-table-shell">
            <table className="info-table">
              <tbody>
                <tr>
                  <th>{t("system.runtimeStatus")}</th>
                  <td>
                    <div className="cell-stack">
                      <div className="table-actions">
                        <span className={getRuntimeBadgeClass(systemStatus.runtimeStatus)}>
                          {getRuntimeLabel(systemStatus.runtimeStatus)}
                        </span>
                        {systemStatus.runtimeStatus === "running" ? (
                          <>
                            <button
                              className="secondary-pill table-action-button"
                              disabled={runtimeActionPending}
                              onClick={() => restartRuntimeMutation.mutate()}
                              type="button"
                            >
                              {restartRuntimeMutation.isPending ? t("common.loading") : t("common.restart")}
                            </button>
                            <button
                              className="primary-pill table-action-button"
                              disabled={runtimeActionPending}
                              onClick={() => stopRuntimeMutation.mutate()}
                              type="button"
                            >
                              {stopRuntimeMutation.isPending ? t("common.loading") : t("common.stop")}
                            </button>
                          </>
                        ) : (
                          <button
                            className="primary-pill table-action-button"
                            disabled={runtimeActionPending}
                            onClick={() => startRuntimeMutation.mutate()}
                            type="button"
                          >
                            {startRuntimeMutation.isPending ? t("common.loading") : t("common.start")}
                          </button>
                        )}
                      </div>
                      {systemStatus.runtimeError ? <div className="table-secondary danger">{systemStatus.runtimeError}</div> : null}
                    </div>
                  </td>
                </tr>
                <tr>
                  <th>{t("system.currentConfig")}</th>
                  <td>
                    <div className="table-actions">
                      <span className="summary-value-inline">{systemStatus.currentConfigName || t("common.empty")}</span>
                      <button
                        className="secondary-pill table-action-button"
                        disabled={!systemStatus.currentConfigName}
                        onClick={() => setShowRuntimeConfig(true)}
                        type="button"
                      >
                        {t("system.viewRuntimeConfig")}
                      </button>
                    </div>
                  </td>
                </tr>
                <tr>
                  <th>{t("system.runtimeManaged")}</th>
                  <td>
                    <div className="table-actions">
                      {runtimePorts.map((item) => (
                        <Fragment key={item.key}>
                          <span className="summary-label">{item.label}</span>
                          <span className="summary-value-inline">{item.value || t("common.empty")}</span>
                        </Fragment>
                      ))}
                    </div>
                  </td>
                </tr>
                <tr>
                  <th>{t("core.groupTitle")}</th>
                  <td>
                    <div className="table-actions">
                      <span className="summary-label">{t("core.currentVersion")}</span>
                      <span className="summary-value-inline">{formatVersion(systemStatus.coreVersion, t("core.notReady"))}</span>
                      <span className="summary-label">{t("core.latestVersion")}</span>
                      <span className="summary-value-inline">{formatVersion(systemStatus.coreLatestVersion, t("core.unknown"))}</span>
                      <span className="summary-label">{t("core.versionStatus")}</span>
                      <span className={getCoreVersionBadgeClass()}>{versionStatusLabel}</span>
                      {showCoreActionButton ? (
                        <button
                          className="primary-pill table-action-button"
                          disabled={updateCoreMutation.isPending || uploadCoreMutation.isPending}
                          onClick={() => setShowCoreUpdateDialog(true)}
                          type="button"
                        >
                          {updateCoreMutation.isPending || uploadCoreMutation.isPending ? t("common.loading") : coreActionLabel}
                        </button>
                      ) : null}
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        ) : (
          <p className="body-copy">{t("common.loading")}</p>
        )}
      </Panel>

      <Panel
        actions={
          showCreateForm ? null : (
            <button className="primary-pill" onClick={() => setShowCreateForm(true)} type="button">
              {t("common.create")}
            </button>
          )
        }
        title={t("config.files.title")}
      >
        <SubscriptionTable
          creating={createSubscriptionMutation.isPending}
          onActivate={(id) => activateSubscriptionMutation.mutate(id)}
          onCreate={(payload) => createSubscriptionMutation.mutate(payload)}
          onHideCreateForm={() => setShowCreateForm(false)}
          onPreview={(id) => setPreviewId(id)}
          onSync={(id) => syncSubscriptionMutation.mutate(id)}
          onUpdate={(id, payload) => updateSubscriptionMutation.mutate({ id, ...payload })}
          previewingId={previewQuery.isFetching ? previewId : null}
          savingId={updateSubscriptionMutation.isPending ? updateSubscriptionMutation.variables?.id ?? null : null}
          showCreateForm={showCreateForm}
          subscriptions={subscriptions}
          syncingId={syncSubscriptionMutation.isPending ? syncSubscriptionMutation.variables ?? null : null}
          switchingId={activateSubscriptionMutation.isPending ? activateSubscriptionMutation.variables ?? null : null}
        />
      </Panel>

      {previewId && previewQuery.data ? (
        <div className="modal-backdrop" onClick={() => setPreviewId(null)} role="presentation">
          <div className="modal-card" onClick={(event) => event.stopPropagation()} role="dialog" aria-modal="true">
            <div className="yaml-preview-header">
              <div>
                <p className="mono-label">{t("common.preview")}</p>
                <div className="table-primary">{previewQuery.data.name}</div>
              </div>
              <button className="secondary-pill table-action-button" onClick={() => setPreviewId(null)} type="button">
                {t("common.close")}
              </button>
            </div>
            <pre className="code-block">{previewQuery.data.content}</pre>
          </div>
        </div>
      ) : null}

      {showRuntimeConfig ? (
        <div className="modal-backdrop" onClick={() => setShowRuntimeConfig(false)} role="presentation">
          <div className="modal-card" onClick={(event) => event.stopPropagation()} role="dialog" aria-modal="true">
            <div className="yaml-preview-header">
              <div>
                <p className="mono-label">{t("system.runtimeConfigTitle")}</p>
                <div className="table-primary">{runtimeConfigQuery.data?.name ?? t("common.loading")}</div>
                <div className="table-secondary">{runtimeConfigQuery.data?.path ?? ""}</div>
              </div>
              <button className="secondary-pill table-action-button" onClick={() => setShowRuntimeConfig(false)} type="button">
                {t("common.close")}
              </button>
            </div>
            {runtimeConfigQuery.isLoading ? <p className="body-copy">{t("common.loading")}</p> : null}
            {runtimeConfigQuery.isError ? (
              <div className="status-notice status-notice-error">
                {runtimeConfigQuery.error instanceof Error ? runtimeConfigQuery.error.message : t("status.error")}
              </div>
            ) : null}
            {runtimeConfigQuery.data ? <pre className="code-block">{runtimeConfigQuery.data.content}</pre> : null}
          </div>
        </div>
      ) : null}

      <ArtifactUpdateDialog
        errorMessage={actionError}
        fileAccept=".gz,.zip,.exe"
        onClose={() => setShowCoreUpdateDialog(false)}
        onSubmitAuto={() => updateCoreMutation.mutate({ source: "auto" })}
        onSubmitUpload={(file) => uploadCoreMutation.mutate(file)}
        onSubmitURL={(url) => updateCoreMutation.mutate({ source: "url", url })}
        open={showCoreUpdateDialog}
        pending={updateCoreMutation.isPending || uploadCoreMutation.isPending}
        title={t("update.coreTitle")}
      />
    </div>
  );
}
