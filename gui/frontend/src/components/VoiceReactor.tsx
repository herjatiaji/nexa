import { useState, useEffect } from 'react';

type OrbState = 'idle' | 'listening' | 'thinking' | 'speaking';

export default function VoiceReactor() {
  const [orbState, setOrbState] = useState<OrbState>('idle');
  const [voiceActive, setVoiceActive] = useState(false);
  const [statusMsg, setStatusMsg] = useState('Voice Engine Offline');

  useEffect(() => {
    // Fetch initial voice status
    window.go.gui.App.GetStatus().then(status => {
      if (status.voiceActive) {
        setVoiceActive(true);
        setStatusMsg('Listening for "Hey Nexa"...');
      }
    });

    if (window.runtime) {
      const cancelStatus = window.runtime.EventsOn('nexa:status', (status: string) => {
        if (status === 'thinking') setOrbState('thinking');
        else if (status === 'listening') setOrbState('listening');
        else if (status === 'speaking') setOrbState('speaking');
        else setOrbState('idle');
      });

      const cancelVoiceState = window.runtime.EventsOn('nexa:voice:state', (active: boolean) => {
        setVoiceActive(active);
        if (active) setStatusMsg('Listening for "Hey Nexa"...');
        else setStatusMsg('Voice Engine Offline');
      });

      const cancelWake = window.runtime.EventsOn('nexa:voice:wake', (word: string) => {
        setOrbState('listening');
        setStatusMsg(`⚡ Activated by "${word}"!`);
      });

      return () => {
        if (cancelStatus) cancelStatus();
        if (cancelVoiceState) cancelVoiceState();
        if (cancelWake) cancelWake();
      };
    }
  }, []);

  const toggleVoice = async () => {
    if (voiceActive) {
      setStatusMsg('Stopping voice engine...');
      await window.go.gui.App.StopVoiceEngine();
      setVoiceActive(false);
      setStatusMsg('Voice Engine Offline');
    } else {
      setStatusMsg('Starting voice engine...');
      const res = await window.go.gui.App.StartVoiceEngine();
      if (res === 'OK' || res === 'Voice engine already active') {
        setVoiceActive(true);
        setStatusMsg('Listening for "Hey Nexa"...');
      } else {
        setStatusMsg(res);
      }
    }
  };

  return (
    <>
      <div className="card-header">
        <span className="card-title">Voice Reactor</span>
        <button
          onClick={toggleVoice}
          style={{
            background: voiceActive ? 'rgba(52, 211, 153, 0.15)' : 'rgba(255, 255, 255, 0.05)',
            border: `1px solid ${voiceActive ? '#34d399' : 'rgba(255, 255, 255, 0.1)'}`,
            color: voiceActive ? '#34d399' : 'var(--text-muted)',
            padding: '4px 10px',
            borderRadius: '12px',
            fontSize: '11px',
            cursor: 'pointer',
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: '5px'
          }}
        >
          <span>🎤</span>
          <span>{voiceActive ? 'VOICE ON' : 'START VOICE'}</span>
        </button>
      </div>
      <div className="reactor-box">
        <div className="orb-container">
          <div className="wave-ring" />
          <div className={`orb ${orbState === 'thinking' ? 'thinking' : orbState === 'listening' ? 'listening' : ''}`} />
        </div>
        <div className="reactor-label" style={{ color: voiceActive ? 'var(--cyan)' : 'var(--text-dim)' }}>
          {statusMsg}
        </div>
      </div>
    </>
  );
}
