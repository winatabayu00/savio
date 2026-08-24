import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ProtectedRoute } from './protected-route';
import { GuestRoute } from './guest-route';
import { AppLayout } from '@/app/layouts/app-layout';
import { AuthLayout } from '@/app/layouts/auth-layout';
import { LoginPage } from '@/features/auth/pages/login-page';
import { RegisterPage } from '@/features/auth/pages/register-page';
import { DashboardPage } from '@/features/dashboard/pages/dashboard-page';
import { AccountsPage } from '@/features/accounts/pages/accounts-page';
import { CategoriesPage } from '@/features/categories/pages/categories-page';
import { TransactionsPage } from '@/features/transactions/pages/transactions-page';
import { TransfersPage } from '@/features/transfers/pages/transfers-page';
import { RecurringPage } from '@/features/recurring/pages/recurring-page';
import { AnalyticsPage } from '@/features/analytics/pages/analytics-page';
import { BudgetsPage } from '@/features/budgets/pages/budgets-page';
import { GoalsPage } from '@/features/goals/pages/goals-page';
import { ForecastPage } from '@/features/forecast/pages/forecast-page';
import { ScenariosPage } from '@/features/scenarios/pages/scenarios-page';

export const router = createBrowserRouter([
  {
    element: <AuthLayout />,
    children: [
      {
        path: '/login',
        element: (
          <GuestRoute>
            <LoginPage />
          </GuestRoute>
        ),
      },
      {
        path: '/register',
        element: (
          <GuestRoute>
            <RegisterPage />
          </GuestRoute>
        ),
      },
    ],
  },
  {
    element: (
      <ProtectedRoute>
        <AppLayout />
      </ProtectedRoute>
    ),
    children: [
      { path: '/dashboard', element: <DashboardPage /> },
      { path: '/accounts', element: <AccountsPage /> },
      { path: '/categories', element: <CategoriesPage /> },
      { path: '/transactions', element: <TransactionsPage /> },
      { path: '/transfers', element: <TransfersPage /> },
      { path: '/recurring', element: <RecurringPage /> },
      { path: '/analytics', element: <AnalyticsPage /> },
      { path: '/budgets', element: <BudgetsPage /> },
      { path: '/goals', element: <GoalsPage /> },
      { path: '/forecast', element: <ForecastPage /> },
      { path: '/scenarios', element: <ScenariosPage /> },
    ],
  },
  { path: '/', element: <Navigate to="/dashboard" replace /> },
  { path: '*', element: <Navigate to="/dashboard" replace /> },
]);