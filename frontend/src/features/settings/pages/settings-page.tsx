import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useAuth } from '@/app/providers/auth-provider';
import { getAIConfig, updateAIConfig } from '@/features/settings/api/ai-config.api';
import {
  getTelegramConfig,
  updateTelegramConfig,
} from '@/features/settings/api/telegram-config.api';
import { updateSettings } from '@/features/settings/api/settings.api';
import type {
  AIConfigInput,
  TelegramConfigInput,
  UserSettings,
} from '@/shared/api/types';

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

      <AiConfigSection canEdit={role === 'OWNER'} />
      <TelegramConfigSection canEdit={role === 'OWNER'} />
    </div>
  );
}

function AiConfigSection({ canEdit }: { canEdit: boolean }) {
  const aiConfig = useQuery({
    queryKey: ['ai-config'],
    queryFn: getAIConfig,
    enabled: canEdit,
  });

  const saveConfig = useMutation({
    mutationFn: (input: AIConfigInput) => updateAIConfig(input),
    onSuccess: () => {
      void aiConfig.refetch();
      setDraft(null);
    },
  });

  const ai = aiConfig.data;
  const [draft, setDraft] = useState<AIConfigInput | null>(null);

  const update = (patch: AIConfigInput) =>
    setDraft((d) => (ai ? { ...d, ...patch } : d));

  const f = ai ? { ...ai, ...draft } : null;
  if (!f) {
    if (aiConfig.isPending || !canEdit) return null;
    return <p className="text-sm text-red-600">Gagal memuat konfigurasi AI.</p>;
  }

  const dirty =
    !!draft &&
    (draft.enabled !== undefined || draft.provider !== undefined ||
      draft.base_url !== undefined || draft.model !== undefined ||
      draft.timeout_seconds !== undefined || (draft.api_key ?? '') !== '');

  const Input = (props: React.InputHTMLAttributes<HTMLInputElement>) => (
    <input
      {...props}
      className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none"
    />
  );

  return (
    <section className="rounded-xl border border-gray-200 bg-white p-6">
      <h2 className="mb-1 text-base font-semibold text-gray-900">Kecerdasan Buatan (AI)</h2>
      <p className="mb-4 text-sm text-gray-500">
        Konfigurasi penyedia AI. Simpan di sini menggantikan pengaturan via variabel lingkungan (ENV).
      </p>

      {!canEdit && (
        <p className="mb-4 text-sm text-gray-500">
          Hanya pemilik ruang kerja yang dapat mengubah konfigurasi AI.
        </p>
      )}

      <label className="flex items-center justify-between py-2">
        <span className="text-sm text-gray-700">Aktifkan AI</span>
        <input
          type="checkbox"
          checked={f.enabled}
          onChange={(e) => update({ enabled: e.target.checked })}
          className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
        />
      </label>

      <div className="mt-2">
        <label htmlFor="ai_provider" className="text-sm text-gray-700">
          Penyedia
        </label>
        <select
          id="ai_provider"
          value={f.provider === 'mock' ? 'mock' : 'openai'}
          onChange={(e) => update({ provider: e.target.value })}
          disabled={!canEdit}
          className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none disabled:bg-gray-50"
        >
          <option value="openai">OpenAI-compatible (OpenAI, Groq, Ollama...)</option>
          <option value="mock">Mock (uji coba offline)</option>
        </select>
      </div>

      <div className="mt-3">
        <label htmlFor="ai_base_url" className="text-sm text-gray-700">
          Base URL API
        </label>
        <Input
          id="ai_base_url"
          value={f.base_url}
          onChange={(e) => update({ base_url: e.target.value })}
          placeholder="https://api.openai.com/v1"
          disabled={!canEdit}
        />
      </div>

      <div className="mt-3">
        <label htmlFor="ai_api_key" className="text-sm text-gray-700">
          API Key
        </label>
        <Input
          id="ai_api_key"
          type="password"
          value={draft?.api_key ?? ''}
          onChange={(e) => update({ api_key: e.target.value })}
          placeholder={f.api_key_masked || 'Belum diatur'}
          autoComplete="new-password"
          disabled={!canEdit}
        />
        <p className="mt-1 text-xs text-gray-400">
          Kunci tersimpan memiliki: {f.api_key_masked || '—'}. Kosongkan untuk mempertahankan kunci yang ada.
        </p>
      </div>

      <div className="mt-3">
        <label htmlFor="ai_model" className="text-sm text-gray-700">
          Model
        </label>
        <Input
          id="ai_model"
          value={f.model}
          onChange={(e) => update({ model: e.target.value })}
          placeholder="gpt-4o-mini"
          disabled={!canEdit}
        />
      </div>

      <div className="mt-3">
        <label htmlFor="ai_timeout" className="text-sm text-gray-700">
          Timeout (detik)
        </label>
        <Input
          id="ai_timeout"
          type="number"
          min={1}
          max={120}
          value={f.timeout_seconds}
          onChange={(e) => update({ timeout_seconds: Number(e.target.value) })}
          disabled={!canEdit}
        />
      </div>

      {saveConfig.isError && (
        <p className="mt-3 text-sm text-red-600">Gagal menyimpan konfigurasi AI. Silakan coba lagi.</p>
      )}
      {saveConfig.isPending && <p className="mt-3 text-sm text-gray-500">Menyimpan…</p>}

      {canEdit && (
        <button
          type="button"
          disabled={!dirty || saveConfig.isPending}
          onClick={() => saveConfig.mutate(draft ?? {})}
          className="mt-4 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Simpan Konfigurasi AI
        </button>
      )}
    </section>
  );
}

