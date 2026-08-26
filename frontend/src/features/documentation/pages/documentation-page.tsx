import { Link } from 'react-router-dom';

const capabilities = [
  ['Track', 'Create accounts, income, expenses, transfers, and recurring plans.'],
  ['Understand', 'Review cashflow, category spending, budgets, and financial goals.'],
  ['Forecast', 'See a deterministic 90-day cashflow forecast from your financial records.'],
  ['Simulate', 'Compare non-destructive scenarios before changing real financial data.'],
  ['Explain', 'Use Savio Copilot for grounded explanations of deterministic results.'],
];

export function DocumentationPage() {
  return (
    <main className="container py-5" style={{ maxWidth: 960 }}>
      <header className="d-flex flex-wrap align-items-center justify-content-between gap-3 border-bottom pb-4 mb-5">
        <Link to="/" className="text-decoration-none text-dark fw-bolder fs-20">
          Savio
        </Link>
        <div className="d-flex gap-2">
          <a className="btn btn-outline-primary" href="/docs">API docs</a>
          <Link className="btn btn-primary" to="/login">Open Savio</Link>
        </div>
      </header>

      <section className="mb-5">
        <p className="text-primary text-uppercase fw-bold fs-11 mb-2">Documentation</p>
        <h1 className="display-5 fw-bolder mb-3">Your cashflow, made understandable.</h1>
        <p className="fs-16 text-muted mb-0" style={{ maxWidth: 720 }}>
          Savio keeps actual records, projections, scenarios, and AI explanations distinct so you can make informed decisions.
        </p>
      </section>

      <section className="mb-5" aria-labelledby="getting-started">
        <h2 id="getting-started" className="fs-20 fw-bolder mb-3">Getting started</h2>
        <ol className="text-muted ps-3 mb-0 lh-lg">
          <li>Create an account. Savio creates your personal workspace.</li>
          <li>Add an account, then record income and expenses.</li>
          <li>Set category budgets or goals, then review Dashboard and Analytics.</li>
          <li>Calculate a forecast. Use Scenario Simulator before committing to a decision.</li>
        </ol>
      </section>

      <section className="mb-5" aria-labelledby="features">
        <h2 id="features" className="fs-20 fw-bolder mb-3">How Savio works</h2>
        <div className="row g-3">
          {capabilities.map(([title, description]) => (
            <article key={title} className="col-md-6">
              <div className="border rounded p-4 h-100">
                <h3 className="fs-16 fw-bolder mb-2">{title}</h3>
                <p className="text-muted mb-0">{description}</p>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="mb-5" aria-labelledby="financial-rules">
        <h2 id="financial-rules" className="fs-20 fw-bolder mb-3">Financial rules</h2>
        <div className="border-start border-primary border-3 ps-3 text-muted">
          <p className="mb-2"><strong className="text-dark">Actual:</strong> Posted income and expenses affect account balances and analytics.</p>
          <p className="mb-2"><strong className="text-dark">Projected:</strong> Forecasts are deterministic estimates, not actual balances.</p>
          <p className="mb-2"><strong className="text-dark">Scenario:</strong> Simulations never change your accounts or transaction history.</p>
          <p className="mb-0"><strong className="text-dark">AI-generated:</strong> AI explains data. It does not calculate or change financial records.</p>
        </div>
      </section>

      <section id="api-docs" className="border rounded p-4 bg-light" aria-labelledby="api-reference">
        <p className="text-primary text-uppercase fw-bold fs-11 mb-2">API reference</p>
        <h2 id="api-reference" className="fs-20 fw-bolder mb-2">REST API</h2>
        <p className="text-muted mb-3">
          Savio API endpoints use the <code>/api/v1</code> prefix. Authentication uses secure cookies; state-changing requests require CSRF protection.
        </p>
        <div className="row g-2 fs-13">
          <div className="col-sm-6"><code>/auth</code> Registration, sessions, current user</div>
          <div className="col-sm-6"><code>/accounts</code> Accounts and balances</div>
          <div className="col-sm-6"><code>/transactions</code> Income, expenses, adjustments</div>
          <div className="col-sm-6"><code>/analytics</code> Cashflow and spending summaries</div>
          <div className="col-sm-6"><code>/forecast</code> Deterministic forecasts</div>
          <div className="col-sm-6"><code>/scenarios</code> Non-destructive simulations</div>
        </div>
      </section>
    </main>
  );
}
