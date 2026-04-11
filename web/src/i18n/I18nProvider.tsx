import { startTransition, createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react'
import { defaultLocale, localeLabels, messages, type Locale, type MessageKey } from './messages'

type TranslateValues = Record<string, string | number>

type I18nContextValue = {
  locale: Locale
  localeLabels: Record<Locale, string>
  setLocale: (locale: Locale) => void
  t: (key: MessageKey, values?: TranslateValues) => string
}

const STORAGE_KEY = 'mihomo-manager.locale'

const I18nContext = createContext<I18nContextValue | null>(null)

function interpolate(template: string, values?: TranslateValues) {
  if (!values) {
    return template
  }

  return template.replace(/\{(\w+)\}/g, (_, key) => String(values[key] ?? ''))
}

function getInitialLocale(): Locale {
  if (typeof window === 'undefined') {
    return defaultLocale
  }

  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored === 'zh-CN' || stored === 'en') {
    return stored
  }

  return navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en'
}

export function I18nProvider({ children }: PropsWithChildren) {
  const [locale, setLocaleState] = useState<Locale>(getInitialLocale)

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, locale)
    document.documentElement.lang = locale
  }, [locale])

  const value = useMemo<I18nContextValue>(() => {
    return {
      locale,
      localeLabels,
      setLocale: (nextLocale) => {
        startTransition(() => {
          setLocaleState(nextLocale)
        })
      },
      t: (key, values) => interpolate(messages[locale][key], values),
    }
  }, [locale])

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  const context = useContext(I18nContext)

  if (!context) {
    throw new Error('useI18n must be used inside I18nProvider')
  }

  return context
}
