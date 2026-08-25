import type { ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';

export function PageHeader({ children }: { children?: ReactNode }) {
  const { pathname } = useLocation();
  const segments = pathname.split('/').filter(Boolean);
  const folder = segments.length > 0 ? segments[0] : 'Dashboard';
  const file = segments.length > 1 ? segments[1] : folder;

  return (
    <div className="page-header">
      <div className="page-header-left d-flex align-items-center">
        <div className="page-header-title">
          <h5 className="m-b-10 text-capitalize">{folder}</h5>
        </div>
        <ul className="breadcrumb">
          <li className="breadcrumb-item">
            <Link to="/dashboard">Home</Link>
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