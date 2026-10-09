import { useCallback, useEffect, useRef, useState } from "react";
import { EventsOff, EventsOn } from "../../wailsjs/runtime/runtime";
import { CancelGeneration, History, SendMessage } from "../../wailsjs/go/main/App";
import type { ChatEvent, ChatMessage, ChatStatus } from "../types/chat";
const CONVERSATION_ID = "default";

// useChat owns send/cancel + streaming state. Matches deltas by requestId.
export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ChatStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const activeRequest = useRef<string | null>(null);
  // Guard sync: activeRequest baru terisi setelah SendMessage resolve,
  // klik ganda sebelum itu harus ditolak di sini.
  const sendingRef = useRef(false);
  useEffect(() => {
    History(CONVERSATION_ID, 100).then((rows) => {
      const list = (rows ?? []) as { id: string; role: string; content: string; status: string }[];
      setMessages(
        list
          .filter((r) => r.role === "user" || r.role === "assistant")
          .map((r) => ({
            id: r.id,
            role: r.role as "user" | "assistant",
            content: r.content,
            status: (r.status === "cancelled" ? "cancelled" : r.status === "failed" || r.status === "error" ? "error" : "complete") as ChatStatus,
          }))
      );
    }).catch(() => {});
  }, []);
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
          sendingRef.current = false;
          break;
        case "cancelled":
          setStatus("cancelled");
          setMessages((prev) =>
            prev.map((m) => (m.id === ev.messageId ? { ...m, status: "cancelled" } : m))
          );
          activeRequest.current = null;
          sendingRef.current = false;
          break;
        case "failed":
          setStatus("error");
          setError(ev.error?.message ?? "Request gagal.");
          setMessages((prev) =>
            prev.map((m) => (m.id === ev.messageId ? { ...m, status: "error" } : m))
          );
          activeRequest.current = null;
          sendingRef.current = false;
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
    if (!trimmed || sendingRef.current || activeRequest.current) return;
    sendingRef.current = true;
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
      sendingRef.current = false;
    }
  }, []);

  const cancel = useCallback(async () => {
    // Saat sending, requestId belum ada: batalkan via backend lalu lepas guard.
    await CancelGeneration(CONVERSATION_ID);
    sendingRef.current = false;
  }, []);

  const busy = status === "sending" || status === "streaming";

  return { messages, status, error, busy, send, cancel, conversationId: CONVERSATION_ID };
}
