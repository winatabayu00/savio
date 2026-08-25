import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { Modal } from '@/shared/components/ui/modal';
import { listAuditLogs, type AuditLog } from '@/features/audit/api/audit.api';

const actorStyle = {
  USER: 'bg-soft-secondary text-secondary',
  AI: 'bg-soft-primary text-primary',
  SYSTEM: 'bg-soft-warning text-warning',
};

function formatAction(action: string) {
  return action.replace('.', ' ');
}

function Snapshot({ data }: { data: Record<string, unknown> | null }) {
  if (!data || Object.keys(data).length === 0) return <p className="fs-13 text-muted">No data recorded.</p>;
  return <pre className="overflow-auto rounded bg-dark p-3 fs-12 text-light">{JSON.stringify(data, null, 2)}</pre>;
}

export function AuditLogsPage() {
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<AuditLog>();
  const logs = useQuery({ queryKey: ['audit-logs', page], queryFn: () => listAuditLogs(page) });
  const totalPages = logs.data ? Math.max(1, Math.ceil(logs.data.meta.total / logs.data.meta.limit)) : 1;

  return (
    <div>
      <div className="border-bottom pb-4">
        <p className="font-monospace fs-12 fw-semibold text-uppercase text-primary">Workspace control</p>
        <h1 className="mt-2 fs-20 fw-bolder text-dark">Audit trail</h1>
        <p className="mt-1 fs-13 text-muted mb-0">Immutable record of important workspace changes. Shown only to owners.</p>
      </div>

      {logs.isLoading ? <p className="mt-4 fs-13 text-muted">Loading audit trail...</p> : null}
      {logs.isError ? <div className="mt-4 alert alert-danger">Could not load the audit trail. Your financial data is unchanged.</div> : null}
      {logs.data?.data.length === 0 ? <div className="mt-4"><EmptyState title="No audit entries yet" description="Important workspace changes will appear here." /></div> : null}

      {logs.data?.data.length ? (
        <div className="card mt-4">
          <div className="table-responsive">
          <table className="table table-hover align-middle">
            <thead className="table-light fs-12 text-uppercase text-muted">
              <tr><th className="px-4 py-3 fw-medium">When</th><th className="px-4 py-3 fw-medium">Actor</th><th className="px-4 py-3 fw-medium">Change</th><th className="d-none d-md-table-cell px-4 py-3 fw-medium">Reason</th></tr>
            </thead>
            <tbody>
              {logs.data.data.map((log) => (
                <tr key={log.id} className="border-bottom" onClick={() => setSelected(log)}>
                  <td className="text-nowrap px-4 py-3 text-muted">{new Date(log.occurred_at).toLocaleString()}</td>
                  <td className="px-4 py-3"><div className="fw-medium text-dark">{log.actor_name ?? log.actor_type}</div><span className={`badge ${actorStyle[log.actor_type]}`}>{log.actor_type}</span></td>
                  <td className="px-4 py-3"><div className="fw-medium text-capitalize text-dark">{formatAction(log.action)}</div><div className="fs-12 text-muted">{log.resource_type}</div></td>
                  <td className="d-none d-md-table-cell px-4 py-3 text-muted text-truncate">{log.reason ?? '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
          </div>
        </div>
      ) : null}

      {logs.data && totalPages > 1 ? <div className="mt-4 d-flex align-items-center gap-2"><Button variant="secondary" disabled={page === 1} onClick={() => setPage((value) => value - 1)}>Previous</Button><span className="fs-13 text-secondary">Page {page} of {totalPages}</span><Button variant="secondary" disabled={page === totalPages} onClick={() => setPage((value) => value + 1)}>Next</Button></div> : null}

      <Modal open={Boolean(selected)} onClose={() => setSelected(undefined)} title="Audit entry">
        {selected ? <div className="d-flex flex-column gap-4"><div className="row g-3 fs-13"><div className="col-6"><p className="text-muted">Action</p><p className="mt-1 fw-medium text-capitalize">{formatAction(selected.action)}</p></div><div className="col-6"><p className="text-muted">Actor</p><p className="mt-1 fw-medium">{selected.actor_name ?? selected.actor_type}</p></div>{selected.reason ? <div className="col-12"><p className="text-muted">Reason</p><p className="mt-1">{selected.reason}</p></div> : null}</div><div><h2 className="mb-3 fs-15 fw-semibold text-dark">Before</h2><Snapshot data={selected.before_data} /></div><div><h2 className="mb-3 fs-15 fw-semibold text-dark">After</h2><Snapshot data={selected.after_data} /></div></div> : null}
      </Modal>
    </div>
  );
}
