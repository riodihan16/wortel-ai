import { useCallback, useEffect, useRef, useState } from "react";
import { EventsOff, EventsOn } from "../../wailsjs/runtime/runtime";
import { CancelGeneration, SendMessage } from "../../wailsjs/go/main/App";
import type { ChatEvent, ChatMessage, ChatStatus } from "../types/chat";
const CONVERSATION_ID = "default";

// useChat owns send/cancel + streaming state. Matches deltas by requestId.
export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ChatStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const activeRequest = useRef<string | null>(null);
  useEffect(() => {

    const handler = (ev: ChatEvent) => {
      if (ev.conversationId !== CONVERSATION_ID) return;
      // Started dikirim backend sebelum requestId tiba di frontend, adopsi di sini.
      if (ev.type === "started" && activeRequest.current === null) {
        activeRequest.current = ev.requestId;
      } else if (ev.requestId !== activeRequest.current) {
        return;
      }
      switch (ev.type) {
        case "started":
          setStatus("streaming");
          setMessages((prev) => [
            ...prev,
            { id: ev.messageId, role: "assistant", content: "", status: "streaming" },
          ]);
          break;
        case "delta":
          setMessages((prev) =>
            prev.map((m) =>
              m.id === ev.messageId ? { ...m, content: m.content + (ev.delta ?? "") } : m
            )
          );
          break;
        case "completed":
          setStatus("complete");
          setMessages((prev) =>
            prev.map((m) => (m.id === ev.messageId ? { ...m, status: "complete" } : m))
          );
          activeRequest.current = null;
          break;
        case "cancelled":
          setStatus("cancelled");
          setMessages((prev) =>
            prev.map((m) => (m.id === ev.messageId ? { ...m, status: "cancelled" } : m))
          );
          activeRequest.current = null;
          break;
        case "failed":
          setStatus("error");
          setError(ev.error?.message ?? "Request gagal.");
          setMessages((prev) =>
            prev.map((m) => (m.id === ev.messageId ? { ...m, status: "error" } : m))
          );
          activeRequest.current = null;
          break;
      }
    };
    EventsOn("chat:event", handler);
    return () => {
      EventsOff("chat:event");
    };
  }, []);

  const send = useCallback(async (text: string) => {
    const trimmed = text.trim();
    if (!trimmed || activeRequest.current) return;
    setError(null);
    setStatus("sending");
    setMessages((prev) => [
      ...prev,
      { id: `user-${Date.now()}`, role: "user", content: trimmed, status: "complete" },
    ]);
    try {
      const requestId = (await SendMessage(CONVERSATION_ID, trimmed)) as string;
      activeRequest.current = requestId;
    } catch (e) {
      setStatus("error");
      setError(e instanceof Error ? e.message : "Gagal mengirim.");
    }
  }, []);

  const cancel = useCallback(async () => {
    if (!activeRequest.current) return;
    await CancelGeneration(CONVERSATION_ID);
  }, []);

  const busy = status === "sending" || status === "streaming";

  return { messages, status, error, busy, send, cancel, conversationId: CONVERSATION_ID };
}
