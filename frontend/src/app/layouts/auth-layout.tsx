import { Outlet } from 'react-router-dom';

export function AuthLayout() {
  return (
    <main className="auth-minimal-wrapper">
      <div className="auth-minimal-inner">
        <div className="minimal-card-wrapper">
          <div className="card mb-4 mt-5 mx-4 mx-sm-0 position-relative">
            <div className="wd-50 bg-primary p-2 rounded-circle shadow-lg position-absolute translate-middle top-0 start-50 d-flex align-items-center justify-content-center fw-bold text-white">
              S
            </div>
            <div className="card-body p-sm-5">
              <div className="mb-4 text-center">
                <div className="fs-20 fw-bolder text-dark">Savio</div>
                <p className="fs-12 fw-medium text-muted mb-0">
                  Personal cashflow intelligence
                </p>
              </div>
              <Outlet />
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}