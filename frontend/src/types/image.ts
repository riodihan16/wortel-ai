import type { PublicError } from "./chat";

/** One image-generation lifecycle event from the Go backend. */
export type ImageEventType = "started" | "completed" | "cancelled" | "failed";

export interface ImageEvent {
  type: ImageEventType;
  requestId: string;
  conversationId: string;
  messageId: string;
  mime?: string;
  prompt?: string;
  error?: PublicError;
}

export type ImageStatus = "idle" | "pending" | "complete" | "cancelled" | "error";
