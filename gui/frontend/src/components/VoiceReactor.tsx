import { useState, useEffect } from 'react';

type OrbState = 'idle' | 'thinking' | 'speaking';

export default function VoiceReactor() {
  const [state, setState] = useState<OrbState>('idle');

  useEffect(() => {
    if (window.runtime) {
      const cancel = window.runtime.EventsOn('nexa:status', (status: string) => {
        if (status === 'thinking') setState('thinking');
        else if (status === 'speaking') setState('speaking');
        else setState('idle');
      });
      return () => { if (cancel) cancel(); };
    }
  }, []);

  const label = state === 'thinking' ? 'Thinking...'
    : state === 'speaking' ? 'Speaking...'
    : 'Ready';

  return (
    <>
      <div className="card-header">
        <span className="card-title">Voice Reactor</span>
      </div>
      <div className="reactor-box">
        <div className="orb-container">
          <div className="wave-ring" />
          <div className={`orb ${state === 'thinking' ? 'thinking' : ''}`} />
        </div>
        <div className="reactor-label">{label}</div>
      </div>
    </>
  );
}
