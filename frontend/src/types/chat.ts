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

export type AttachmentKind = "image" | "pdf" | "text";

/** Frontend-safe attachment description. Bytes stay in the Go backend. */
export interface AttachmentMeta {
  id: string;
  name: string;
  mime: string;
  size: number;
  kind: AttachmentKind;
}

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  status: ChatStatus;
  /** Why this reply failed, from the backend's failed event. */
  errorText?: string;
  attachments?: AttachmentMeta[];
}
