import { useState, useEffect } from 'react';
import ReactDOM from 'react-dom';

export default function PermissionsModal() {
  const [open, setOpen] = useState(false);
  const [perms, setPerms] = useState<Record<string, string>>({});

  const fetchPermissions = async () => {
    try {
      if (window.go?.gui?.App?.GetPermissions) {
        const data = await window.go.gui.App.GetPermissions();
        setPerms(data || {});
      }
    } catch (e) {
      console.error('Failed to fetch permissions', e);
    }
  };

  useEffect(() => {
    if (open) fetchPermissions();
  }, [open]);

  const handleToggle = async (cap: string, currentLevel: string) => {
    const nextLevel = currentLevel === 'ALLOW' ? 'CONFIRM' : currentLevel === 'CONFIRM' ? 'DENY' : 'ALLOW';
    try {
      await window.go.gui.App.SetPermission(cap, nextLevel);
      setPerms(prev => ({ ...prev, [cap]: nextLevel }));
    } catch (e) {
      console.error('Failed to update permission', e);
    }
  };

  const modalJSX = open ? (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100vw',
        height: '100vh',
        background: 'rgba(5, 9, 18, 0.85)',
        backdropFilter: 'blur(12px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 99999,
        WebkitAppRegion: 'no-drag',
      } as React.CSSProperties}
    >
      <div
        style={{
          background: 'rgba(15, 23, 42, 0.95)',
          border: '1px solid var(--border)',
          borderRadius: '16px',
          width: '460px',
          maxHeight: '85vh',
          padding: '24px',
          display: 'flex',
          flexDirection: 'column',
          gap: '16px',
          boxShadow: '0 0 40px rgba(56, 189, 248, 0.25)',
          overflow: 'hidden',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontSize: '15px', fontWeight: 700, color: '#fff', letterSpacing: '1px' }}>
            🛡️ OS AGENT PERMISSION MODEL
          </span>
          <button
            onClick={() => setOpen(false)}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--text-muted)',
              fontSize: '16px',
              cursor: 'pointer',
              padding: '4px',
            }}
          >
            ✕
          </button>
        </div>

        <div style={{ fontSize: '12px', color: 'var(--text-muted)', lineHeight: '1.4' }}>
          Configure granular security access rules for NEXA OS capabilities (ALLOW, CONFIRM, DENY):
        </div>

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: '8px',
            maxHeight: '340px',
            overflowY: 'auto',
            paddingRight: '4px',
          }}
        >
          {Object.keys(perms).map(cap => {
            const lvl = perms[cap];
            const color =
              lvl === 'ALLOW' ? 'var(--success)' : lvl === 'CONFIRM' ? 'var(--warning)' : 'var(--danger)';
            return (
              <div
                key={cap}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '10px 14px',
                  background: 'rgba(255, 255, 255, 0.03)',
                  border: '1px solid rgba(255, 255, 255, 0.05)',
                  borderRadius: '8px',
                  fontSize: '12px',
                  fontFamily: "'JetBrains Mono', monospace",
                }}
              >
                <span style={{ color: '#fff', fontWeight: 500 }}>{cap}</span>
                <button
                  onClick={() => handleToggle(cap, lvl)}
                  style={{
                    background: 'rgba(255, 255, 255, 0.05)',
                    border: `1px solid ${color}`,
                    color: color,
                    padding: '4px 12px',
                    borderRadius: '6px',
                    fontSize: '11px',
                    fontWeight: 700,
                    cursor: 'pointer',
                    minWidth: '70px',
                  }}
                >
                  {lvl}
                </button>
              </div>
            );
          })}
        </div>

        <button
          onClick={() => setOpen(false)}
          style={{
            background: 'linear-gradient(135deg, var(--cyan), #0284c7)',
            border: 'none',
            color: '#fff',
            padding: '10px',
            borderRadius: '10px',
            fontWeight: 600,
            fontSize: '13px',
            cursor: 'pointer',
            marginTop: '4px',
          }}
        >
          Save & Close
        </button>
      </div>
    </div>
  ) : null;

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        style={{
          background: 'rgba(56, 189, 248, 0.08)',
          border: '1px solid var(--border)',
          color: 'var(--cyan)',
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
        <span>🛡️</span>
        <span>PERMISSIONS</span>
      </button>

      {modalJSX && ReactDOM.createPortal(modalJSX, document.body)}
    </>
  );
}
