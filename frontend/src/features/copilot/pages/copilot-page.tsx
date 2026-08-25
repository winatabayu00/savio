import { useState } from 'react';
import { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { TextField } from '@/shared/components/ui/text-field';
import { useAIStatus, useCopilot } from '@/features/ai/hooks/use-ai';
import { useSpeechRecognition, useSpeechSynthesis } from '@/features/ai/hooks/use-voice';
import type { CopilotDTO } from '@/features/ai/api/ai.api';

const SUGGESTIONS = [
  'Why did I spend more this month?',
  'Where did my money go?',
  'What are my largest recurring expenses?',
  'Which budget is at risk?',
  'What does my forecast look like?',
  'Can I afford a 15M laptop?',
  'Am I on track for my goal?',
];

export function CopilotPage() {
  const { data: status } = useAIStatus();
  const { ask } = useCopilot();
  const [mode, setMode] = useState<'talk' | 'text'>('text');
  const [question, setQuestion] = useState('');
  const [response, setResponse] = useState<CopilotDTO | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [thinking, setThinking] = useState(false);
  const speech = useSpeechSynthesis();

  const run = async (q: string) => {
    setQuestion(q);
    setError(null);
    setThinking(true);
    setResponse(null);
    try {
      const res = await ask.mutateAsync(q);
      setResponse(res);
      if (mode === 'talk') speech.speak(res.answer);
    } catch (err) {
      const apiErr = err as ApiError;
      setError(
        apiErr?.status === 503 || apiErr?.status === 422
          ? 'AI is unavailable right now. Try again shortly, or check the deterministic pages (Forecast, Scenarios, Analytics) in the meantime.'
          : apiErr?.message ?? 'Something went wrong.',
      );
    } finally {
      setThinking(false);
    }
  };

  const disabled = status?.enabled === false;

  const onVoiceFinal = (text: string) => {
    setQuestion(text);
    if (text.trim()) void run(text);
  };
  const mic = useSpeechRecognition(onVoiceFinal);

  return (
    <div className="mx-auto max-w-3xl">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold">Savio Copilot</h1>
        <div className="inline-flex rounded-lg border border-gray-200 bg-white p-0.5 text-sm">
          <button
            type="button"
            onClick={() => {
              setMode('text');
              speech.stopSpeaking();
            }}
            className={`rounded-md px-3 py-1.5 font-medium transition-colors ${
              mode === 'text' ? 'bg-brand text-white' : 'text-gray-600 hover:bg-gray-50'
            }`}
          >
            Text
          </button>
          <button
            type="button"
            onClick={() => setMode('talk')}
            disabled={!mic.supported}
            className={`rounded-md px-3 py-1.5 font-medium transition-colors disabled:opacity-50 ${
              mode === 'talk' ? 'bg-brand text-white' : 'text-gray-600 hover:bg-gray-50'
            }`}
            title={mic.supported ? 'Chat by voice' : 'Mic is not supported in this browser'}
          >
            Talk
          </button>
        </div>
      </div>
      <p className="mt-1 text-sm text-gray-500">
        Ask about your finances in plain language. Answers are grounded in your real data with deterministic tools.
      </p>

      {disabled ? (
        <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
          AI is disabled for this installation. Finance features still work; Copilot is unavailable.
        </div>
      ) : null}

      {mode === 'talk' && !mic.supported ? (
        <div className="mt-4 rounded-xl border border-gray-200 bg-gray-50 p-4 text-sm text-gray-600">
          Voice chat needs Chrome, Edge, or Safari. Text mode works in every browser.
        </div>
      ) : null}

      <div className="mt-6 space-y-3">
        <form
          className="flex gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (question.trim() && !thinking) void run(question);
          }}
        >
          <div className="flex-1">
            <TextField
              label="Ask Copilot"
              placeholder={mode === 'talk' ? 'Say it, or type a question…' : 'e.g. Can I afford a 20M laptop?'}
              value={question}
              disabled={disabled}
              onChange={(e) => setQuestion(e.target.value)}
            />
          </div>
          {mode === 'talk' && mic.supported ? (
            <button
              type="button"
              disabled={disabled || thinking}
              onClick={() => (mic.listening ? mic.stop() : mic.start())}
              title={mic.listening ? 'Stop listening' : 'Speak your question'}
              aria-label={mic.listening ? 'Stop listening' : 'Speak your question'}
              className={`mt-6 h-10 w-10 shrink-0 rounded-full border-2 flex items-center justify-center transition-colors disabled:opacity-50 ${
                mic.listening
                  ? 'border-red-500 bg-red-500 text-white animate-pulse'
                  : 'border-brand text-brand hover:bg-brand/10'
              }`}
            >
              {mic.listening ? (
                <span className="h-3 w-3 rounded-sm bg-white" aria-hidden />
              ) : (
                <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
                  <rect x="9" y="2" width="6" height="12" rx="3" />
                  <path d="M5 10a7 7 0 0 0 14 0M12 17v5" />
                </svg>
              )}
            </button>
          ) : null}
          <Button type="submit" disabled={disabled || thinking || !question.trim()} className="mt-6">
            {thinking ? 'Analyzing…' : 'Ask'}
          </Button>
        </form>
        {mic.error ? <p className="text-sm text-red-600">{mic.error}</p> : null}
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {SUGGESTIONS.map((s) => (
          <button
            key={s}
            type="button"
            disabled={disabled || thinking}
            onClick={() => void run(s)}
            className="rounded-full border border-gray-200 bg-white px-3 py-1.5 text-xs text-gray-600 hover:border-brand hover:text-brand disabled:opacity-50"
          >
            {s}
          </button>
        ))}
      </div>

      {thinking ? (
        <div className="mt-6 rounded-2xl border border-gray-200 bg-white p-6 text-sm text-gray-500">
          Gathering your facts and forming an answer…
        </div>
      ) : null}

      {error ? (
        <div className="mt-6 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">{error}</div>
      ) : null}

      {response ? (
        <div className="mt-6 rounded-2xl border border-gray-200 bg-white p-6">
          <span className="rounded-full bg-brand/10 px-2.5 py-1 text-xs font-semibold text-brand">
            Copilot · grounded answer
          </span>

          {response.facts.length > 0 ? (
            <div className="mt-5">
              <h3 className="text-sm font-semibold uppercase tracking-wide text-gray-500">Supporting facts</h3>
              <ul className="mt-2 space-y-2">
                {response.facts.map((f, i) => (
                  <li key={i} className="flex items-center justify-between rounded-lg bg-gray-50 px-4 py-2.5 text-sm">
                    <span className="text-gray-600">
                      <span className="mr-2 rounded bg-gray-200 px-1.5 py-0.5 text-xs font-medium text-gray-500">{f.tool}</span>
                      {f.label}
                    </span>
                    <span className="font-semibold text-gray-900">{f.value}</span>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}

          <div className="mt-5 border-t border-gray-100 pt-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold uppercase tracking-wide text-gray-500">Answer</h3>
              {speech.supported ? (
                <div className="flex gap-1">
                  <button
                    type="button"
                    onClick={() => (speech.speaking ? speech.stopSpeaking() : speech.speak(response.answer))}
                    className="inline-flex items-center gap-1.5 rounded-full border border-gray-200 px-2.5 py-1 text-xs font-medium text-gray-600 hover:border-brand hover:text-brand"
                    aria-label={speech.speaking ? 'Stop reading' : 'Read answer aloud'}
                  >
                    {speech.speaking ? (
                      <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
                        <rect x="6" y="6" width="4" height="12" rx="1" />
                        <rect x="14" y="6" width="4" height="12" rx="1" />
                      </svg>
                    ) : (
                      <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
                        <path d="M3 10v4h3l4 4V6l-4 4H3z" />
                        <path d="M14 9a4 4 0 0 1 0 6M17 6a8 8 0 0 1 0 12" />
                      </svg>
                    )}
                    {speech.speaking ? 'Stop' : 'Listen'}
                  </button>
                </div>
              ) : null}
            </div>
            <p className="mt-2 text-gray-800">{response.answer}</p>
          </div>

          {response.clarification ? (
            <p className="mt-4 rounded-lg bg-blue-50 p-3 text-sm text-blue-800">
              Need more detail: {response.clarification}
            </p>
          ) : null}

          {response.actions.length > 0 ? (
            <div className="mt-4">
              <h4 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Suggested next steps</h4>
              <ul className="mt-1 list-inside list-disc space-y-0.5 text-sm text-gray-600">
                {response.actions.map((a, i) => (
                  <li key={i}>{a}</li>
                ))}
              </ul>
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}