import { useState, useEffect, useRef } from 'react';
import type { MascotState } from '../wails.d';

interface WindowInfo {
  title: string;
  appName: string;
  timestamp: string;
}

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

      // Desktop Observer Event Listener
      const cancelWindowChanged = window.runtime.EventsOn('window.changed', (info: WindowInfo) => {
        if (info.appName) {
          setToast(`Focused: ${info.appName}`);
          // Flash mascot aura subtly on desktop app switch
          setMascot(prev => ({
            ...prev,
            message: `Focused on ${info.appName}`,
          }));
        }
        resetIdleTimer();
      });

      return () => {
        if (cancelEmotion) cancelEmotion();
        if (cancelWake) cancelWake();
        if (cancelWindowChanged) cancelWindowChanged();
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
        justifyContent: 'center',
        padding: '16px',
        position: 'relative',
        userSelect: 'none',
        WebkitAppRegion: 'drag', // Click and drag floating avatar window anywhere on screen
      } as React.CSSProperties}
    >
      {/* Floating Toast Speech Bubble */}
      {toast && (
        <div
          style={{
            position: 'absolute',
            top: '-32px',
            background: 'linear-gradient(135deg, rgba(255, 255, 255, 0.4) 0%, rgba(255, 255, 255, 0.15) 100%)',
            border: '1.5px solid rgba(255, 255, 255, 0.8)',
            borderRadius: '18px',
            padding: '6px 14px',
            fontSize: '11px',
            fontWeight: 700,
            color: '#ffffff',
            boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.9), 0 8px 24px rgba(0, 77, 64, 0.3)',
            backdropFilter: 'blur(20px)',
            maxWidth: '240px',
            textAlign: 'center',
            zIndex: 10,
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            animation: 'fadeIn 0.3s ease-out',
            textShadow: '0 1px 3px rgba(0, 0, 0, 0.4)',
            WebkitAppRegion: 'no-drag',
          } as React.CSSProperties}
        >
          {toast}
        </div>
      )}

      {/* Water Droplet Liquid Glass Mascot Sphere */}
      <div
        style={{
          width: '130px',
          height: '130px',
          borderRadius: '50%',
          background: `radial-gradient(circle at 35% 35%, rgba(255, 255, 255, 0.9) 0%, ${mascot.auraColor}44 40%, ${mascot.auraColor}cc 85%, ${mascot.auraColor} 100%)`,
          border: '2px solid rgba(255, 255, 255, 0.85)',
          boxShadow: `inset 0 4px 12px rgba(255, 255, 255, 0.9), inset 0 -6px 16px ${mascot.auraColor}, 0 0 35px ${mascot.auraColor}88, 0 12px 30px rgba(0, 77, 64, 0.4)`,
          backdropFilter: 'blur(25px)',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          cursor: 'grab',
          transition: 'all 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275)',
        }}
      >
        {/* Specular Light Reflection Highlight */}
        <div
          style={{
            position: 'absolute',
            top: '12%',
            left: '20%',
            width: '40px',
            height: '20px',
            borderRadius: '50%',
            background: 'linear-gradient(180deg, rgba(255, 255, 255, 0.85) 0%, rgba(255, 255, 255, 0.1) 100%)',
            transform: 'rotate(-25deg)',
            pointerEvents: 'none',
          }}
        />

        {/* Animated Face Expressions */}
        <div
          style={{
            fontSize: mascot.emotion === 'yawn' ? '20px' : '22px',
            fontWeight: 900,
            color: '#ffffff',
            letterSpacing: '2px',
            textShadow: '0 2px 6px rgba(0, 0, 0, 0.5), 0 0 12px rgba(255, 255, 255, 0.8)',
            zIndex: 2,
            fontFamily: "'JetBrains Mono', monospace",
          }}
        >
          {mascot.eyeSymbol}
        </div>

        {/* Mascot Name Badge */}
        <div
          style={{
            fontSize: '9px',
            fontWeight: 800,
            color: 'rgba(255, 255, 255, 0.9)',
            textTransform: 'uppercase',
            letterSpacing: '1.5px',
            marginTop: '4px',
            textShadow: '0 1px 2px rgba(0,0,0,0.5)',
            zIndex: 2,
          }}
        >
          NEXA
        </div>
      </div>
    </div>
  );
}
