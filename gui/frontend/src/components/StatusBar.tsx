import { useState, useEffect } from 'react';
import type { StatusInfo } from '../wails.d';
import TraceInspector from './TraceInspector';

export default function StatusBar() {
  const [status, setStatus] = useState<StatusInfo | null>(null);

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const data = await window.go.gui.App.GetStatus();
        setStatus(data);
      } catch (e) {
        console.error('Failed to fetch status', e);
      }
    };
    fetchStatus();
  }, []);

  return (
    <div className="header-right" style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
      <TraceInspector />
      <span className="provider-tag">
        {status ? `${status.provider.toUpperCase()} (${status.model})` : 'Connecting...'}
      </span>
      <div className="status-badge">
        <div className="status-dot" />
        <span>v{status?.version || '1.3.0'}</span>
      </div>
    </div>
  );
}
