import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/shared/components/ui/button';
import { EmptyState } from '@/shared/components/ui/empty-state';
import { Modal } from '@/shared/components/ui/modal';
import { listAuditLogs, type AuditLog } from '@/features/audit/api/audit.api';

const actorStyle = {
  USER: 'bg-slate-100 text-slate-700',
  AI: 'bg-violet-100 text-violet-700',
  SYSTEM: 'bg-amber-100 text-amber-700',
};

function formatAction(action: string) {
  return action.replace('.', ' ');
}

function Snapshot({ data }: { data: Record<string, unknown> | null }) {
  if (!data || Object.keys(data).length === 0) return <p className="text-sm text-gray-400">No data recorded.</p>;
  return <pre className="max-h-64 overflow-auto rounded-lg bg-slate-950 p-3 text-xs leading-5 text-slate-100">{JSON.stringify(data, null, 2)}</pre>;
}

export function AuditLogsPage() {
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<AuditLog>();
  const logs = useQuery({ queryKey: ['audit-logs', page], queryFn: () => listAuditLogs(page) });
  const totalPages = logs.data ? Math.max(1, Math.ceil(logs.data.meta.total / logs.data.meta.limit)) : 1;

  return (
    <div>
      <div className="border-b-2 border-slate-900 pb-5">
        <p className="font-mono text-xs font-semibold uppercase tracking-[0.18em] text-brand">Workspace control</p>
        <h1 className="mt-2 text-2xl font-semibold text-slate-950">Audit trail</h1>
        <p className="mt-1 text-sm text-gray-500">Immutable record of important workspace changes. Shown only to owners.</p>
      </div>

      {logs.isLoading ? <p className="mt-6 text-sm text-gray-500">Loading audit trail...</p> : null}
      {logs.isError ? <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">Could not load the audit trail. Your financial data is unchanged.</div> : null}
      {logs.data?.data.length === 0 ? <div className="mt-6"><EmptyState title="No audit entries yet" description="Important workspace changes will appear here." /></div> : null}

      {logs.data?.data.length ? (
        <div className="mt-6 overflow-hidden rounded-xl border border-gray-200 bg-white">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-gray-200 bg-gray-50 text-xs uppercase tracking-wide text-gray-500">
              <tr><th className="px-4 py-3 font-medium">When</th><th className="px-4 py-3 font-medium">Actor</th><th className="px-4 py-3 font-medium">Change</th><th className="hidden px-4 py-3 font-medium md:table-cell">Reason</th></tr>
            </thead>
            <tbody>
              {logs.data.data.map((log) => (
                <tr key={log.id} className="cursor-pointer border-b border-gray-100 last:border-0 hover:bg-slate-50" onClick={() => setSelected(log)}>
                  <td className="whitespace-nowrap px-4 py-3 text-gray-500">{new Date(log.occurred_at).toLocaleString()}</td>
                  <td className="px-4 py-3"><div className="font-medium text-gray-900">{log.actor_name ?? log.actor_type}</div><span className={`inline-block rounded-full px-2 py-0.5 text-xs font-semibold ${actorStyle[log.actor_type]}`}>{log.actor_type}</span></td>
                  <td className="px-4 py-3"><div className="font-medium capitalize text-gray-900">{formatAction(log.action)}</div><div className="text-xs text-gray-400">{log.resource_type}</div></td>
                  <td className="hidden max-w-xs truncate px-4 py-3 text-gray-500 md:table-cell">{log.reason ?? '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}

      {logs.data && totalPages > 1 ? <div className="mt-4 flex items-center gap-2"><Button variant="secondary" disabled={page === 1} onClick={() => setPage((value) => value - 1)}>Previous</Button><span className="text-sm text-gray-600">Page {page} of {totalPages}</span><Button variant="secondary" disabled={page === totalPages} onClick={() => setPage((value) => value + 1)}>Next</Button></div> : null}

      <Modal open={Boolean(selected)} onClose={() => setSelected(undefined)} title="Audit entry">
        {selected ? <div className="space-y-5"><div className="grid grid-cols-2 gap-3 text-sm"><div><p className="text-gray-500">Action</p><p className="mt-1 font-medium capitalize">{formatAction(selected.action)}</p></div><div><p className="text-gray-500">Actor</p><p className="mt-1 font-medium">{selected.actor_name ?? selected.actor_type}</p></div>{selected.reason ? <div className="col-span-2"><p className="text-gray-500">Reason</p><p className="mt-1">{selected.reason}</p></div> : null}</div><div><h2 className="mb-2 text-sm font-semibold">Before</h2><Snapshot data={selected.before_data} /></div><div><h2 className="mb-2 text-sm font-semibold">After</h2><Snapshot data={selected.after_data} /></div></div> : null}
      </Modal>
    </div>
  );
}
