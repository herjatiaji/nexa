const tools = [
  { icon: '🧠', name: 'Memory Store' },
  { icon: '🛡️', name: 'Risk Analyzer' },
  { icon: '👁️', name: 'Screen Vision' },
  { icon: '🌐', name: 'Web & Weather' },
  { icon: '🖥️', name: 'Desktop Apps' },
  { icon: '🔌', name: 'MCP Protocol' },
  { icon: '📁', name: 'File System' },
  { icon: '⚙️', name: 'Terminal' },
];

export default function ToolsList() {
  return (
    <>
      <div className="card-header">
        <span className="card-title">Registered Tools</span>
      </div>
      <div className="tools-list">
        {tools.map(t => (
          <div key={t.name} className="tool-item">
            <span>{t.icon}</span>
            <span>{t.name}</span>
          </div>
        ))}
      </div>
    </>
  );
}
