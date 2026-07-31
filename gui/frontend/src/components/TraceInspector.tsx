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
          background: open ? 'rgba(192, 132, 252, 0.2)' : 'rgba(255, 255, 255, 0.05)',
          border: '1px solid rgba(192, 132, 252, 0.3)',
          color: 'var(--purple)',
          padding: '4px 10px',
          borderRadius: '12px',
          fontSize: '11px',
          cursor: 'pointer',
          fontWeight: 600,
          display: 'flex',
          alignItems: 'center',
          gap: '5px',
        }}
      >
        <span>⚡</span>
        <span>TRACES ({traces.length})</span>
      </button>

      {open && (
        <div
          style={{
            position: 'absolute',
            top: '36px',
            right: 0,
            width: '420px',
            maxHeight: '480px',
            background: 'rgba(15, 23, 42, 0.95)',
            border: '1px solid var(--purple)',
            borderRadius: '12px',
            boxShadow: '0 0 25px rgba(0,0,0,0.8)',
            backdropFilter: 'blur(16px)',
            zIndex: 1000,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              padding: '10px 14px',
              borderBottom: '1px solid rgba(255,255,255,0.08)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <span style={{ fontSize: '12px', fontWeight: 700, color: 'var(--purple)', letterSpacing: '1px' }}>
              NEXA TRACE LOG (DEV MODE)
            </span>
            <button
              onClick={() => setTraces([])}
              style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: '11px' }}
            >
              Clear
            </button>
          </div>
          <div style={{ padding: '12px', overflowY: 'auto', flex: 1, display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {traces.length === 0 ? (
              <div style={{ fontSize: '11px', color: 'var(--text-muted)' }}>No agent traces recorded yet. Issue a command to see thought steps.</div>
            ) : (
              traces.map((t, idx) => (
                <div
                  key={idx}
                  style={{
                    background: 'rgba(255,255,255,0.03)',
                    border: '1px solid rgba(255,255,255,0.05)',
                    borderRadius: '8px',
                    padding: '8px 10px',
                    fontSize: '11px',
                    fontFamily: "'JetBrains Mono', monospace",
                  }}
                >
                  <div style={{ color: 'var(--text-muted)', fontSize: '10px', marginBottom: '4px' }}>
                    Step #{t.index || idx + 1} • {t.timestamp}
                  </div>
                  {t.thought && <div style={{ color: '#fff', marginBottom: '4px' }}>💭 {t.thought}</div>}
                  {t.tool && (
                    <div style={{ color: 'var(--warning)', margin: '4px 0' }}>
                      🔧 {t.tool} ({t.arguments})
                    </div>
                  )}
                  {t.result && (
                    <div style={{ color: 'var(--success)', whiteSpace: 'pre-wrap', maxHeight: '100px', overflowY: 'auto' }}>
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
