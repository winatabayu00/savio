import { useEffect, useRef, useState } from 'react';
import { ApiError } from '@/shared/api/client';
import { Button } from '@/shared/components/ui/button';
import { TextField } from '@/shared/components/ui/text-field';
import {
  useAIStatus,
  useConversation,
  useConversationMutations,
  useConversations,
} from '@/features/ai/hooks/use-ai';
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

function initialConversation() {
  return new URLSearchParams(window.location.search).get('conversation');
}

function AssistantAnswer({ response, onSpeak }: { response: CopilotDTO; onSpeak: (text: string) => void }) {
  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-5">
      <span className="rounded-full bg-brand/10 px-2.5 py-1 text-xs font-semibold text-brand">Lenna · grounded answer</span>
      {response.facts.length > 0 ? (
        <div className="mt-4">
          <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-500">Supporting facts</h3>
          <ul className="mt-2 space-y-2">
            {response.facts.map((fact, index) => (
              <li key={`${fact.label}-${index}`} className="flex flex-col gap-1 rounded-lg bg-gray-50 px-3 py-2 text-sm sm:flex-row sm:justify-between">
                <span className="text-gray-600">{fact.label}</span>
                <strong>{fact.value}</strong>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
      <div className="mt-4 border-t border-gray-100 pt-4">
        <div className="flex justify-between gap-3">
          <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-500">Answer</h3>
          <button type="button" onClick={() => onSpeak(response.answer)} className="text-xs font-medium text-brand" aria-label="Read answer aloud">Listen</button>
        </div>
        <p className="mt-2 text-gray-800">{response.answer}</p>
      </div>
      {response.clarification ? <p className="mt-4 rounded-lg bg-blue-50 p-3 text-sm text-blue-800">Need more detail: {response.clarification}</p> : null}
      {response.actions.length > 0 ? (
        <div className="mt-4 text-sm text-gray-600">
          <h4 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Suggested next steps</h4>
          <ul className="mt-1 list-inside list-disc">{response.actions.map((action) => <li key={action}>{action}</li>)}</ul>
        </div>
      ) : null}
    </div>
  );
}

export function CopilotPage() {
  const { data: status } = useAIStatus();
  const conversations = useConversations();
  const mutations = useConversationMutations();
  const [activeID, setActiveID] = useState<string | null>(initialConversation);
  const conversation = useConversation(activeID);
  const [mode, setMode] = useState<'talk' | 'text'>('text');
  const [question, setQuestion] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const endRef = useRef<HTMLDivElement>(null);
  const speech = useSpeechSynthesis();
  const busy = mutations.create.isPending || mutations.send.isPending;
  const disabled = status?.enabled === false;

  const select = (id: string | null) => {
    setActiveID(id);
    setDrawerOpen(false);
    const url = new URL(window.location.href);
    id ? url.searchParams.set('conversation', id) : url.searchParams.delete('conversation');
    window.history.replaceState({}, '', url);
  };

  useEffect(() => {
    endRef.current?.scrollIntoView?.({ behavior: 'smooth' });
  }, [conversation.data?.messages?.length]);

  useEffect(() => {
    if (!drawerOpen) return;
    const close = (event: KeyboardEvent) => event.key === 'Escape' && setDrawerOpen(false);
    document.addEventListener('keydown', close);
    return () => document.removeEventListener('keydown', close);
  }, [drawerOpen]);

  const run = async (text: string) => {
    const value = text.trim();
    if (!value || busy) return;
    setQuestion(value);
    setError(null);
    try {
      let id = activeID;
      if (!id) {
        const created = await mutations.create.mutateAsync();
        id = created.id;
        select(id);
      }
      const updated = await mutations.send.mutateAsync({ id, question: value });
      setQuestion('');
      const answer = updated.messages?.[updated.messages.length - 1]?.response?.answer;
      if (mode === 'talk' && answer) speech.speak(answer);
    } catch (err) {
      const apiErr = err as ApiError;
      setError(apiErr?.status === 429
        ? 'Copilot request limit reached. Try again shortly.'
        : apiErr?.status === 503 || apiErr?.status === 422
          ? 'AI is unavailable right now. Your finance data remains safe.'
          : apiErr?.message ?? 'Something went wrong.');
    }
  };

  const mic = useSpeechRecognition((text) => {
    setQuestion(text);
    if (text.trim()) void run(text);
  });

  const threadList = (
    <div className="flex h-full flex-col">
      <Button type="button" onClick={() => select(null)}>New conversation</Button>
      <div className="mt-4 flex-1 space-y-1 overflow-y-auto">
        {conversations.isPending ? <p className="p-2 text-sm text-gray-500">Loading conversations…</p> : null}
        {conversations.data?.length === 0 ? <p className="p-2 text-sm text-gray-500">No conversations yet.</p> : null}
        {conversations.data?.map((row) => (
          <div key={row.id} className={`group flex rounded-lg ${activeID === row.id ? 'bg-brand/10' : 'hover:bg-gray-50'}`}>
            <button type="button" onClick={() => select(row.id)} className="min-h-11 min-w-0 flex-1 truncate px-3 text-left text-sm font-medium">
              {row.title ?? 'New conversation'}
            </button>
            <button
              type="button"
              aria-label={`Delete ${row.title ?? 'conversation'}`}
              className="min-h-11 px-3 text-gray-400 hover:text-red-600"
              onClick={async () => {
                await mutations.remove.mutateAsync(row.id);
                if (activeID === row.id) select(null);
              }}
            >×</button>
          </div>
        ))}
      </div>
    </div>
  );

  return (
    <div className="grid min-h-[70vh] gap-5 lg:grid-cols-[15rem_minmax(0,1fr)]">
      <aside className="hidden rounded-2xl border border-gray-200 bg-white p-3 lg:block">{threadList}</aside>
      <section className="min-w-0">
        <header className="flex items-start justify-between gap-3">
          <div>
            <div className="flex items-center gap-2">
              <button type="button" onClick={() => setDrawerOpen(true)} className="min-h-11 rounded-lg border px-3 text-sm lg:hidden">Conversations</button>
              <h1 className="text-2xl font-semibold">Lenna</h1>
            </div>
            <p className="mt-1 text-sm text-gray-500">Read-only financial guidance grounded in deterministic Savio tools.</p>
          </div>
          <div className="inline-flex rounded-lg border border-gray-200 bg-white p-0.5 text-sm">
            {(['text', 'talk'] as const).map((value) => (
              <button key={value} type="button" disabled={value === 'talk' && !mic.supported} onClick={() => setMode(value)} className={`rounded-md px-3 py-1.5 capitalize ${mode === value ? 'bg-brand text-white' : 'text-gray-600'}`}>{value === 'text' ? 'Text' : 'Talk'}</button>
            ))}
          </div>
        </header>

        {disabled ? <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">AI is disabled. Finance features remain available.</div> : null}

        <div className="mt-6 space-y-5" aria-live="polite">
          {activeID && conversation.isPending ? <div className="rounded-2xl border bg-white p-6 text-sm text-gray-500">Loading conversation…</div> : null}
          {(conversation.data?.messages ?? []).map((message) => message.role === 'USER' ? (
            <div key={message.id} className="ml-auto max-w-[85%] rounded-2xl rounded-br-md bg-brand/10 px-4 py-3 text-sm text-gray-800">{message.content}</div>
          ) : message.response ? (
            <AssistantAnswer key={message.id} response={message.response} onSpeak={(text) => speech.speak(text)} />
          ) : null)}
          {busy ? <div className="rounded-2xl border bg-white p-5 text-sm text-gray-500">Gathering current facts…</div> : null}
          <div ref={endRef} />
        </div>

        {!conversation.data?.messages?.length ? (
          <div className="mt-5 flex flex-wrap gap-2">
            {SUGGESTIONS.map((suggestion) => <button key={suggestion} type="button" disabled={disabled || busy} onClick={() => void run(suggestion)} className="rounded-full border bg-white px-3 py-1.5 text-xs text-gray-600 hover:border-brand hover:text-brand">{suggestion}</button>)}
          </div>
        ) : null}

        {error ? <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">{error}</div> : null}
        {mic.error ? <p className="mt-2 text-sm text-red-600">{mic.error}</p> : null}

        <form className="sticky bottom-2 mt-6 flex items-end gap-2 rounded-2xl border border-gray-200 bg-white p-3 shadow-sm" onSubmit={(event) => { event.preventDefault(); void run(question); }}>
          <div className="flex-1"><TextField label="Ask Lenna" placeholder="e.g. Can I afford a 20M laptop?" value={question} maxLength={2000} disabled={disabled} onChange={(event) => setQuestion(event.target.value)} /></div>
          {mode === 'talk' && mic.supported ? <button type="button" disabled={disabled || busy} onClick={() => mic.listening ? mic.stop() : mic.start()} aria-label={mic.listening ? 'Stop listening' : 'Speak your question'} className="mb-0 h-10 w-10 rounded-full border text-brand">{mic.listening ? '■' : 'Mic'}</button> : null}
          <Button type="submit" disabled={disabled || busy || !question.trim()}>{busy ? 'Analyzing…' : 'Ask'}</Button>
        </form>
      </section>

      {drawerOpen ? (
        <div role="dialog" aria-modal="true" aria-label="Copilot conversations" className="fixed inset-0 z-50 lg:hidden">
          <button type="button" aria-label="Close conversations" onClick={() => setDrawerOpen(false)} className="absolute inset-0 bg-black/40" />
          <aside className="relative h-full w-[min(20rem,85vw)] bg-white p-4 shadow-xl">
            <button type="button" onClick={() => setDrawerOpen(false)} className="mb-4 min-h-11 text-sm font-medium">Close</button>
            {threadList}
          </aside>
        </div>
      ) : null}
    </div>
  );
}
