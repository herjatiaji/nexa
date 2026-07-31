import { useState, useRef, useEffect } from 'react';
import type { ChatResult } from '../wails.d';

interface ToolCall {
  name: string;
  args: string;
}

interface Message {
  role: 'user' | 'assistant';
  content: string;
  toolCalls?: ToolCall[];
  source?: 'text' | 'voice';
}

export default function ChatStream() {
  const [messages, setMessages] = useState<Message[]>([
    { role: 'assistant', content: 'Hello! I am NEXA, your personal AI assistant. Ask me anything or issue a desktop command.' }
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const streamRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (streamRef.current) {
      streamRef.current.scrollTop = streamRef.current.scrollHeight;
    }
  }, [messages, loading]);

  useEffect(() => {
    if (window.runtime) {
      const cancelCmd = window.runtime.EventsOn('nexa:voice:command', (cmd: string) => {
        setLoading(true);
        setMessages(prev => [...prev, { role: 'user', content: cmd, source: 'voice' }]);
      });

      const cancelRes = window.runtime.EventsOn('nexa:voice:result', (result: ChatResult) => {
        setLoading(false);
        setMessages(prev => [
          ...prev,
          {
            role: 'assistant',
            content: result.response,
            toolCalls: result.toolCalls,
            source: 'voice'
          }
        ]);
      });

      return () => {
        if (cancelCmd) cancelCmd();
        if (cancelRes) cancelRes();
      };
    }
  }, []);

  const sendMessage = async () => {
    const val = input.trim();
    if (!val || loading) return;

    setInput('');
    setLoading(true);
    setMessages(prev => [...prev, { role: 'user', content: val, source: 'text' }]);

    try {
      const result = await window.go.gui.App.Chat(val);
      setMessages(prev => [
        ...prev,
        {
          role: 'assistant',
          content: result.response,
          toolCalls: result.toolCalls,
          source: 'text'
        }
      ]);
    } catch (err: any) {
      setMessages(prev => [
        ...prev,
        { role: 'assistant', content: `❌ Error: ${err?.message || err}` }
      ]);
    } finally {
      setLoading(false);
      inputRef.current?.focus();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  return (
    <div className="card chat-card">
      <div className="card-header">
        <span className="card-title">Conversation</span>
      </div>
      <div className="chat-messages" ref={streamRef}>
        {messages.map((msg, i) => (
          <div key={i} className={`msg ${msg.role}`}>
            {msg.toolCalls && msg.toolCalls.length > 0 && (
              <div>
                {msg.toolCalls.map((tc, j) => (
                  <span key={j} className="tool-badge">
                    🔧 {tc.name}{tc.args ? ` · ${tc.args.substring(0, 60)}` : ''}
                  </span>
                ))}
              </div>
            )}
            <div className="bubble" dangerouslySetInnerHTML={{
              __html: msg.content
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/\n/g, '<br/>')
            }} />
            <div className="msg-meta">
              {msg.role === 'user' ? (msg.source === 'voice' ? '🎙️ You (Voice)' : 'You') : 'NEXA Engine'}
            </div>
          </div>
        ))}
        {loading && (
          <div className="msg assistant">
            <div className="bubble typing-indicator">⏳ NEXA is thinking...</div>
          </div>
        )}
      </div>
      <div className="chat-input-bar">
        <input
          ref={inputRef}
          type="text"
          className="chat-input"
          placeholder="Ask NEXA anything..."
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={loading}
        />
        <button className="send-btn" onClick={sendMessage} disabled={loading}>
          Send
        </button>
      </div>
    </div>
  );
}
