import { useI18n } from "@/i18n/I18nProvider";

export function BrandTitle() {
  const { t } = useI18n();

  return (
    <div className="brand-title">
      <img alt="" className="brand-logo" src="/graydeck-logo.svg" />
      <h1 className="app-title">{t("layout.title")}</h1>
    </div>
  );
}
