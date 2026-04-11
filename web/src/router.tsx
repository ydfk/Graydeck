import { createBrowserRouter } from 'react-router-dom'
import { AppLayout } from '@/ui/AppLayout'
import { ConfigPage } from '@/ui/pages/ConfigPage'
import { OverviewPage } from '@/ui/pages/OverviewPage'
import { SystemPage } from '@/ui/pages/SystemPage'
import { UpdatesPage } from '@/ui/pages/UpdatesPage'
import { ZashboardPage } from '@/ui/pages/ZashboardPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <OverviewPage /> },
      { path: 'config', element: <ConfigPage /> },
      { path: 'updates', element: <UpdatesPage /> },
      { path: 'system', element: <SystemPage /> },
      { path: 'zashboard', element: <ZashboardPage /> },
    ],
  },
])
