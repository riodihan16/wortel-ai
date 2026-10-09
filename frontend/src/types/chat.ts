export type ChatEventType = "started" | "delta" | "completed" | "cancelled" | "failed";

export interface PublicError {
  code: string;
  message: string;
  retryable: boolean;
  statusCode?: number;
}

export interface ChatEvent {
  type: ChatEventType;
  requestId: string;
  conversationId: string;
  messageId: string;
  delta?: string;
  error?: PublicError;
}

export type ChatStatus = "idle" | "sending" | "streaming" | "complete" | "cancelled" | "error";

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  status: ChatStatus;
}
