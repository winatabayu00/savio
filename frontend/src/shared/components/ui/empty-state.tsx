import type { ReactNode } from 'react';

export function EmptyState({ title, description, action }: { title: string; description?: string; action?: ReactNode }) {
  return (
    <div className="card text-center border-dashed">
      <div className="card-body py-5">
        <p className="fs-14 fw-medium text-muted">{title}</p>
        {description ? <p className="fs-13 text-muted mt-1 mb-0">{description}</p> : null}
        {action ? <div className="mt-3">{action}</div> : null}
      </div>
    </div>
  );
}