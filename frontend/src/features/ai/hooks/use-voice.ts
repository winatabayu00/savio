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
      const utterance = new SpeechSynthesisUtterance(text);
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