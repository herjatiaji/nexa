import { useState, useEffect, useCallback } from 'react';

interface Props {
  refreshTrigger: number;
}

export default function MemoryPanel({ refreshTrigger }: Props) {
  const [memories, setMemories] = useState<Record<string, string>>({});

  const fetchMemories = useCallback(async () => {
    try {
      const data = await window.go.gui.App.GetMemories();
      setMemories(data || {});
    } catch (e) {
      console.error('Failed to fetch memories', e);
    }
  }, []);

  useEffect(() => {
    fetchMemories();
  }, [fetchMemories, refreshTrigger]);

  const handleDelete = async (key: string) => {
    try {
      await window.go.gui.App.DeleteMemory(key);
      fetchMemories();
    } catch (e) {
      console.error('Failed to delete memory', e);
    }
  };

  const keys = Object.keys(memories);

  return (
    <div className="card">
      <div className="card-header">
        <span className="card-title">Persistent Memory</span>
        <button className="refresh-btn" onClick={fetchMemories}>🔄</button>
      </div>
      <div className="memory-list">
        {keys.length === 0 ? (
          <div className="empty-text">No memories saved yet.</div>
        ) : (
          keys.map(k => (
            <div key={k} className="mem-card">
              <div className="mem-content">
                <div className="mem-key">{k}</div>
                <div className="mem-val">{memories[k]}</div>
              </div>
              <button className="mem-delete" onClick={() => handleDelete(k)} title="Delete">✕</button>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
