import { useState, useEffect, useRef } from 'react';
import type { MascotState } from '../wails.d';

export default function FloatingCompanion() {
  const [mascot, setMascot] = useState<MascotState>({
    emotion: 'idle',
    eyeSymbol: '( ◉ ◉ )',
    auraColor: '#00e5ff',
    message: 'NEXA Companion Online',
    timestamp: '',
  });
  const [toast, setToast] = useState<string | null>('Hello, sir! Ask me anything or say "Hey Nexa".');
  const idleTimerRef = useRef<any>(null);

  const resetIdleTimer = () => {
    if (idleTimerRef.current) clearTimeout(idleTimerRef.current);
    idleTimerRef.current = setTimeout(() => {
      setMascot({
        emotion: 'yawn',
        eyeSymbol: '( 💤 💤 )',
        auraColor: '#80deea',
        message: 'Yawn... Still here if you need me.',
        timestamp: '',
      });
      setToast('Yawn... Still here if you need me 💤');
    }, 180000); // 3 minutes idle
  };

  useEffect(() => {
    resetIdleTimer();

    if (window.runtime) {
      const cancelEmotion = window.runtime.EventsOn('nexa:emotion', (data: MascotState) => {
        setMascot(data);
        if (data.message) {
          setToast(data.message);
        }
        resetIdleTimer();
      });

      const cancelWake = window.runtime.EventsOn('nexa:voice:wake', (word: string) => {
        setMascot({
          emotion: 'listening',
          eyeSymbol: '( ⚡ ⚡ )',
          auraColor: '#00e676',
          message: `Activated by "${word}"!`,
          timestamp: '',
        });
        setToast(`⚡ Activated by "${word}"! Listening...`);
        resetIdleTimer();
      });

      return () => {
        if (cancelEmotion) cancelEmotion();
        if (cancelWake) cancelWake();
        if (idleTimerRef.current) clearTimeout(idleTimerRef.current);
      };
    }
  }, []);

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '12px',
        padding: '16px',
        position: 'relative',
      }}
    >
      {/* Floating Notification Speech Bubble */}
      {toast && (
        <div
          style={{
            background: 'linear-gradient(180deg, rgba(255, 255, 255, 0.45) 0%, rgba(255, 255, 255, 0.2) 100%)',
            border: '1.5px solid rgba(255, 255, 255, 0.85)',
            borderRadius: '18px 18px 18px 4px',
            padding: '10px 16px',
            color: '#ffffff',
            fontSize: '12px',
            fontWeight: 600,
            maxWidth: '240px',
            textAlign: 'center',
            boxShadow: 'inset 0 2px 4px rgba(255, 255, 255, 0.9), 0 8px 24px rgba(0, 77, 64, 0.35)',
            backdropFilter: 'blur(20px)',
            textShadow: '0 1px 3px rgba(0, 61, 51, 0.5)',
            animation: 'fadeIn 0.3s ease',
          }}
        >
          {toast}
        </div>
      )}

      {/* Solarpunk Mascot Sphere Avatar */}
      <div
        style={{
          position: 'relative',
          width: '96px',
          height: '96px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <div className="wave-ring" style={{ borderColor: mascot.auraColor }} />
        <div
          className={`orb ${mascot.emotion}`}
          style={{
            boxShadow: `inset 0 4px 8px rgba(255, 255, 255, 0.9), 0 0 35px ${mascot.auraColor}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '15px',
            fontWeight: 700,
            color: '#ffffff',
            letterSpacing: '1px',
            textShadow: '0 2px 6px rgba(0,0,0,0.5)',
          }}
        >
          {mascot.eyeSymbol}
        </div>
      </div>

      <div
        style={{
          fontSize: '11px',
          fontWeight: 700,
          color: mascot.auraColor,
          letterSpacing: '1.2px',
          textTransform: 'uppercase',
          textShadow: '0 1px 3px rgba(0,61,51,0.5)',
        }}
      >
        NEXA COMPANION • {mascot.emotion}
      </div>
    </div>
  );
}
