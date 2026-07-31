import { useState, useEffect } from 'react';

type OrbState = 'idle' | 'listening' | 'thinking' | 'speaking';

export default function VoiceReactor() {
  const [orbState, setOrbState] = useState<OrbState>('idle');
  const [voiceActive, setVoiceActive] = useState(false);
  const [statusMsg, setStatusMsg] = useState('Voice Engine Offline');

  useEffect(() => {
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
            background: voiceActive
              ? 'linear-gradient(180deg, #a7ffeb 0%, #00e676 50%, #00c853 100%)'
              : 'linear-gradient(180deg, rgba(255,255,255,0.3) 0%, rgba(255,255,255,0.1) 100%)',
            border: '1px solid rgba(255, 255, 255, 0.8)',
            color: '#ffffff',
            padding: '5px 12px',
            borderRadius: '20px',
            fontSize: '11px',
            cursor: 'pointer',
            fontWeight: 700,
            display: 'flex',
            alignItems: 'center',
            gap: '6px',
            boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.9), 0 2px 8px rgba(0, 0, 0, 0.15)',
            textShadow: '0 1px 2px rgba(0,0,0,0.3)',
          }}
        >
          <span>🎤</span>
          <span>{voiceActive ? 'VOICE ON' : 'START VOICE'}</span>
        </button>
      </div>
      <div className="reactor-box">
        <div className="orb-container">
          <div className="wave-ring" />
          <div
            className={`orb ${
              orbState === 'thinking' ? 'thinking' : orbState === 'listening' ? 'listening' : ''
            }`}
          />
        </div>
        <div className="reactor-label">{statusMsg}</div>
      </div>
    </>
  );
}
