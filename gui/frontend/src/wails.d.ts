declare global {
  interface Window {
    go: {
      gui: {
        App: {
          Chat(message: string): Promise<ChatResult>;
          StartVoiceEngine(): Promise<string>;
          StopVoiceEngine(): Promise<string>;
          GetTraces(): Promise<TraceStep[]>;
          CreatePlan(prompt: string): Promise<Plan>;
          SearchMemories(query: string): Promise<SemanticMatch[]>;
          GetPermissions(): Promise<Record<string, string>>;
          SetPermission(capability: string, level: string): Promise<void>;
          GetMemories(): Promise<Record<string, string>>;
          DeleteMemory(key: string): Promise<void>;
          GetStatus(): Promise<StatusInfo>;
          ResetConversation(): Promise<void>;
          GetBrainState(): Promise<BrainSnapshot>;
          GetCognitiveTraces(limit?: number): Promise<CognitiveTrace[]>;
          GetBrainMetrics(): Promise<BrainMetricsTelemetry>;
          ExplainLastDecision(): Promise<DecisionExplanation>;
        };
      };
    };
    runtime: {
      EventsOn(eventName: string, callback: (...data: any[]) => void): () => void;
      EventsOff(eventName: string, ...additionalEventNames: string[]): void;
    };
  }
}

export interface CognitiveContext {
  activity: string;
  focused_app: string;
  confidence: number;
  idle_duration: number;
  problem_detected: boolean;
  time_of_day: string;
}

export interface SocialState {
  attention_score: number;
  stress_level: number;
  can_interrupt: boolean;
  urgency_threshold: number;
}

export interface BrainSnapshot {
  Context: CognitiveContext;
  Social: SocialState;
  Autonomy: number;
  CycleCount: number;
  LastTickAt: string;
}

export interface CognitiveTrace {
  id: string;
  cycle_id: string;
  stage: string;
  component: string;
  input: any;
  output: any;
  duration: number;
  timestamp: string;
}

export interface BrainMetricsTelemetry {
  total_cycles: number;
  avg_tick_latency_ms: number;
  total_decisions: number;
  suggestions_emitted: number;
  suggestions_accepted: number;
  silent_observations: number;
  uptime_seconds: number;
  last_cycle_time: string;
}

export interface DecisionExplanation {
  decision: string;
  summary: string;
  factors: string[];
  social_score: number;
  autonomy: string;
  confidence: number;
  timestamp: string;
}

export interface ToolCallLog {
  name: string;
  args: string;
}

export interface ChatResult {
  response: string;
  toolCalls: ToolCallLog[];
  memories: Record<string, string>;
}

export interface StatusInfo {
  provider: string;
  model: string;
  tts: boolean;
  voiceActive: boolean;
  version: string;
}

export interface MascotState {
  emotion: string;
  message?: string;
  eyeSymbol: string;
  auraColor: string;
  timestamp: string;
}

export interface TraceStep {
  index: number;
  thought: string;
  tool: string;
  arguments: string;
  result: string;
  response: string;
  timestamp: string;
}

export interface TaskStep {
  id: number;
  description: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  result?: string;
}

export interface Plan {
  goal: string;
  steps: TaskStep[];
  created_at: string;
}

export interface SemanticMatch {
  fact: {
    key: string;
    value: string;
    updated_at: string;
  };
  score: number;
}
