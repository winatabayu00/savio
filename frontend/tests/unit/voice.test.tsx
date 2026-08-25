import { afterEach, describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { render, renderHook, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { server, AUTH_BASE, csrfHandler } from '../mocks/handlers';
import { api } from '@/shared/api/client';
import { CopilotPage } from '@/features/copilot/pages/copilot-page';
import { useSpeechRecognition, useSpeechSynthesis } from '@/features/ai/hooks/use-voice';

api.defaults.baseURL = AUTH_BASE;

let lastRecognition: FakeRecognition | null = null;

class FakeRecognition extends EventTarget {
  lang = '';
  interimResults = true;
  maxAlternatives = 1;
  onresult: SpeechRecognition['onresult'] = null;
  onerror: SpeechRecognition['onerror'] = null;
  onend: SpeechRecognition['onend'] = null;
  start = vi.fn();
  stop = vi.fn();
  abort = vi.fn();
  constructor() {
    super();
    lastRecognition = this;
  }
}

const speakMock = vi.fn();
const cancelMock = vi.fn();

function stubVoiceSupport() {
  vi.stubGlobal('SpeechRecognition', FakeRecognition);
  lastRecognition = null;
}

function stubTts() {
  vi.stubGlobal('speechSynthesis', {
    speak: speakMock,
    cancel: cancelMock,
    getVoices: () => [
      { lang: 'id-ID', name: 'Google Indonesian' },
      { lang: 'en-US', name: 'Samantha' },
    ],
  });
  vi.stubGlobal(
    'SpeechSynthesisUtterance',
    class {
      lang = '';
      rate = 1;
      voice = null;
      onend: (() => void) | null = null;
      onerror: (() => void) | null = null;
      constructor(public text = '') {}
    },
  );
}

function fireFinal(text: string) {
  const rec = lastRecognition;
  if (!rec) throw new Error('no recognition instance');
  const ev = { results: [[{ transcript: text }]], resultIndex: 0 } as unknown as SpeechRecognitionEvent;
  act(() => rec.onresult?.call(rec, ev));
  act(() => rec.onend?.call(rec, new Event('end')));
}

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <CopilotPage />
    </QueryClientProvider>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  speakMock.mockReset();
  cancelMock.mockReset();
  server.resetHandlers();
});

describe('useSpeechRecognition', () => {
  it('reports unsupported when the browser has no SpeechRecognition', () => {
    const { result } = renderHook(() => useSpeechRecognition(vi.fn()));
    expect(result.current.supported).toBe(false);
    act(() => result.current.start());
    expect(result.current.error).toBe('Mic is not supported in this browser.');
  });

  it('starts listening in id-ID and stops cleanly', () => {
    stubVoiceSupport();
    const { result } = renderHook(() => useSpeechRecognition(vi.fn()));
    act(() => result.current.start());
    const rec = lastRecognition!;
    expect(rec.start).toHaveBeenCalledTimes(1);
    expect(rec.lang).toBe('id-ID');
    expect(result.current.listening).toBe(true);
    act(() => result.current.stop());
    expect(rec.stop).toHaveBeenCalledTimes(1);
    expect(result.current.listening).toBe(false);
  });

  it('sends the final transcript to the callback', () => {
    stubVoiceSupport();
    const onFinal = vi.fn();
    const { result } = renderHook(() => useSpeechRecognition(onFinal));
    act(() => result.current.start());
    fireFinal('berapa saldo saya');
    expect(onFinal).toHaveBeenCalledWith('berapa saldo saya');
  });

  it('surfaces a permission denial', () => {
    stubVoiceSupport();
    const { result } = renderHook(() => useSpeechRecognition(vi.fn()));
    act(() => result.current.start());
    const rec = lastRecognition!;
    act(() =>
      rec.onerror?.call(rec, { error: 'not-allowed' } as unknown as SpeechRecognitionErrorEvent),
    );
    expect(result.current.error).toBe('Microphone permission denied.');
  });
});

describe('useSpeechSynthesis', () => {
  it('no-ops without tts support', () => {
    const { result } = renderHook(() => useSpeechSynthesis());
    act(() => result.current.speak('hello'));
    expect(result.current.supported).toBe(false);
    expect(result.current.speaking).toBe(false);
  });

  it('reads Indonesian with an id voice and stops on demand', () => {
    stubTts();
    const { result } = renderHook(() => useSpeechSynthesis());
    act(() => result.current.speak('Halo'));
    expect(speakMock).toHaveBeenCalledTimes(1);
    const utterance = speakMock.mock.calls[0][0] as { lang: string; rate: number; voice: { lang: string } };
    expect(utterance.lang).toBe('id-ID');
    expect(utterance.rate).toBe(1);
    expect(utterance.voice.lang).toBe('id-ID');
    expect(result.current.speaking).toBe(true);
    cancelMock.mockReset();
    act(() => result.current.stopSpeaking());
    expect(cancelMock).toHaveBeenCalledTimes(1);
    expect(result.current.speaking).toBe(false);
  });

  it('clears speaking when the utterance ends', () => {
    stubTts();
    const { result } = renderHook(() => useSpeechSynthesis());
    act(() => result.current.speak('selesai'));
    const utterance = speakMock.mock.calls[0][0] as { onend: (() => void) | null };
    act(() => utterance.onend?.());
    expect(result.current.speaking).toBe(false);
  });
});

