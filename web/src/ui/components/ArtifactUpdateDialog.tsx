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
      onSubmitURL(url.trim());
      return;
    }

    if (!file) {
      setLocalError(t("update.noFile"));
      return;
    }

    onSubmitUpload(file);
  }

  return (
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div className="modal-card" onClick={(event) => event.stopPropagation()} role="dialog" aria-modal="true">
        <div className="yaml-preview-header">
          <div>
            <p className="mono-label">{title}</p>
            <div className="table-primary">{t("update.submit")}</div>
          </div>
          <button className="secondary-pill table-action-button" onClick={onClose} type="button">
            {t("common.close")}
          </button>
        </div>

        <div className="update-source-row">
          <button
            className={source === "auto" ? "secondary-pill active-source" : "secondary-pill"}
            onClick={() => setSource("auto")}
            type="button"
          >
            {t("update.sourceAuto")}
          </button>
          <button
            className={source === "url" ? "secondary-pill active-source" : "secondary-pill"}
            onClick={() => setSource("url")}
            type="button"
          >
            {t("update.sourceUrl")}
          </button>
          <button
            className={source === "upload" ? "secondary-pill active-source" : "secondary-pill"}
            onClick={() => setSource("upload")}
            type="button"
          >
            {t("update.sourceUpload")}
          </button>
        </div>

        {source === "url" ? (
          <label className="login-field">
            <span className="summary-label">{t("update.urlLabel")}</span>
            <input className="table-input" onChange={(event) => setURL(event.target.value)} value={url} />
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

        {localError || errorMessage ? <div className="status-notice status-notice-error">{localError || errorMessage}</div> : null}

        <button className="primary-pill login-submit" disabled={pending} onClick={handleSubmit} type="button">
          {pending ? t("common.loading") : t("update.submit")}
        </button>
      </div>
    </div>
  );
}
