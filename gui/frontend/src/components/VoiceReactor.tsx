import { useState, useEffect } from 'react';
import FloatingCompanion from './FloatingCompanion';

export default function VoiceReactor() {
  const [voiceActive, setVoiceActive] = useState(false);

  useEffect(() => {
    window.go.gui.App.GetStatus().then(status => {
      if (status.voiceActive) {
        setVoiceActive(true);
      }
    });

    if (window.runtime) {
      const cancelVoiceState = window.runtime.EventsOn('nexa:voice:state', (active: boolean) => {
        setVoiceActive(active);
      });

      return () => {
        if (cancelVoiceState) cancelVoiceState();
      };
    }
  }, []);

  const toggleVoice = async () => {
    if (voiceActive) {
      await window.go.gui.App.StopVoiceEngine();
      setVoiceActive(false);
    } else {
      const res = await window.go.gui.App.StartVoiceEngine();
      if (res === 'OK' || res === 'Voice engine already active') {
        setVoiceActive(true);
      }
    }
  };

  return (
    <>
      <div className="card-header">
        <span className="card-title">AI Mascot Companion</span>
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
      <FloatingCompanion />
    </>
  );
}