function TelegramConfigSection({ canEdit }: { canEdit: boolean }) {
  const tgConfig = useQuery({
    queryKey: ['telegram-config'],
    queryFn: getTelegramConfig,
    enabled: canEdit,
  });

  const saveConfig = useMutation({
    mutationFn: (input: TelegramConfigInput) => updateTelegramConfig(input),
    onSuccess: () => {
      void tgConfig.refetch();
      setDraft(null);
    },
  });

  const cfg = tgConfig.data;
  const [draft, setDraft] = useState<TelegramConfigInput | null>(null);

  const update = (patch: TelegramConfigInput) =>
    setDraft((d) => (cfg ? { ...d, ...patch } : d));

  const f = cfg ? { ...cfg, ...draft } : null;
  if (!f) {
    if (tgConfig.isPending || !canEdit) return null;
    return <p className="text-sm text-red-600">Gagal memuat konfigurasi Telegram.</p>;
  }

  const dirty =
    !!draft &&
    (draft.enabled !== undefined ||
      draft.chat_id !== undefined ||
      (draft.bot_token ?? '') !== '');

  const Input = (props: React.InputHTMLAttributes<HTMLInputElement>) => (
    <input
      {...props}
      className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none"
    />
  );

  return (
    <section className="rounded-xl border border-gray-200 bg-white p-6">
      <h2 className="mb-1 text-base font-semibold text-gray-900">Telegram Recap</h2>
      <p className="mb-4 text-sm text-gray-500">
        Sambungkan bot Telegram agar pengeluaran bisa dicatat langsung dari chat.
        Cukup kirim <span className="font-medium">nama + nominal</span>, misal{' '}
        <span className="font-medium">chocolate hazelnut dutch 24000</span> — Savio
        mencatatnya ke ruang kerja ini dengan kategori otomatis.
      </p>

      {!canEdit && (
        <p className="mb-4 text-sm text-gray-500">
          Hanya pemilik ruang kerja yang dapat mengubah konfigurasi Telegram.
        </p>
      )}

      <label className="flex items-center justify-between py-2">
        <span className="text-sm text-gray-700">Aktifkan Telegram Recap</span>
        <input
          type="checkbox"
          checked={f.enabled}
          onChange={(e) => update({ enabled: e.target.checked })}
          className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
        />
      </label>

      <div className="mt-3">
        <label htmlFor="tg_token" className="text-sm text-gray-700">
          Bot Token
        </label>
        <Input
          id="tg_token"
          type="password"
          value={draft?.bot_token ?? ''}
          onChange={(e) => update({ bot_token: e.target.value })}
          placeholder={f.bot_token_masked || 'Buat bot via @BotFather, lalu salin token-nya'}
          autoComplete="new-password"
          disabled={!canEdit}
        />
        <p className="mt-1 text-xs text-gray-400">
          Token tersimpan: {f.bot_token_masked || '—'}. Kosongkan untuk mempertahankan token yang ada.
        </p>
      </div>

      <div className="mt-3">
        <label htmlFor="tg_chat" className="text-sm text-gray-700">
          ID Chat yang Diizinkan
        </label>
        <Input
          id="tg_chat"
          value={draft?.chat_id ?? f.chat_id}
          onChange={(e) => update({ chat_id: e.target.value })}
          placeholder="contoh: 123456789"
          disabled={!canEdit}
        />
        <p className="mt-1 text-xs text-gray-400">
          Kirim <span className="font-medium">/start</span> ke bot Anda, lalu aktifkan{' '}
          <span className="font-medium">Mode Pengembang</span> &gt;{' '}
          <span className="font-medium">Chat ID</span> di Telegram untuk melihat ID chat
          pribadi Anda. Kosongkan agar pesan hanya dibalas instruksi.
        </p>
      </div>

      <div className="mt-3 rounded-lg bg-emerald-50 p-3 text-xs text-emerald-800">
        Transaksi yang masuk akan dicatat sebagai pengeluaran <b>POSTED</b> (langsung
        memengaruhi saldo) dengan kategori otomatis dan sumber 'TELEGRAM'.
      </div>

      {saveConfig.isError && (
        <p className="mt-3 text-sm text-red-600">Gagal menyimpan konfigurasi Telegram. Silakan coba lagi.</p>
      )}
      {saveConfig.isPending && <p className="mt-3 text-sm text-gray-500">Menyimpan…</p>}

      {canEdit && (
        <button
          type="button"
          disabled={!dirty || saveConfig.isPending}
          onClick={() => saveConfig.mutate(draft ?? {})}
          className="mt-4 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Simpan Konfigurasi Telegram
        </button>
      )}
    </section>
  );
}
