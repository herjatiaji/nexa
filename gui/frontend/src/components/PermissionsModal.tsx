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
        style={{
          background: 'linear-gradient(135deg, rgba(255, 255, 255, 0.35) 0%, rgba(255, 255, 255, 0.15) 100%)',
          border: '1.5px solid rgba(255, 255, 255, 0.8)',
          borderRadius: '24px',
          width: '470px',
          maxHeight: '85vh',
          padding: '24px',
          display: 'flex',
          flexDirection: 'column',
          gap: '16px',
          boxShadow: 'inset 0 2px 4px rgba(255, 255, 255, 0.9), 0 20px 50px rgba(0, 77, 64, 0.5)',
          backdropFilter: 'blur(30px)',
          overflow: 'hidden',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontSize: '15px', fontWeight: 700, color: '#ffffff', letterSpacing: '1px', textShadow: '0 1px 3px rgba(0,0,0,0.3)' }}>
            🛡️ OS AGENT PERMISSION MODEL
          </span>
          <button
            onClick={() => setOpen(false)}
            style={{
              background: 'none',
              border: 'none',
              color: '#ffffff',
              fontSize: '16px',
              cursor: 'pointer',
              padding: '4px',
            }}
          >
            ✕
          </button>
        </div>

        <div style={{ fontSize: '12px', color: '#e0f7fa', lineHeight: '1.4', textShadow: '0 1px 2px rgba(0,0,0,0.2)' }}>
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
            const bgGrad =
              lvl === 'ALLOW'
                ? 'linear-gradient(180deg, #a7ffeb 0%, #00e676 100%)'
                : lvl === 'CONFIRM'
                ? 'linear-gradient(180deg, #ffe082 0%, #ffb300 100%)'
                : 'linear-gradient(180deg, #ff8a80 0%, #ff5252 100%)';
            return (
              <div
                key={cap}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '10px 14px',
                  background: 'linear-gradient(180deg, rgba(255, 255, 255, 0.25) 0%, rgba(255, 255, 255, 0.1) 100%)',
                  border: '1px solid rgba(255, 255, 255, 0.45)',
                  borderRadius: '12px',
                  fontSize: '12px',
                  fontFamily: "'JetBrains Mono', monospace",
                  boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.7)',
                }}
              >
                <span style={{ color: '#ffffff', fontWeight: 600, textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}>{cap}</span>
                <button
                  onClick={() => handleToggle(cap, lvl)}
                  style={{
                    background: bgGrad,
                    border: '1px solid rgba(255, 255, 255, 0.9)',
                    color: '#ffffff',
                    padding: '4px 14px',
                    borderRadius: '14px',
                    fontSize: '11px',
                    fontWeight: 700,
                    cursor: 'pointer',
                    minWidth: '76px',
                    boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.9), 0 2px 6px rgba(0,0,0,0.2)',
                    textShadow: '0 1px 2px rgba(0,0,0,0.3)',
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
            background: 'linear-gradient(180deg, #a7ffeb 0%, #00e676 50%, #00c853 100%)',
            border: '1px solid rgba(255, 255, 255, 0.9)',
            color: '#ffffff',
            padding: '10px',
            borderRadius: '16px',
            fontWeight: 700,
            fontSize: '13px',
            cursor: 'pointer',
            marginTop: '4px',
            boxShadow: 'inset 0 2px 4px rgba(255, 255, 255, 0.9), 0 4px 14px rgba(0, 230, 118, 0.5)',
            textShadow: '0 1px 2px rgba(0, 70, 30, 0.5)',
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
          background: 'linear-gradient(180deg, rgba(255, 255, 255, 0.35) 0%, rgba(255, 255, 255, 0.15) 100%)',
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
        <span>🛡️</span>
        <span>PERMISSIONS</span>
      </button>

      {modalJSX && ReactDOM.createPortal(modalJSX, document.body)}
    </>
  );
}
