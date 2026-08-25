import { useCallback, useEffect, useRef, useState } from 'react';

function getRecognition(): (typeof SpeechRecognition) | null {
  const w = window as unknown as Record<string, unknown>;
  return (w.SpeechRecognition as typeof SpeechRecognition) || (w.webkitSpeechRecognition as typeof SpeechRecognition) || null;
}

function getSpeech(): SpeechSynthesis | null {
  return 'speechSynthesis' in window ? window.speechSynthesis : null;
}

export function speechSupported(): boolean {
  return getRecognition() !== null;
}

export function ttsSupported(): boolean {
  return getSpeech() !== null;
}

export function useSpeechRecognition(onFinal: (text: string) => void) {
  const [listening, setListening] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const recRef = useRef<SpeechRecognition | null>(null);
  const cbRef = useRef(onFinal);
  cbRef.current = onFinal;

  const stop = useCallback(() => {
    const rec = recRef.current;
    if (!rec) return;
    rec.onresult = null;
    rec.onerror = null;
    rec.onend = null;
    rec.stop();
    recRef.current = null;
    setListening(false);
  }, []);

  const start = useCallback(() => {
    const Ctor = getRecognition();
    if (!Ctor) {
      setError('Mic is not supported in this browser.');
      return;
    }
    stop();
    const rec = new Ctor();
    rec.lang = 'id-ID';
    rec.interimResults = false;
    rec.maxAlternatives = 1;
    rec.onresult = (event: SpeechRecognitionEvent) => {
      const text = event.results?.[0]?.[0]?.transcript?.trim() ?? '';
      if (text) cbRef.current(text);
    };
    rec.onerror = (event: SpeechRecognitionErrorEvent) => {
      setError(event?.error === 'not-allowed' ? 'Microphone permission denied.' : 'Could not hear you. Try again.');
    };
    rec.onend = () => {
      recRef.current = null;
      setListening(false);
    };
    recRef.current = rec;
    setError(null);
    setListening(true);
    rec.start();
  }, [stop]);

  useEffect(() => () => stop(), [stop]);

  return { supported: speechSupported(), listening, error, start, stop };
}

const ONES = ['', 'satu', 'dua', 'tiga', 'empat', 'lima', 'enam', 'tujuh', 'delapan', 'sembilan'];
const TEENS = ['sepuluh', 'sebelas', 'dua belas', 'tiga belas', 'empat belas', 'lima belas', 'enam belas', 'tujuh belas', 'delapan belas', 'sembilan belas'];
const SCALES = ['', 'ribu', 'juta', 'miliar', 'triliun'];

function under1000(n: number): string {
  let s = '';
  const h = Math.floor(n / 100);
  const r = n % 100;
  if (h) s += (h === 1 ? 'seratus' : `${ONES[h]} ratus`);
  if (r) {
    if (s) s += ' ';
    if (r < 10) s += ONES[r];
    else if (r < 20) s += TEENS[r - 10];
    else {
      const t = Math.floor(r / 10);
      const o = r % 10;
      s += t === 1 ? 'sepuluh' : `${ONES[t]} puluh`;
      if (o) s += ` ${ONES[o]}`;
    }
  }
  return s || 'nol';
}

function toWords(n: number): string {
  if (n === 0) return 'nol';
  const parts: string[] = [];
  let group = 0;
  while (n > 0) {
    const g = n % 1000;
    if (g) {
      const scale = SCALES[group];
      // "1.000" reads "seribu", not "satu ribu"
      let words = under1000(g);
      if (scale === 'ribu' && g === 1) words = 'seribu';
      else if (scale === 'juta' && g === 1) words = 'satu juta';
      else if (scale) words += ` ${scale}`;
      parts.unshift(words);
    }
    n = Math.floor(n / 1000);
    group++;
  }
  return parts.join(' ');
}

// TTS in browsers without an Indonesian voice spells "10000" digit-by-digit
// ("satu nol nol…"). Expand numbers to Indonesian words so the read-aloud is
// natural on every engine (ponytail: supports integers up to trillions).
function expandNumber(raw: string): string {
  const trimmed = raw.replace(/[.,]$/, '');
  const parts = trimmed.split(/[.,]/);
  const hasDecimal = parts.length > 1 && parts[parts.length - 1].length !== 3;
  const ints = hasDecimal ? parts.slice(0, -1) : parts;
  if (ints.length > 1 && ints.every((p) => p.length === 3)) {
    return toWords(Number(ints.join('')));
  }
  const clean = ints.join('');
  if (!/^\d+$/.test(clean)) return raw;
  const n = Number(clean);
  if (!Number.isSafeInteger(n)) return raw;
  return toWords(n);
}

export function numberToWordsID(text: string): string {
  return text
    .replace(/Rp/gi, 'rupiah ')
    .replace(/\d[\d.,]*/g, expandNumber)
    .replace(/\s+/g, ' ')
    .trim();
}

export function useSpeechSynthesis() {
  const [speaking, setSpeaking] = useState(false);
  const current = useRef<SpeechSynthesisUtterance | null>(null);

  const stopSpeaking = useCallback(() => {
    const synth = getSpeech();
    if (synth) synth.cancel();
    current.current = null;
    setSpeaking(false);
  }, []);

  const speak = useCallback(
    (text: string, onEnd?: () => void) => {
      const synth = getSpeech();
      stopSpeaking();
      if (!synth || !text) return;
      const utterance = new SpeechSynthesisUtterance(numberToWordsID(text));
      utterance.lang = 'id-ID';
      utterance.rate = 1;
      const voices = synth.getVoices();
      const idVoice = voices.find((v) => v.lang.toLowerCase().startsWith('id'));
      if (idVoice) utterance.voice = idVoice;
      utterance.onend = () => {
        current.current = null;
        setSpeaking(false);
        onEnd?.();
      };
      utterance.onerror = () => {
        current.current = null;
        setSpeaking(false);
      };
      current.current = utterance;
      setSpeaking(true);
      synth.speak(utterance);
    },
    [stopSpeaking],
  );

  useEffect(() => () => stopSpeaking(), [stopSpeaking]);

  return { supported: ttsSupported(), speaking, speak, stopSpeaking };
}