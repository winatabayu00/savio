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
  const [showFacts, setShowFacts] = useState(false);
  return (
    <div className="card shadow-sm">
      <div className="card-body">
        <div className="d-flex align-items-center justify-content-between gap-2">
          <span className="badge bg-soft-primary text-primary">Lenna · grounded answer</span>
          {response.facts.length > 0 ? (
            <button type="button" onClick={() => setShowFacts((v) => !v)} aria-expanded={showFacts} className="btn btn-link text-muted p-0 fs-12 fw-medium">
              {showFacts ? 'Hide supporting facts' : 'Show supporting facts'}
            </button>
          ) : null}
        </div>
        {showFacts && response.facts.length > 0 ? (
          <div className="mt-3">
            <h3 className="fs-12 text-uppercase fw-semibold text-muted">Supporting facts</h3>
            <ul className="mt-2 list-unstyled d-flex flex-column gap-2 mb-0">
              {response.facts.map((fact, index) => (
                <li key={`${fact.label}-${index}`} className="d-flex flex-column gap-1 bg-light rounded-3 p-2 fs-13 flex-sm-row justify-content-sm-between">
                  <span className="text-secondary">{fact.label}</span>
                  <strong>{fact.value}</strong>
                </li>
              ))}
            </ul>
          </div>
        ) : null}
        <div className="mt-3 border-top pt-3">
          <div className="d-flex justify-content-between gap-3">
            <h3 className="fs-12 text-uppercase fw-semibold text-muted mb-0">Answer</h3>
            <button type="button" onClick={() => onSpeak(response.answer)} className="btn btn-link text-primary p-0 fs-12 fw-medium" aria-label="Read answer aloud">Listen</button>
          </div>
          <p className="mt-2 text-dark mb-0">{response.answer}</p>
        </div>
        {response.clarification ? <p className="mt-3 alert alert-primary p-3 fs-13 mb-0">Need more detail: {response.clarification}</p> : null}
        {response.actions.length > 0 ? (
          <div className="mt-3 fs-13 text-secondary">
            <h4 className="fs-12 text-uppercase fw-semibold text-muted">Suggested next steps</h4>
            <ul className="mt-1 mb-0 ps-3">{response.actions.map((action) => <li key={action}>{action}</li>)}</ul>
          </div>
        ) : null}
      </div>
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
    <div className="d-flex flex-column h-100">
      <Button type="button" onClick={() => select(null)}>New conversation</Button>
      <div className="mt-3 flex-grow-1 d-flex flex-column gap-1 overflow-auto">
        {conversations.isPending ? <p className="p-2 fs-13 text-muted mb-0">Loading conversations…</p> : null}
        {conversations.data?.length === 0 ? <p className="p-2 fs-13 text-muted mb-0">No conversations yet.</p> : null}
        {conversations.data?.map((row) => (
          <div key={row.id} className={`d-flex rounded-3 ${activeID === row.id ? 'bg-soft-primary' : ''}`}>
            <button type="button" onClick={() => select(row.id)} className="btn btn-link text-dark text-start text-truncate flex-grow-1 px-3 fs-13 fw-medium">
              {row.title ?? 'New conversation'}
            </button>
            <button
              type="button"
              aria-label={`Delete ${row.title ?? 'conversation'}`}
              className="btn btn-link text-muted px-3"
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
    <div className="row g-4 mt-1">
      <aside className="col-lg-3 d-none d-lg-block">
        <div className="card h-100">
          <div className="card-body p-2">{threadList}</div>
        </div>
      </aside>
      <section className="col-lg-9">
        <header className="d-flex align-items-start justify-content-between gap-3">
          <div>
            <div className="d-flex align-items-center gap-2">
              <button type="button" onClick={() => setDrawerOpen(true)} className="btn btn-outline-secondary d-lg-none fs-13">Conversations</button>
              <h1 className="fs-20 fw-bolder mb-0">Lenna</h1>
            </div>
            <p className="mt-1 fs-13 text-muted mb-0">Read-only financial guidance grounded in deterministic Savio tools.</p>
          </div>
          <div className="d-inline-flex border rounded-3 bg-white p-1 fs-13">
            {(['text', 'talk'] as const).map((value) => (
              <button key={value} type="button" disabled={value === 'talk' && !mic.supported} onClick={() => setMode(value)} className={`rounded-2 px-2 py-1 text-capitalize ${mode === value ? 'bg-primary text-white' : 'text-secondary'}`}>{value === 'text' ? 'Text' : 'Talk'}</button>
            ))}
          </div>
        </header>

        {disabled ? <div className="mt-4 alert alert-warning p-3 fs-13 mb-0">AI is disabled. Finance features remain available.</div> : null}

        <div className="mt-4 d-flex flex-column gap-3" aria-live="polite">
          {activeID && conversation.isPending ? <div className="card"><div className="card-body p-4 fs-13 text-muted">Loading conversation…</div></div> : null}
          {(conversation.data?.messages ?? []).map((message) => message.role === 'USER' ? (
            <div key={message.id} className="ms-auto align-self-end rounded-4 bg-soft-primary px-3 py-2 fs-13 text-dark" style={{ maxWidth: '85%' }}>{message.content}</div>
          ) : message.response ? (
            <AssistantAnswer key={message.id} response={message.response} onSpeak={(text) => speech.speak(text)} />
          ) : null)}
          {busy ? <div className="card"><div className="card-body p-3 fs-13 text-muted">Gathering current facts…</div></div> : null}
          <div ref={endRef} />
        </div>

        {!conversation.data?.messages?.length ? (
          <div className="mt-4 d-flex flex-wrap gap-2">
            {SUGGESTIONS.map((suggestion) => <button key={suggestion} type="button" disabled={disabled || busy} onClick={() => void run(suggestion)} className="btn btn-outline-primary rounded-pill fs-12">{suggestion}</button>)}
          </div>
        ) : null}

        {error ? <div className="mt-4 alert alert-warning p-3 fs-13 mb-0">{error}</div> : null}
        {mic.error ? <p className="mt-2 fs-13 text-danger mb-0">{mic.error}</p> : null}

        <form className="sticky-bottom mt-4 d-flex align-items-end gap-2 border rounded-3 bg-white p-2 shadow-sm" style={{ bottom: '0.5rem' }} onSubmit={(event) => { event.preventDefault(); void run(question); }}>
          <div className="flex-grow-1"><TextField label="Ask Lenna" placeholder="e.g. Can I afford a 20M laptop?" value={question} maxLength={2000} disabled={disabled} onChange={(event) => setQuestion(event.target.value)} /></div>
          {mode === 'talk' && mic.supported ? <button type="button" disabled={disabled || busy} onClick={() => mic.listening ? mic.stop() : mic.start()} aria-label={mic.listening ? 'Stop listening' : 'Speak your question'} className="btn btn-outline-primary rounded-circle d-flex align-items-center justify-content-center flex-shrink-0" style={{ width: 40, height: 40 }}>{mic.listening ? '■' : 'Mic'}</button> : null}
          <Button type="submit" disabled={disabled || busy || !question.trim()}>{busy ? 'Analyzing…' : 'Ask'}</Button>
        </form>
      </section>

      {drawerOpen ? (
        <div role="dialog" aria-modal="true" aria-label="Copilot conversations" className="position-fixed top-0 start-0 w-100 h-100 d-lg-none" style={{ zIndex: 1050 }}>
          <button type="button" aria-label="Close conversations" onClick={() => setDrawerOpen(false)} className="position-absolute top-0 start-0 w-100 h-100 border-0" style={{ background: 'rgba(0,0,0,0.4)' }} />
          <aside className="position-relative h-100 bg-white p-3 shadow-lg" style={{ width: 'min(20rem, 85vw)' }}>
            <button type="button" onClick={() => setDrawerOpen(false)} className="btn btn-link fs-13 fw-medium p-0 mb-3">Close</button>
            {threadList}
          </aside>
        </div>
      ) : null}
    </div>
  );
}
