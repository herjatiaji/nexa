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
          GetMemories(): Promise<Record<string, string>>;
          DeleteMemory(key: string): Promise<void>;
          GetStatus(): Promise<StatusInfo>;
          ResetConversation(): Promise<void>;
        };
      };
    };
    runtime: {
      EventsOn(eventName: string, callback: (...data: any[]) => void): () => void;
      EventsOff(eventName: string, ...additionalEventNames: string[]): void;
    };
  }
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
