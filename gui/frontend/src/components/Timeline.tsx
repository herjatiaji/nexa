import type { CognitiveTrace } from '../wails.d';

interface TimelineProps {
  traces: CognitiveTrace[];
}

export default function Timeline({ traces }: TimelineProps) {
  if (!traces || traces.length === 0) {
    return (
      <div style={{ color: 'rgba(255,255,255,0.4)', textAlign: 'center', padding: '20px 0', fontSize: '13px' }}>
        No cognitive traces recorded yet. Active cycles will stream here.
      </div>
    );
  }

  const getStageColor = (stage: string) => {
    switch (stage) {
      case 'perceive':
        return '#00e5ff'; // Cyan
      case 'update':
        return '#76ff03'; // Bright Green
      case 'think':
        return '#ffd600'; // Amber
      case 'decide':
        return '#ff4081'; // Pink
      case 'execute':
        return '#b388ff'; // Purple
      default:
        return '#80d8ff';
    }
  };

  return (
    <div className="cognitive-timeline" style={{ display: 'flex', flexDirection: 'column', gap: '8px', maxHeight: '280px', overflowY: 'auto', paddingRight: '4px' }}>
      {traces.map((tr) => (
        <div
          key={tr.id || `${tr.cycle_id}_${tr.component}_${tr.timestamp}`}
          style={{
            background: 'rgba(10, 25, 30, 0.65)',
            borderLeft: `4px solid ${getStageColor(tr.stage)}`,
            borderRadius: '8px',
            padding: '8px 12px',
            fontSize: '12px',
            fontFamily: 'monospace',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
            <span style={{ color: getStageColor(tr.stage), fontWeight: 'bold', textTransform: 'uppercase' }}>
              [{tr.stage}] {tr.component}
            </span>
            <span style={{ color: 'rgba(255,255,255,0.4)', fontSize: '11px' }}>
              {tr.cycle_id} • {(tr.duration / 1e6).toFixed(1)}ms
            </span>
          </div>

          {tr.input && (
            <div style={{ color: 'rgba(255,255,255,0.7)', fontSize: '11px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              <span style={{ color: 'rgba(255,255,255,0.4)' }}>in: </span>
              {typeof tr.input === 'object' ? JSON.stringify(tr.input) : String(tr.input)}
            </div>
          )}

          {tr.output && (
            <div style={{ color: '#a7ffeb', fontSize: '11px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              <span style={{ color: 'rgba(255,255,255,0.4)' }}>out: </span>
              {typeof tr.output === 'object' ? JSON.stringify(tr.output) : String(tr.output)}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
