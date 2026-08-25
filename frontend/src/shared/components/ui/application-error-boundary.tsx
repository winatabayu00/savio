import { Component, type ReactNode } from 'react';

interface Props { children: ReactNode }
interface State { failed: boolean }

export class ApplicationErrorBoundary extends Component<Props, State> {
  state: State = { failed: false };

  static getDerivedStateFromError(): State {
    return { failed: true };
  }

  render() {
    if (this.state.failed) {
      return (
        <div className="auth-minimal-wrapper">
            <div className="auth-minimal-inner">
              <div className="minimal-card-wrapper">
                <div className="card mb-4 mt-5 mx-4 mx-sm-0">
                  <div className="card-body p-sm-5 text-center">
                    <p className="fs-13 text-uppercase text-primary fw-bold mb-2">Savio</p>
                    <h2 className="fs-20 fw-bolder mb-2">Something went wrong</h2>
                    <p className="fs-13 text-muted mb-3">Your financial data is safe. Reload this page to continue.</p>
                    <button type="button" className="btn btn-primary" onClick={() => window.location.reload()}>
                      Reload page
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
      );
    }
    return this.props.children;
  }
}
