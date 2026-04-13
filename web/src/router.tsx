import { createBrowserRouter } from 'react-router-dom'
import { AppLayout } from '@/ui/AppLayout'
import { LogsPage } from '@/ui/pages/LogsPage'
import { OverviewPage } from '@/ui/pages/OverviewPage'
import { ZashboardPage } from '@/ui/pages/ZashboardPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <OverviewPage /> },
      { path: 'zashboard', element: <ZashboardPage /> },
      { path: 'logs', element: <LogsPage /> },
    ],
  },
])
