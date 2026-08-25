import { useEffect, useMemo, useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useAuth } from '@/app/providers/auth-provider';
import { updateSettings } from '@/features/settings/api/settings.api';
import type { UserSettings } from '@/shared/api/types';

export function SettingsPage() {
  const { auth, refresh } = useAuth();
  const [form, setForm] = useState<UserSettings | null>(null);

  useEffect(() => {
    if (auth) setForm(auth.settings);
  }, [auth]);

  const notSaved = useMemo(
    () => form && auth && JSON.stringify(form) !== JSON.stringify(auth.settings),
    [form, auth],
  );

  const save = useMutation({
    mutationFn: (input: Partial<UserSettings>) => updateSettings(input),
    onSuccess: () => void refresh(),
  });

  if (!auth || !form) return null;
  const { user, workspace, role } = auth;

  const set = <K extends keyof UserSettings>(key: K, value: UserSettings[K]) =>
    setForm((f) => (f ? { ...f, [key]: value } : f));

  const rows = (items: Array<[string, string]>) =>
    items.map(([label, value]) => (
      <div
        key={label}
        className="flex items-center justify-between border-b border-gray-100 py-3 last:border-0"
      >
        <span className="text-sm text-gray-500">{label}</span>
        <span className="text-sm font-medium text-gray-800">{value || '-'}</span>
      </div>
    ));

  const Toggle = ({ label, key, value }: { label: string; key: keyof UserSettings; value: boolean }) => (
    <label className="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
      <span className="text-sm text-gray-700">{label}</span>
      <input
        type="checkbox"
        checked={value}
        onChange={(e) => set(key, e.target.checked)}
        className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
      />
    </label>
  );

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">Pengaturan</h1>
        <p className="mt-1 text-sm text-gray-500">Preferensi akun dan ruang kerja Anda.</p>
      </div>

      <section className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="mb-2 text-base font-semibold text-gray-900">Akun</h2>
        <div>{rows([['Nama', user.name], ['Email', user.email], ['Peran', role]])}</div>
      </section>

      <section className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="mb-2 text-base font-semibold text-gray-900">Ruang Kerja</h2>
        <div>
          {rows([
            ['Nama', workspace.name],
            ['Mata uang dasar', workspace.base_currency],
            ['Zona waktu', workspace.timezone],
          ])}
        </div>
      </section>

      <section className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="mb-2 text-base font-semibold text-gray-900">Preferensi</h2>
        <Toggle label="AI Insights" key="ai_insights_enabled" value={form.ai_insights_enabled} />
        <Toggle label="AI Copilot" key="ai_copilot_enabled" value={form.ai_copilot_enabled} />
        <Toggle label="Notifikasi" key="notifications_enabled" value={form.notifications_enabled} />

        <div className="py-3 border-b border-gray-100 last:border-0">
          <label htmlFor="budget_threshold" className="text-sm text-gray-700">
            Ambang peringatan anggaran (%)
          </label>
          <input
            id="budget_threshold"
            type="number"
            min={0}
            max={100}
            value={form.budget_warning_threshold ?? ''}
            onChange={(e) => {
              const n = e.target.value === '' ? 0 : Number(e.target.value);
              set('budget_warning_threshold', n);
            }}
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none"
          />
        </div>

        <div className="py-3 border-b border-gray-100 last:border-0">
          <label htmlFor="low_balance" className="text-sm text-gray-700">
            Ambang saldo rendah (minor units)
          </label>
          <input
            id="low_balance"
            type="number"
            min={0}
            value={form.low_balance_threshold ?? ''}
            onChange={(e) =>
              set('low_balance_threshold', e.target.value === '' ? null : Number(e.target.value))
            }
            className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none"
          />
        </div>

        {save.isError && (
          <p className="mt-3 text-sm text-red-600">
            Gagal menyimpan perubahan. Silakan coba lagi.
          </p>
        )}
        {save.isPending && <p className="mt-3 text-sm text-gray-500">Menyimpan…</p>}

        <button
          type="button"
          disabled={!notSaved || save.isPending}
          onClick={() => save.mutate(form)}
          className="mt-4 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Simpan Perubahan
        </button>
      </section>
    </div>
  );
}