describe('Copilot talk mode', () => {
  const answer = () =>
    HttpResponse.json({
      success: true,
      data: {
        answer: 'Pengeluaranmu naik 12%.',
        facts: [{ tool: 'compare_periods', label: 'Pengeluaran bulan ini', value: '12.000.000' }],
        tool_used: 'compare_periods',
        sources: [],
        actions: [],
      },
    });

  it('disables Talk and shows a hint on unsupported browsers', async () => {
    server.use(
      csrfHandler,
      http.get(`${AUTH_BASE}/ai/status`, () =>
        HttpResponse.json({ success: true, data: { enabled: true, state: 'online' } }),
      ),
    );
    renderPage();
    await waitFor(() => expect(screen.getByText('Savio Copilot')).toBeInTheDocument());
    const talk = screen.getByRole('button', { name: 'Talk' });
    expect(talk).toBeDisabled();
  });

  it('asks by text when the mic is unsupported', async () => {
    server.use(
      csrfHandler,
      http.get(`${AUTH_BASE}/ai/status`, () =>
        HttpResponse.json({ success: true, data: { enabled: true, state: 'online' } }),
      ),
      http.post(`${AUTH_BASE}/ai/copilot`, answer),
    );
    renderPage();
    await waitFor(() =>
      expect(screen.getByPlaceholderText(/e.g. Can I afford a 20M laptop\?/)).toBeInTheDocument(),
    );
    await userEvent.type(screen.getByPlaceholderText(/e.g. Can I afford a 20M laptop\?/), 'berapa saldo?');
    await userEvent.click(screen.getByRole('button', { name: 'Ask' }));
    await waitFor(() => expect(screen.getByText('Pengeluaranmu naik 12%.')).toBeInTheDocument());
  });

  it('submits voice transcripts automatically and reads the answer aloud', async () => {
    stubVoiceSupport();
    stubTts();
    server.use(
      csrfHandler,
      http.get(`${AUTH_BASE}/ai/status`, () =>
        HttpResponse.json({ success: true, data: { enabled: true, state: 'online' } }),
      ),
      http.post(`${AUTH_BASE}/ai/copilot`, answer),
    );
    renderPage();
    await waitFor(() => expect(screen.getByRole('button', { name: 'Talk' })).toBeEnabled());

    await userEvent.click(screen.getByRole('button', { name: 'Talk' }));
    await userEvent.click(screen.getByRole('button', { name: 'Speak your question' }));
    expect(lastRecognition!.start).toHaveBeenCalledTimes(1);

    fireFinal('seberapa besar pengeluaran bulan ini');
    await waitFor(() => expect(screen.getByText('Pengeluaranmu naik 12%.')).toBeInTheDocument());
    expect(speakMock).toHaveBeenCalledWith(expect.objectContaining({ text: 'Pengeluaranmu naik 12%.' }));
  });

  it('replays the answer when Listen is clicked', async () => {
    stubTts();
    server.use(
      csrfHandler,
      http.get(`${AUTH_BASE}/ai/status`, () =>
        HttpResponse.json({ success: true, data: { enabled: true, state: 'online' } }),
      ),
      http.post(`${AUTH_BASE}/ai/copilot`, answer),
    );
    renderPage();
    await waitFor(() =>
      expect(screen.getByPlaceholderText(/e.g. Can I afford a 20M laptop\?/)).toBeInTheDocument(),
    );
    await userEvent.type(screen.getByPlaceholderText(/e.g. Can I afford a 20M laptop\?/), 'berapa saldo?');
    await userEvent.click(screen.getByRole('button', { name: 'Ask' }));
    await waitFor(() => expect(screen.getByText('Pengeluaranmu naik 12%.')).toBeInTheDocument());
    expect(speakMock).not.toHaveBeenCalled();

    await userEvent.click(screen.getByRole('button', { name: 'Read answer aloud' }));
    expect(speakMock).toHaveBeenCalledWith(expect.objectContaining({ text: 'Pengeluaranmu naik 12%.' }));
  });
});