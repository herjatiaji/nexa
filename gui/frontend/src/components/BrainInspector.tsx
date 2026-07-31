import { useState, useEffect } from 'react';
import ReactDOM from 'react-dom';
import Timeline from './Timeline';
import DecisionCard from './DecisionCard';
import type { BrainSnapshot, CognitiveTrace, BrainMetricsTelemetry, DecisionExplanation } from '../wails.d';

export default function BrainInspector() {
  const [isOpen, setIsOpen] = useState(false);
  const [brainState, setBrainState] = useState<BrainSnapshot | null>(null);
  const [traces, setTraces] = useState<CognitiveTrace[]>([]);
  const [metrics, setMetrics] = useState<BrainMetricsTelemetry | null>(null);
  const [explanation, setExplanation] = useState<DecisionExplanation | null>(null);

  useEffect(() => {
    if (!isOpen) return;

    const fetchData = async () => {
      try {
        if (window.go?.gui?.App) {
          const bs = await window.go.gui.App.GetBrainState();
          const trs = await window.go.gui.App.GetCognitiveTraces(50);
          const mtr = await window.go.gui.App.GetBrainMetrics();
          const exp = await window.go.gui.App.ExplainLastDecision();

          setBrainState(bs);
          setTraces(trs || []);
          setMetrics(mtr);
          setExplanation(exp);
        }
      } catch (err) {
        console.error('Failed to fetch brain observability data', err);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 1500);
    return () => clearInterval(interval);
  }, [isOpen]);

  const modalJSX = (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(2, 10, 15, 0.82)',
        backdropFilter: 'blur(12px)',
        zIndex: 99999,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px',
        ['WebkitAppRegion' as any]: 'no-drag',
      }}
      onClick={() => setIsOpen(false)}
    >
      <div
        style={{
          width: '920px',
          maxHeight: '85vh',
          background: 'linear-gradient(145deg, rgba(8, 28, 36, 0.95) 0%, rgba(4, 16, 22, 0.98) 100%)',
          borderRadius: '24px',
          border: '1px solid rgba(0, 229, 255, 0.3)',
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.7), inset 0 1px 2px rgba(255, 255, 255, 0.2)',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div
          style={{
            padding: '16px 24px',
            borderBottom: '1px solid rgba(0, 229, 255, 0.15)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            background: 'rgba(0, 229, 255, 0.04)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <span style={{ fontSize: '20px' }}>🧠</span>
            <span style={{ color: '#00e5ff', fontWeight: 700, fontSize: '15px', letterSpacing: '0.5px' }}>
              NEXA BRAIN INSPECTOR & TELEMETRY
            </span>
          </div>
          <button
            onClick={() => setIsOpen(false)}
            style={{
              background: 'rgba(255, 255, 255, 0.1)',
              border: 'none',
              color: '#fff',
              borderRadius: '50%',
              width: '28px',
              height: '28px',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 'bold',
            }}
          >
            ✕
          </button>
        </div>

        {/* Content Body */}
        <div style={{ padding: '20px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '18px' }}>
          {/* Top Status Cards */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px' }}>
            <div style={{ background: 'rgba(0, 40, 50, 0.4)', padding: '12px', borderRadius: '12px', border: '1px solid rgba(0, 229, 255, 0.15)' }}>
              <div style={{ fontSize: '11px', color: 'rgba(255,255,255,0.5)' }}>Focused App</div>
              <div style={{ fontSize: '13px', color: '#00e5ff', fontWeight: 'bold', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', marginTop: '2px' }}>
                {brainState?.Context?.focused_app || 'None'}
              </div>
            </div>

            <div style={{ background: 'rgba(0, 40, 50, 0.4)', padding: '12px', borderRadius: '12px', border: '1px solid rgba(0, 229, 255, 0.15)' }}>
              <div style={{ fontSize: '11px', color: 'rgba(255,255,255,0.5)' }}>Activity State</div>
              <div style={{ fontSize: '13px', color: '#76ff03', fontWeight: 'bold', marginTop: '2px' }}>
                {brainState?.Context?.activity || 'idle'}
              </div>
            </div>

            <div style={{ background: 'rgba(0, 40, 50, 0.4)', padding: '12px', borderRadius: '12px', border: '1px solid rgba(0, 229, 255, 0.15)' }}>
              <div style={{ fontSize: '11px', color: 'rgba(255,255,255,0.5)' }}>Attention Score</div>
              <div style={{ fontSize: '13px', color: '#ffd600', fontWeight: 'bold', marginTop: '2px' }}>
                {((brainState?.Social?.attention_score || 0.5) * 100).toFixed(0)}%
              </div>
            </div>

            <div style={{ background: 'rgba(0, 40, 50, 0.4)', padding: '12px', borderRadius: '12px', border: '1px solid rgba(0, 229, 255, 0.15)' }}>
              <div style={{ fontSize: '11px', color: 'rgba(255,255,255,0.5)' }}>Avg Tick Latency</div>
              <div style={{ fontSize: '13px', color: '#ff4081', fontWeight: 'bold', marginTop: '2px' }}>
                {(metrics?.avg_tick_latency_ms || 0).toFixed(1)} ms
              </div>
            </div>
          </div>

          {/* Decision Explanation Card */}
          <div>
            <div style={{ fontSize: '12px', color: 'rgba(255,255,255,0.5)', textTransform: 'uppercase', marginBottom: '8px', letterSpacing: '0.5px' }}>
              💬 Latest Decision Reasoning
            </div>
            <DecisionCard explanation={explanation} />
          </div>

          {/* Cognitive Timeline */}
          <div>
            <div style={{ fontSize: '12px', color: 'rgba(255,255,255,0.5)', textTransform: 'uppercase', marginBottom: '8px', letterSpacing: '0.5px' }}>
              📜 Real-Time Cognitive Trace Stream
            </div>
            <Timeline traces={traces} />
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        style={{
          background: 'linear-gradient(135deg, rgba(0, 229, 255, 0.2) 0%, rgba(0, 150, 180, 0.3) 100%)',
          border: '1px solid rgba(0, 229, 255, 0.4)',
          borderRadius: '16px',
          color: '#00e5ff',
          padding: '4px 10px',
          fontSize: '12px',
          fontWeight: 600,
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          gap: '6px',
          transition: 'all 0.2s ease',
          ['WebkitAppRegion' as any]: 'no-drag',
        }}
      >
        <span>🧠</span>
        <span>BRAIN</span>
      </button>

      {isOpen && ReactDOM.createPortal(modalJSX, document.body)}
    </>
  );
}
