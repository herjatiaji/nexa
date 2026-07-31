import { useState, useEffect } from 'react';
import FloatingCompanion from './FloatingCompanion';

interface SpeechPlanEvent {
  chunks: Array<{
    text: string;
    emotion: string;
    prosody: { speedScale: number; pitchShift: number; sentenceSilence: number; emphasis: number };
  }>;
  dominantEmotion: string;
  confidence: number;
  personality: string;
}

const emotionColors: Record<string, string> = {
  neutral: '#00e5ff',
  happy: '#76ff03',
  excited: '#ffd740',
  thoughtful: '#ba68c8',
  urgent: '#ff5252',
  sad: '#80deea',
  confident: '#00e676',
};

const emotionLabels: Record<string, string> = {
  neutral: '🔵 Neutral',
  happy: '😊 Happy',
  excited: '🔥 Excited',
  thoughtful: '🤔 Thoughtful',
  urgent: '⚡ Urgent',
  sad: '😔 Concern',
  confident: '💪 Confident',
};

export default function VoiceReactor() {
  const [voiceActive, setVoiceActive] = useState(false);
  const [speechPlan, setSpeechPlan] = useState<SpeechPlanEvent | null>(null);
  const [isSpeaking, setIsSpeaking] = useState(false);

  useEffect(() => {
    window.go.gui.App.GetStatus().then(status => {
      if (status.voiceActive) {
        setVoiceActive(true);
      }
    });

    if (window.runtime) {
      const cancelVoiceState = window.runtime.EventsOn('nexa:voice:state', (active: boolean) => {
        setVoiceActive(active);
      });

      const cancelSpeechPlan = window.runtime.EventsOn('nexa:speech:plan', (plan: SpeechPlanEvent) => {
        setSpeechPlan(plan);
        setIsSpeaking(true);
        // Auto-clear after speech finishes (estimate ~3s per chunk)
        const duration = Math.max(3000, (plan.chunks?.length || 1) * 4000);
        setTimeout(() => {
          setIsSpeaking(false);
        }, duration);
      });

      return () => {
        if (cancelVoiceState) cancelVoiceState();
        if (cancelSpeechPlan) cancelSpeechPlan();
      };
    }
  }, []);

  const toggleVoice = async () => {
    if (voiceActive) {
      await window.go.gui.App.StopVoiceEngine();
      setVoiceActive(false);
    } else {
      const res = await window.go.gui.App.StartVoiceEngine();
      if (res === 'OK' || res === 'Voice engine already active') {
        setVoiceActive(true);
      }
    }
  };

  const dominantEmotion = speechPlan?.dominantEmotion || 'neutral';
  const emotionColor = emotionColors[dominantEmotion] || '#00e5ff';

  return (
    <>
      <div className="card-header">
        <span className="card-title">AI Mascot Companion</span>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {/* Speech emotion indicator badge */}
          {isSpeaking && speechPlan && (
            <div
              style={{
                background: `linear-gradient(135deg, ${emotionColor}44, ${emotionColor}88)`,
                border: `1px solid ${emotionColor}`,
                borderRadius: '14px',
                padding: '3px 10px',
                fontSize: '10px',
                fontWeight: 700,
                color: '#ffffff',
                display: 'flex',
                alignItems: 'center',
                gap: '4px',
                animation: 'pulse 1.5s ease-in-out infinite',
                boxShadow: `0 0 12px ${emotionColor}66`,
                textShadow: '0 1px 2px rgba(0,0,0,0.4)',
              }}
            >
              <span>{emotionLabels[dominantEmotion] || '🔵 Speaking'}</span>
              {speechPlan.confidence > 0 && (
                <span style={{ opacity: 0.7, fontSize: '9px' }}>
                  {Math.round(speechPlan.confidence * 100)}%
                </span>
              )}
            </div>
          )}
          <button
            onClick={toggleVoice}
            style={{
              background: voiceActive
                ? 'linear-gradient(180deg, #a7ffeb 0%, #00e676 50%, #00c853 100%)'
                : 'linear-gradient(180deg, rgba(255,255,255,0.3) 0%, rgba(255,255,255,0.1) 100%)',
              border: '1px solid rgba(255, 255, 255, 0.8)',
              color: '#ffffff',
              padding: '5px 12px',
              borderRadius: '20px',
              fontSize: '11px',
              cursor: 'pointer',
              fontWeight: 700,
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              boxShadow: 'inset 0 1px 2px rgba(255, 255, 255, 0.9), 0 2px 8px rgba(0, 0, 0, 0.15)',
              textShadow: '0 1px 2px rgba(0,0,0,0.3)',
            }}
          >
            <span>🎤</span>
            <span>{voiceActive ? 'VOICE ON' : 'START VOICE'}</span>
          </button>
        </div>
      </div>
      <FloatingCompanion />
    </>
  );
}
