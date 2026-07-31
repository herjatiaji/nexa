import type { DecisionExplanation } from '../wails.d';

interface DecisionCardProps {
  explanation: DecisionExplanation | null;
}

export default function DecisionCard({ explanation }: DecisionCardProps) {
  if (!explanation) {
    return (
      <div style={{ background: 'rgba(0, 40, 50, 0.4)', borderRadius: '12px', padding: '16px', border: '1px solid rgba(0, 229, 255, 0.15)', color: 'rgba(255,255,255,0.5)', fontSize: '13px' }}>
        Awaiting initial decision calculation...
      </div>
    );
  }

  const getDecisionBadge = (d: string) => {
    switch (d) {
      case 'SILENT':
      case 'OBSERVE':
      case 'IGNORE':
        return { label: '🤫 SILENT OBSERVE', bg: 'rgba(128, 203, 196, 0.2)', color: '#80cbc4' };
      case 'TOAST':
      case 'SUGGEST':
        return { label: '💡 TOAST SUGGESTION', bg: 'rgba(255, 214, 0, 0.25)', color: '#ffd600' };
      case 'SPEAK':
        return { label: '🗣️ SPOKEN RESPONSE', bg: 'rgba(0, 229, 255, 0.25)', color: '#00e5ff' };
      case 'EXECUTE':
        return { label: '⚡ AUTONOMOUS ACTION', bg: 'rgba(179, 136, 255, 0.25)', color: '#b388ff' };
      default:
        return { label: d, bg: 'rgba(255,255,255,0.1)', color: '#fff' };
    }
  };

  const badge = getDecisionBadge(explanation.decision);

  return (
    <div
      style={{
        background: 'linear-gradient(135deg, rgba(5, 30, 40, 0.85) 0%, rgba(10, 20, 30, 0.9) 100%)',
        borderRadius: '16px',
        padding: '16px',
        border: '1px solid rgba(0, 229, 255, 0.25)',
        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
        <span
          style={{
            background: badge.bg,
            color: badge.color,
            padding: '4px 10px',
            borderRadius: '20px',
            fontSize: '11px',
            fontWeight: 'bold',
            letterSpacing: '0.5px',
          }}
        >
          {badge.label}
        </span>
        <span style={{ fontSize: '11px', color: 'rgba(255,255,255,0.5)' }}>
          Conf: {(explanation.confidence * 100).toFixed(0)}% • Autonomy: {explanation.autonomy}
        </span>
      </div>

      <div style={{ fontSize: '13px', color: '#e0f7fa', fontWeight: 500, marginBottom: '12px', lineHeight: '1.4' }}>
        "{explanation.summary}"
      </div>

      {explanation.factors && explanation.factors.length > 0 && (
        <div style={{ borderTop: '1px solid rgba(255,255,255,0.08)', paddingTop: '10px' }}>
          <div style={{ fontSize: '11px', color: 'rgba(255,255,255,0.4)', textTransform: 'uppercase', marginBottom: '6px', letterSpacing: '0.5px' }}>
            Decision Factors
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
            {explanation.factors.map((f, i) => (
              <div key={i} style={{ fontSize: '11px', color: 'rgba(255,255,255,0.8)', display: 'flex', alignItems: 'center', gap: '6px' }}>
                <span style={{ color: '#00e5ff' }}>•</span> {f}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
