import { useState, useEffect } from 'react';

interface TraceStep {
  index: number;
  thought: string;
  tool: string;
  arguments: string;
  result: string;
  response: string;
  timestamp: string;
}

export default function TraceInspector() {
  const [traces, setTraces] = useState<TraceStep[]>([]);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (window.runtime) {
      const cancel = window.runtime.EventsOn('nexa:trace', (step: TraceStep) => {
        setTraces(prev => [...prev, step]);
      });
      return () => {
        if (cancel) cancel();
      };
    }
  }, []);

  const fetchTraces = async () => {
    try {
      if (window.go?.gui?.App?.GetTraces) {
        const data = await window.go.gui.App.GetTraces();
        setTraces(data || []);
      }
    } catch (e) {
      console.error('Failed to fetch traces', e);
    }
  };

  return (
    <div style={{ position: 'relative' }}>
      <button
        onClick={() => {
          setOpen(!open);
          if (!open) fetchTraces();
        }}
        style={{
          background: open
            ? 'linear-gradient(180deg, #e1bee7 0%, #ba68c8 50%, #8e24aa 100%)'
            : 'linear-gradient(180deg, rgba(255,255,255,0.35) 0%, rgba(255,255,255,0.15) 100%)',
          border: '1px solid rgba(255, 255, 255, 0.7)',
          color: '#ffffff',
          padding: '5px 12px',
          borderRadius: '20px',
          fontSize: '11px',
          cursor: 'pointer',
          fontWeight: 700,
          display: 'flex',
          alignItems: 'center',
          gap: '5px',
          boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.9), 0 2px 8px rgba(0, 0, 0, 0.15)',
          textShadow: '0 1px 2px rgba(0,0,0,0.3)',
        }}
      >
        <span>⚡</span>
        <span>TRACES ({traces.length})</span>
      </button>

      {open && (
        <div
          style={{
            position: 'absolute',
            top: '40px',
            right: 0,
            width: '430px',
            maxHeight: '490px',
            background: 'linear-gradient(135deg, rgba(255, 255, 255, 0.35) 0%, rgba(255, 255, 255, 0.15) 100%)',
            border: '1.5px solid rgba(255, 255, 255, 0.8)',
            borderRadius: '20px',
            boxShadow: 'inset 0 2px 4px rgba(255, 255, 255, 0.9), 0 16px 40px rgba(0, 77, 64, 0.5)',
            backdropFilter: 'blur(30px)',
            zIndex: 1000,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              padding: '12px 16px',
              background: 'linear-gradient(180deg, rgba(255, 255, 255, 0.3) 0%, rgba(255, 255, 255, 0.1) 100%)',
              borderBottom: '1px solid rgba(255,255,255,0.4)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <span style={{ fontSize: '12px', fontWeight: 700, color: '#ffffff', letterSpacing: '1px', textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}>
              ⚡ NEXA TRACE LOG (DEV MODE)
            </span>
            <button
              onClick={() => setTraces([])}
              style={{ background: 'none', border: 'none', color: '#ffffff', cursor: 'pointer', fontSize: '11px', opacity: 0.8 }}
            >
              Clear
            </button>
          </div>
          <div style={{ padding: '14px', overflowY: 'auto', flex: 1, display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {traces.length === 0 ? (
              <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>No agent traces recorded yet. Issue a command to see thought steps.</div>
            ) : (
              traces.map((t, idx) => (
                <div
                  key={idx}
                  style={{
                    background: 'linear-gradient(180deg, rgba(255, 255, 255, 0.2) 0%, rgba(255, 255, 255, 0.08) 100%)',
                    border: '1px solid rgba(255,255,255,0.35)',
                    borderRadius: '12px',
                    padding: '10px 12px',
                    fontSize: '11px',
                    fontFamily: "'JetBrains Mono', monospace",
                    boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.6)',
                  }}
                >
                  <div style={{ color: 'var(--text-muted)', fontSize: '10px', marginBottom: '4px' }}>
                    Step #{t.index || idx + 1} • {t.timestamp}
                  </div>
                  {t.thought && <div style={{ color: '#ffffff', marginBottom: '4px', textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}>💭 {t.thought}</div>}
                  {t.tool && (
                    <div style={{ color: 'var(--warning)', margin: '4px 0', textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}>
                      🔧 {t.tool} ({t.arguments})
                    </div>
                  )}
                  {t.result && (
                    <div style={{ color: '#a7ffeb', whiteSpace: 'pre-wrap', maxHeight: '100px', overflowY: 'auto', textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}>
                      ↳ {t.result}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
