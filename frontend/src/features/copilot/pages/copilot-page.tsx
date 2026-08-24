import { useState } from 'react';
import { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { TextField } from '@/shared/components/ui/text-field';
import { useAIStatus, useCopilot } from '@/features/ai/hooks/use-ai';
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
  const [question, setQuestion] = useState('');
  const [response, setResponse] = useState<CopilotDTO | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [thinking, setThinking] = useState(false);

  const run = async (q: string) => {
    setQuestion(q);
    setError(null);
    setThinking(true);
    setResponse(null);
    try {
      const res = await ask.mutateAsync(q);
      setResponse(res);
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

  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="text-2xl font-semibold">Savio Copilot</h1>
      <p className="mt-1 text-sm text-gray-500">
        Ask about your finances in plain language. Answers are grounded in your real data with deterministic tools.
      </p>

      {disabled ? (
        <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
          AI is disabled for this installation. Finance features still work; Copilot is unavailable.
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
              placeholder="e.g. Can I afford a 20M laptop?"
              value={question}
              disabled={disabled}
              onChange={(e) => setQuestion(e.target.value)}
            />
          </div>
          <Button type="submit" disabled={disabled || thinking || !question.trim()} className="mt-6">
            {thinking ? 'Analyzing…' : 'Ask'}
          </Button>
        </form>
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
            <h3 className="text-sm font-semibold uppercase tracking-wide text-gray-500">Answer</h3>
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