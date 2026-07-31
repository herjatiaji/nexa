import { useState } from 'react';
import VoiceReactor from './components/VoiceReactor';
import ToolsList from './components/ToolsList';
import ChatStream from './components/ChatStream';
import MemoryPanel from './components/MemoryPanel';
import StatusBar from './components/StatusBar';

export default function App() {
  const [memRefresh, setMemRefresh] = useState(0);

  return (
    <div className="app">
      {/* Header */}
      <header className="header">
        <div className="brand">
          <div className="logo-ring"><div className="logo-dot" /></div>
          <h1>NEXA</h1>
        </div>
        <StatusBar />
      </header>

      {/* 3-Column Dashboard */}
      <div className="dashboard">
        {/* Left — Voice Reactor + Tools */}
        <div className="card">
          <VoiceReactor />
          <ToolsList />
        </div>

        {/* Center — Chat */}
        <ChatStream />

        {/* Right — Memory Inspector */}
        <MemoryPanel refreshTrigger={memRefresh} />
      </div>
    </div>
  );
}
