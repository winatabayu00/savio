import { Fragment, type ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';
import {
  FiBarChart,
  FiCreditCard,
  FiGrid,
  FiLayout,
  FiMessageSquare,
  FiRepeat,
  FiSettings,
  FiShield,
  FiTag,
  FiTarget,
  FiTrendingUp,
  FiUser,
  FiZap,
} from 'react-icons/fi';
import { useAuth } from '@/app/providers/auth-provider';

interface NavItem {
  to: string;
  label: string;
  icon: ReactNode;
}

interface NavGroup {
  label: string;
  items: NavItem[];
}

const GROUPS: NavGroup[] = [
  {
    label: 'Ringkasan',
    items: [{ to: '/dashboard', label: 'Dashboard', icon: <FiLayout /> }],
  },
  {
    label: 'Keuangan',
    items: [
      { to: '/accounts', label: 'Akun', icon: <FiCreditCard /> },
      { to: '/transactions', label: 'Transaksi', icon: <FiCreditCard /> },
      { to: '/transfers', label: 'Transfer', icon: <FiRepeat /> },
      { to: '/recurring', label: 'Berulang', icon: <FiRepeat /> },
      { to: '/categories', label: 'Kategori', icon: <FiTag /> },
    ],
  },
  {
    label: 'Analisis & Perencanaan',
    items: [
      { to: '/analytics', label: 'Analitik', icon: <FiBarChart /> },
      { to: '/budgets', label: 'Anggaran', icon: <FiTarget /> },
      { to: '/goals', label: 'Target', icon: <FiTarget /> },
    ],
  },
  {
    label: 'Proyeksi',
    items: [
      { to: '/forecast', label: 'Prediksi', icon: <FiTrendingUp /> },
      { to: '/scenarios', label: 'Skenario', icon: <FiGrid /> },
    ],
  },
  {
    label: 'AI',
    items: [
      { to: '/insights', label: 'Wawasan', icon: <FiZap /> },
      { to: '/copilot', label: 'Savio Copilot', icon: <FiMessageSquare /> },
    ],
  },
  {
    label: 'Pengaturan',
    items: [{ to: '/settings', label: 'Pengaturan', icon: <FiSettings /> }],
  },
];

interface NavigationMenuProps {
  navigationOpen: boolean;
  setNavigationOpen: (open: boolean) => void;
}

export function NavigationMenu({ navigationOpen, setNavigationOpen }: NavigationMenuProps) {
  const { auth } = useAuth();
  const { pathname } = useLocation();

  const groups =
    auth?.role === 'OWNER'
      ? [
          ...GROUPS,
          {
            label: 'Kontrol',
            items: [{ to: '/audit-logs', label: 'Audit trail', icon: <FiShield /> }],
          },
        ]
      : GROUPS;

  return (
    <nav className={`nxl-navigation ${navigationOpen ? 'mob-navigation-active' : ''}`}>
      <div className="navbar-wrapper">
        <div className="m-header">
          <Link to="/dashboard" className="b-brand">
            <span className="logo logo-lg fs-5 fw-bold text-dark">Savio</span>
            <span className="logo logo-sm fs-6 fw-bold text-dark">S</span>
          </Link>
        </div>
        <div className="navbar-content">
          <ul className="nxl-navbar">
            {groups.map((group) => (
              <Fragment key={group.label}>
                <li className="nxl-item nxl-caption">
                  <label>{group.label}</label>
                </li>
                {group.items.map((item) => (
                  <li
                    key={item.to}
                    className={`nxl-item ${pathname === item.to ? 'active' : ''}`}
                  >
                    <Link
                      to={item.to}
                      className="nxl-link"
                      onClick={() => setNavigationOpen(false)}
                    >
                      <span className="nxl-micon">
                        <i>{item.icon}</i>
                      </span>
                      <span className="nxl-mtext">{item.label}</span>
                    </Link>
                  </li>
                ))}
              </Fragment>
            ))}
          </ul>
          <div className="card">
            <div className="card-body text-center">
              <i className="fs-4 text-dark">
                <FiUser />
              </i>
              <h6 className="mt-3 fs-13 fw-bolder">{auth?.user.name}</h6>
              <p className="fs-11 my-2 text-muted text-capitalize">{auth?.role}</p>
            </div>
          </div>
        </div>
      </div>
      <div
        onClick={() => setNavigationOpen(false)}
        className={navigationOpen ? 'nxl-menu-overlay' : ''}
      ></div>
    </nav>
  );
}
