import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useAuth } from '@/app/providers/auth-provider';
import { getAIConfig, updateAIConfig } from '@/features/settings/api/ai-config.api';
import {
  getTelegramConfig,
  updateTelegramConfig,
  registerTelegramWebhook,
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
        className="d-flex align-items-center justify-content-between border-bottom py-3"
      >
        <span className="fs-13 text-muted">{label}</span>
        <span className="fs-13 fw-medium text-dark">{value || '-'}</span>
      </div>
    ));

  const Toggle = ({ label, key, value }: { label: string; key: keyof UserSettings; value: boolean }) => (
    <label className="d-flex align-items-center justify-content-between py-3 border-bottom">
      <span className="fs-13 text-secondary">{label}</span>
      <input
        type="checkbox"
        checked={value}
        onChange={(e) => set(key, e.target.checked)}
        className="form-check-input mt-0"
      />
    </label>
  );

  return (
    <div className="d-flex flex-column gap-3">
      <div>
        <h1 className="fs-20 fw-bolder text-dark mb-0">Pengaturan</h1>
        <p className="mt-1 fs-13 text-muted mb-0">Preferensi akun dan ruang kerja Anda.</p>
      </div>

      <section className="card shadow-sm">
        <div className="card-body">
        <h2 className="fs-16 fw-semibold text-dark mb-3">Akun</h2>
        <div>{rows([['Nama', user.name], ['Email', user.email], ['Peran', role]])}</div>
        </div>
      </section>

      <section className="card shadow-sm">
        <div className="card-body">
        <h2 className="fs-16 fw-semibold text-dark mb-3">Ruang Kerja</h2>
        <div>
          {rows([
            ['Nama', workspace.name],
            ['Mata uang dasar', workspace.base_currency],
            ['Zona waktu', workspace.timezone],
          ])}
        </div>
        </div>
      </section>

      <section className="card shadow-sm">
        <div className="card-body">
        <h2 className="fs-16 fw-semibold text-dark mb-3">Preferensi</h2>
        <Toggle label="AI Insights" key="ai_insights_enabled" value={form.ai_insights_enabled} />
        <Toggle label="AI Copilot" key="ai_copilot_enabled" value={form.ai_copilot_enabled} />
        <Toggle label="Notifikasi" key="notifications_enabled" value={form.notifications_enabled} />

        <div className="py-3 border-bottom">
          <label htmlFor="budget_threshold" className="form-label mb-1">
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
            className="form-control"
          />
        </div>

        <div className="py-3 border-bottom">
          <label htmlFor="low_balance" className="form-label mb-1">
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
            className="form-control"
          />
        </div>

        {save.isError && (
          <p className="mt-3 fs-13 text-danger mb-0">
            Gagal menyimpan perubahan. Silakan coba lagi.
          </p>
        )}
        {save.isPending && <p className="mt-3 fs-13 text-muted mb-0">Menyimpan…</p>}

        <button
          type="button"
          disabled={!notSaved || save.isPending}
          onClick={() => save.mutate(form)}
          className="btn btn-success mt-3"
        >
          Simpan Perubahan
        </button>
        </div>
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
    return <p className="fs-13 text-danger">Gagal memuat konfigurasi AI.</p>;
  }

  const dirty =
    !!draft &&
    (draft.enabled !== undefined || draft.provider !== undefined ||
      draft.base_url !== undefined || draft.model !== undefined ||
      draft.persona !== undefined ||
      draft.timeout_seconds !== undefined || (draft.api_key ?? '') !== '');

  const Input = (props: React.InputHTMLAttributes<HTMLInputElement>) => (
    <input
      {...props}
      className="form-control"
    />
  );

  return (
    <section className="card shadow-sm">
      <div className="card-body">
      <h2 className="fs-16 fw-semibold text-dark mb-1">Kecerdasan Buatan (AI)</h2>
      <p className="mb-3 fs-13 text-muted">
        Konfigurasi penyedia AI. Simpan di sini menggantikan pengaturan via variabel lingkungan (ENV).
      </p>

      {!canEdit && (
        <p className="mb-3 fs-13 text-muted">
          Hanya pemilik ruang kerja yang dapat mengubah konfigurasi AI.
        </p>
      )}

      <label className="d-flex align-items-center justify-content-between py-2">
        <span className="fs-13 text-secondary">Aktifkan AI</span>
        <input
          type="checkbox"
          checked={f.enabled}
          onChange={(e) => update({ enabled: e.target.checked })}
          className="form-check-input mt-0"
        />
      </label>

      <div className="mt-2">
        <label htmlFor="ai_provider" className="form-label mb-1">
          Penyedia
        </label>
        <select
          id="ai_provider"
          value={f.provider === 'mock' ? 'mock' : 'openai'}
          onChange={(e) => update({ provider: e.target.value })}
          disabled={!canEdit}
          className="form-select"
        >
          <option value="openai">OpenAI-compatible (OpenAI, Groq, Ollama...)</option>
          <option value="mock">Mock (uji coba offline)</option>
        </select>
      </div>

      <div className="mt-3">
        <label htmlFor="ai_base_url" className="form-label mb-1">
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
        <label htmlFor="ai_api_key" className="form-label mb-1">
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
        <p className="mt-1 fs-12 text-muted mb-0">
          Kunci tersimpan memiliki: {f.api_key_masked || '—'}. Kosongkan untuk mempertahankan kunci yang ada.
        </p>
      </div>

      <div className="mt-3">
        <label htmlFor="ai_model" className="form-label mb-1">
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
        <label htmlFor="ai_persona" className="form-label mb-1">
          Personality AI
        </label>
        <select
          id="ai_persona"
          value={f.persona || 'balanced'}
          onChange={(e) => update({ persona: e.target.value })}
          disabled={!canEdit}
          className="form-select"
        >
          <option value="balanced">Savio Copilot — Asisten keuangan netral</option>
          <option value="lenna">Lenna — Penasihat keuangan pribadi</option>
        </select>
        <p className="mt-1 fs-12 text-muted mb-0">
          Personality membentuk nada jawaban AI Copilot dan AI Insights. Lenna berbicara seperti
          penasihat keuangan pribadi yang tenang dan berwawasan ke depan.
        </p>
      </div>

      <div className="mt-3">
        <label htmlFor="ai_timeout" className="form-label mb-1">
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
        <p className="mt-3 fs-13 text-danger mb-0">Gagal menyimpan konfigurasi AI. Silakan coba lagi.</p>
      )}
      {saveConfig.isPending && <p className="mt-3 fs-13 text-muted mb-0">Menyimpan…</p>}

      {canEdit && (
        <button
          type="button"
          disabled={!dirty || saveConfig.isPending}
          onClick={() => saveConfig.mutate(draft ?? {})}
          className="btn btn-success mt-3"
        >
          Simpan Konfigurasi AI
        </button>
      )}
      </div>
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

  const registerWebhook = useMutation({
    mutationFn: (url: string) => registerTelegramWebhook(url),
    onSuccess: () => void tgConfig.refetch(),
  });

  const [webhookBase, setWebhookBase] = useState('');
  const cfg = tgConfig.data;
  const [draft, setDraft] = useState<TelegramConfigInput | null>(null);

  const update = (patch: TelegramConfigInput) =>
    setDraft((d) => (cfg ? { ...d, ...patch } : d));

  const f = cfg ? { ...cfg, ...draft } : null;
  if (!f) {
    if (tgConfig.isPending || !canEdit) return null;
    return <p className="fs-13 text-danger">Gagal memuat konfigurasi Telegram.</p>;
  }

  const dirty =
    !!draft &&
    (draft.enabled !== undefined ||
      draft.chat_id !== undefined ||
      (draft.bot_token ?? '') !== '');

  const Input = (props: React.InputHTMLAttributes<HTMLInputElement>) => (
    <input
      {...props}
      className="form-control"
    />
  );

  return (
    <section className="card shadow-sm">
      <div className="card-body">
      <h2 className="fs-16 fw-semibold text-dark mb-1">Telegram Recap</h2>
      <p className="mb-3 fs-13 text-muted">
        Sambungkan bot Telegram agar pengeluaran bisa dicatat langsung dari chat.
        Cukup kirim <span className="fw-medium">nama + nominal</span>, misal{' '}
        <span className="fw-medium">chocolate hazelnut dutch 24000</span> — Savio
        mencatatnya ke ruang kerja ini dengan kategori otomatis.
      </p>

      {!canEdit && (
        <p className="mb-3 fs-13 text-muted">
          Hanya pemilik ruang kerja yang dapat mengubah konfigurasi Telegram.
        </p>
      )}

      <label className="d-flex align-items-center justify-content-between py-2">
        <span className="fs-13 text-secondary">Aktifkan Telegram Recap</span>
        <input
          type="checkbox"
          checked={f.enabled}
          onChange={(e) => update({ enabled: e.target.checked })}
          className="form-check-input mt-0"
        />
      </label>

      <div className="mt-3">
        <label htmlFor="tg_token" className="form-label mb-1">
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
        <p className="mt-1 fs-12 text-muted mb-0">
          Token tersimpan: {f.bot_token_masked || '—'}. Kosongkan untuk mempertahankan token yang ada.
        </p>
      </div>

      <div className="mt-3">
        <label htmlFor="tg_chat" className="form-label mb-1">
          ID Chat yang Diizinkan
        </label>
        <Input
          id="tg_chat"
          value={draft?.chat_id ?? f.chat_id}
          onChange={(e) => update({ chat_id: e.target.value })}
          placeholder="contoh: 123456789"
          disabled={!canEdit}
        />
        <p className="mt-1 fs-12 text-muted mb-0">
          Kirim <span className="fw-medium">/start</span> ke bot Anda, lalu aktifkan{' '}
          <span className="fw-medium">Mode Pengembang</span> &gt;{' '}
          <span className="fw-medium">Chat ID</span> di Telegram untuk melihat ID chat
          pribadi Anda. Kosongkan agar pesan hanya dibalas instruksi.
        </p>
      </div>

      <div className="mt-3 alert alert-success p-3 fs-12 mb-0">
        Transaksi yang masuk akan dicatat sebagai pengeluaran <b>POSTED</b> (langsung
        memengaruhi saldo) dengan kategori otomatis dan sumber 'TELEGRAM'.
      </div>

      <div className="mt-4 border-top pt-3">
        <h3 className="fs-13 fw-semibold text-dark mb-0">Webhook (pusat langsung)</h3>
        <p className="mb-2 mt-1 fs-12 text-muted">
          Default: worker mengambil pesan setiap ±25 detik (long-poll). Untuk penerimaan yang
          instan, daftarkan webhook ke URL <b>https</b> publik yang menuju ke backend Savio —
          misalnya tunnel{' '}
          <span className="fw-medium">ngrok http 8080</span> lalu pakai URL
          <span className="fw-medium"> https://xxx.ngrok-free.app</span>. Setiap pesan
          langsung masuk tanpa tundaan polling.
        </p>

        {f.webhook_url ? (
          <div className="bg-light p-3 fs-12 rounded-3">
            <p className="fw-medium text-secondary mb-1">Webhook terdaftar:</p>
            <code className="d-block text-success text-break">{f.webhook_url}</code>
            {canEdit && (
              <button
                type="button"
                disabled={registerWebhook.isPending}
                onClick={() => registerWebhook.mutate('')}
                className="btn btn-outline-secondary btn-sm mt-2"
              >
                Hapus Webhook (kembali ke long-poll)
              </button>
            )}
          </div>
        ) : (
          canEdit && (
            <div className="d-flex gap-2">
              <input
                type="url"
                value={webhookBase}
                onChange={(e) => setWebhookBase(e.target.value)}
                placeholder="https://xxx.ngrok-free.app"
                className="form-control"
              />
              <button
                type="button"
                disabled={registerWebhook.isPending || !webhookBase}
                onClick={() => registerWebhook.mutate(webhookBase)}
                className="btn btn-success flex-shrink-0"
              >
                Daftarkan Webhook
              </button>
            </div>
          )
        )}
        {registerWebhook.isError && (
          <p className="mt-2 fs-12 text-danger mb-0">
            Gagal mendaftarkan webhook. Pastikan URL https publik mengarah ke API Savio.
            Worker tidak diperlukan untuk webhook. Long-poll masih aktif.
          </p>
        )}
      </div>

      {saveConfig.isError && (
        <p className="mt-3 fs-13 text-danger mb-0">Gagal menyimpan konfigurasi Telegram. Silakan coba lagi.</p>
      )}
      {saveConfig.isPending && <p className="mt-3 fs-13 text-muted mb-0">Menyimpan…</p>}

      {canEdit && (
        <button
          type="button"
          disabled={!dirty || saveConfig.isPending}
          onClick={() => saveConfig.mutate(draft ?? {})}
          className="btn btn-success mt-3"
        >
          Simpan Konfigurasi Telegram
        </button>
      )}
      </div>
    </section>
  );
}
