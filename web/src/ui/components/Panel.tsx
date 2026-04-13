import type { PropsWithChildren, ReactNode } from 'react'

type PanelProps = PropsWithChildren<{
  eyebrow?: string
  title: ReactNode
  description?: string
  actions?: ReactNode
}>

export function Panel({ eyebrow, title, description, actions, children }: PanelProps) {
  return (
    <section className="panel">
      <header className="panel-header">
        <div>
          {eyebrow ? <p className="mono-label">{eyebrow}</p> : null}
          <h2 className="panel-title">{title}</h2>
          {description ? <p className="panel-description">{description}</p> : null}
        </div>
        {actions ? <div className="panel-actions">{actions}</div> : null}
      </header>
      <div className="panel-body">{children}</div>
    </section>
  )
}
