import { useEffect, useState } from "react";
import { useI18n } from "@/i18n/I18nProvider";

type UpdateSource = "auto" | "url" | "upload";

type ArtifactUpdateDialogProps = {
  open: boolean;
  pending: boolean;
  title: string;
  fileAccept: string;
  errorMessage?: string;
  onClose: () => void;
  onSubmitAuto: () => void;
  onSubmitURL: (url: string) => void;
  onSubmitUpload: (file: File) => void;
};

export function ArtifactUpdateDialog({
  open,
  pending,
  title,
  fileAccept,
  errorMessage,
  onClose,
  onSubmitAuto,
  onSubmitURL,
  onSubmitUpload,
}: ArtifactUpdateDialogProps) {
  const { t } = useI18n();
  const [source, setSource] = useState<UpdateSource>("auto");
  const [url, setURL] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [localError, setLocalError] = useState("");

  useEffect(() => {
    if (!open) {
      setSource("auto");
      setURL("");
      setFile(null);
      setLocalError("");
    }
  }, [open]);

  if (!open) {
    return null;
  }

  function handleSubmit() {
    setLocalError("");

    if (source === "auto") {
      onSubmitAuto();
      return;
    }

    if (source === "url") {
      const trimmed = url.trim();
      if (!trimmed) {
        setLocalError(t("update.urlEmpty"));
        return;
      }
      onSubmitURL(trimmed);
      return;
    }

    if (!file) {
      setLocalError(t("update.noFile"));
      return;
    }

    onSubmitUpload(file);
  }

  const sourceLabel: Record<UpdateSource, string> = {
    auto: t("update.sourceAuto"),
    url: t("update.sourceUrl"),
    upload: t("update.sourceUpload"),
  };

  const sourceHint: Record<UpdateSource, string> = {
    auto: t("update.sourceAutoHint"),
    url: t("update.sourceUrlHint"),
    upload: t("update.sourceUploadHint"),
  };

  return (
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div
        className="modal-card update-dialog-card"
        onClick={(event) => event.stopPropagation()}
        role="dialog"
        aria-modal="true"
      >
        <div className="update-dialog-header">
          <div>
            <h2 className="update-dialog-title">{t("update.dialogTitle", { name: title })}</h2>
            <p className="update-dialog-desc">{sourceHint[source]}</p>
          </div>
          <button className="secondary-pill" onClick={onClose} type="button">
            {t("common.close")}
          </button>
        </div>

        <div className="segmented-control update-source-segment">
          {(Object.keys(sourceLabel) as UpdateSource[]).map((key) => (
            <button
              className={source === key ? "segmented-button active" : "segmented-button"}
              key={key}
              onClick={() => setSource(key)}
              type="button"
            >
              {sourceLabel[key]}
            </button>
          ))}
        </div>

        <div className="update-dialog-body">
          {source === "auto" ? (
            <p className="body-copy update-dialog-placeholder">{t("update.autoDescription")}</p>
          ) : null}

          {source === "url" ? (
            <label className="login-field">
              <span className="summary-label">{t("update.urlLabel")}</span>
              <input
                autoFocus
                className="table-input"
                onChange={(event) => setURL(event.target.value)}
                placeholder={t("update.urlPlaceholder")}
                value={url}
              />
            </label>
          ) : null}

          {source === "upload" ? (
            <label className="login-field">
              <span className="summary-label">{t("update.fileLabel")}</span>
              <input
                accept={fileAccept}
                className="table-input"
                onChange={(event) => setFile(event.target.files?.[0] ?? null)}
                type="file"
              />
            </label>
          ) : null}
        </div>

        {localError || errorMessage ? (
          <div className="status-notice status-notice-error">{localError || errorMessage}</div>
        ) : null}

        <button className="primary-pill login-submit" disabled={pending} onClick={handleSubmit} type="button">
          {pending ? t("common.loading") : t("update.submit")}
        </button>
      </div>
    </div>
  );
}
