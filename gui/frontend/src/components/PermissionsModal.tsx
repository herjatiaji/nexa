import { useState, useEffect } from 'react';

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
          gap: '5px'
        }}
      >
        <span>🛡️</span>
        <span>PERMISSIONS</span>
      </button>

      {open && (
        <div style={{
          position: 'fixed',
          top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.7)',
          backdropFilter: 'blur(10px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 2000
        }}>
          <div style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: '16px',
            width: '440px',
            padding: '20px',
            display: 'flex',
            flexDirection: 'column',
            gap: '14px',
            boxShadow: '0 0 30px rgba(0,0,0,0.8)'
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '14px', fontWeight: 700, color: '#fff', letterSpacing: '1px' }}>
                🛡️ OS AGENT PERMISSION MODEL
              </span>
              <button
                onClick={() => setOpen(false)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: '14px', cursor: 'pointer' }}
              >
                ✕
              </button>
            </div>
            <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
              Configure granular security access rules for NEXA OS capabilities (ALLOW, CONFIRM, DENY):
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', maxHeight: '300px', overflowY: 'auto' }}>
              {Object.keys(perms).map(cap => {
                const lvl = perms[cap];
                const color = lvl === 'ALLOW' ? 'var(--success)' : lvl === 'CONFIRM' ? 'var(--warning)' : 'var(--danger)';
                return (
                  <div key={cap} style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '8px 12px',
                    background: 'rgba(255,255,255,0.03)',
                    borderRadius: '8px',
                    fontSize: '12px',
                    fontFamily: "'JetBrains Mono', monospace"
                  }}>
                    <span style={{ color: '#fff' }}>{cap}</span>
                    <button
                      onClick={() => handleToggle(cap, lvl)}
                      style={{
                        background: 'rgba(255,255,255,0.05)',
                        border: `1px solid ${color}`,
                        color: color,
                        padding: '4px 10px',
                        borderRadius: '6px',
                        fontSize: '11px',
                        fontWeight: 700,
                        cursor: 'pointer'
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
                padding: '8px',
                borderRadius: '8px',
                fontWeight: 600,
                cursor: 'pointer'
              }}
            >
              Done
            </button>
          </div>
        </div>
      )}
    </>
  );
}
