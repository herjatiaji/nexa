declare global {
  interface Window {
    go: {
      gui: {
        App: {
          Chat(message: string): Promise<ChatResult>;
          StartVoiceEngine(): Promise<string>;
          StopVoiceEngine(): Promise<string>;
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
