import { useState, useEffect } from 'react';
import ReactDOM from 'react-dom';

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

  const modalJSX = open ? (
    <div
      onClick={() => setOpen(false)}
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100vw',
        height: '100vh',
        background: 'rgba(0, 38, 32, 0.75)',
        backdropFilter: 'blur(16px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 99999,
        WebkitAppRegion: 'no-drag',
      } as React.CSSProperties}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          background: 'linear-gradient(135deg, rgba(255, 255, 255, 0.35) 0%, rgba(255, 255, 255, 0.15) 100%)',
          border: '1.5px solid rgba(255, 255, 255, 0.8)',
          borderRadius: '24px',
          width: '560px',
          maxWidth: '90vw',
          maxHeight: '85vh',
          padding: '24px',
          display: 'flex',
          flexDirection: 'column',
          gap: '16px',
          boxShadow: 'inset 0 2px 4px rgba(255, 255, 255, 0.9), 0 20px 50px rgba(0, 77, 64, 0.6)',
          backdropFilter: 'blur(30px)',
          color: '#ffffff',
        }}
      >
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span style={{ fontSize: '18px' }}>⚡</span>
            <span
              style={{
                fontSize: '15px',
                fontWeight: 800,
                letterSpacing: '0.5px',
                textShadow: '0 1px 3px rgba(0,0,0,0.4)',
              }}
            >
              NEXA Trace Log (Dev Mode)
            </span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <button
              onClick={() => setTraces([])}
              style={{
                background: 'rgba(255, 255, 255, 0.2)',
                border: '1px solid rgba(255, 255, 255, 0.5)',
                color: '#ffffff',
                padding: '4px 10px',
                borderRadius: '12px',
                fontSize: '11px',
                cursor: 'pointer',
                fontWeight: 600,
              }}
            >
              Clear Log
            </button>
            <button
              onClick={() => setOpen(false)}
              style={{
                background: 'rgba(255, 255, 255, 0.25)',
                border: '1px solid rgba(255, 255, 255, 0.6)',
                color: '#ffffff',
                width: '28px',
                height: '28px',
                borderRadius: '50%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                fontSize: '14px',
                fontWeight: 700,
              }}
            >
              ✕
            </button>
          </div>
        </div>

        {/* Content List */}
        <div
          style={{
            overflowY: 'auto',
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            gap: '12px',
            paddingRight: '4px',
          }}
        >
          {traces.length === 0 ? (
            <div
              style={{
                textAlign: 'center',
                padding: '40px 20px',
                color: 'rgba(255,255,255,0.7)',
                fontSize: '13px',
              }}
            >
              No agent trace steps recorded yet. Ask NEXA a prompt or run a tool command to inspect execution traces.
            </div>
          ) : (
            traces.map((t, idx) => (
              <div
                key={idx}
                style={{
                  background: 'linear-gradient(180deg, rgba(255, 255, 255, 0.25) 0%, rgba(255, 255, 255, 0.1) 100%)',
                  border: '1px solid rgba(255, 255, 255, 0.5)',
                  borderRadius: '16px',
                  padding: '14px',
                  boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.7), 0 4px 12px rgba(0,0,0,0.15)',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '8px',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    fontSize: '11px',
                    fontWeight: 700,
                    color: 'rgba(255,255,255,0.85)',
                  }}
                >
                  <span
                    style={{
                      background: 'rgba(0, 229, 255, 0.3)',
                      padding: '2px 8px',
                      borderRadius: '10px',
                      border: '1px solid rgba(0, 229, 255, 0.5)',
                    }}
                  >
                    Step #{t.index || idx + 1}
                  </span>
                  <span>{t.timestamp}</span>
                </div>

                {t.thought && (
                  <div
                    style={{
                      fontSize: '12px',
                      color: '#ffffff',
                      lineHeight: '1.4',
                      textShadow: '0 1px 2px rgba(0,0,0,0.3)',
                    }}
                  >
                    💭 <span style={{ fontWeight: 600 }}>Thought:</span> {t.thought}
                  </div>
                )}

                {t.tool && (
                  <div
                    style={{
                      fontSize: '12px',
                      color: '#ffd740',
                      fontWeight: 600,
                      background: 'rgba(0,0,0,0.2)',
                      padding: '8px 12px',
                      borderRadius: '10px',
                      border: '1px solid rgba(255, 215, 64, 0.4)',
                      fontFamily: "'JetBrains Mono', monospace",
                    }}
                  >
                    <div>🔧 Tool: {t.tool}</div>
                    {t.arguments && (
                      <div
                        style={{
                          fontSize: '11px',
                          color: 'rgba(255,255,255,0.85)',
                          marginTop: '4px',
                          whiteSpace: 'pre-wrap',
                          wordBreak: 'break-word',
                        }}
                      >
                        Args: {t.arguments}
                      </div>
                    )}
                  </div>
                )}

                {t.result && (
                  <div
                    style={{
                      fontSize: '11px',
                      color: '#a7ffeb',
                      background: 'rgba(0,0,0,0.25)',
                      padding: '8px 12px',
                      borderRadius: '10px',
                      border: '1px solid rgba(167, 255, 235, 0.3)',
                      fontFamily: "'JetBrains Mono', monospace",
                      whiteSpace: 'pre-wrap',
                      maxHeight: '120px',
                      overflowY: 'auto',
                      wordBreak: 'break-word',
                    }}
                  >
                    ↳ Result: {t.result}
                  </div>
                )}

                {t.response && (
                  <div
                    style={{
                      fontSize: '12px',
                      color: '#ffffff',
                      background: 'rgba(255,255,255,0.15)',
                      padding: '8px 12px',
                      borderRadius: '10px',
                      lineHeight: '1.4',
                    }}
                  >
                    💬 Response: {t.response}
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  ) : null;

  return (
    <>
      <button
        onClick={() => {
          setOpen(true);
          fetchTraces();
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

      {modalJSX && ReactDOM.createPortal(modalJSX, document.body)}
    </>
  );
}
