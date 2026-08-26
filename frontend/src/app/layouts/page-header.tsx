import type { ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';

export function PageHeader({ children }: { children?: ReactNode }) {
  const { pathname } = useLocation();
  const segments = pathname.split('/').filter(Boolean);
  const routeLabels: Record<string, string> = {
    accounts: 'Akun',
    analytics: 'Analitik',
    'audit-logs': 'Audit trail',
    budgets: 'Anggaran',
    categories: 'Kategori',
    copilot: 'Savio Copilot',
    dashboard: 'Dashboard',
    forecast: 'Cashflow Forecast',
    goals: 'Target',
    insights: 'Wawasan AI',
    recurring: 'Transaksi berulang',
    scenarios: 'Scenario Simulator',
    settings: 'Pengaturan',
    transactions: 'Transaksi',
    transfers: 'Transfer',
  };
  const folder = routeLabels[segments[0]] ?? 'Dashboard';
  const file = segments[1] ? (routeLabels[segments[1]] ?? segments[1]) : folder;

  return (
    <div className="page-header">
      <div className="page-header-left d-flex align-items-center">
        <div className="page-header-title">
          <h5 className="m-b-10 text-capitalize">{folder}</h5>
        </div>
        <ul className="breadcrumb">
          <li className="breadcrumb-item">
            <Link to="/dashboard">Beranda</Link>
          </li>
          <li className="breadcrumb-item text-capitalize">{file}</li>
        </ul>
      </div>
      <div className="page-header-right ms-auto">
        <div className="page-header-right-items">{children}</div>
      </div>
    </div>
  );
}
