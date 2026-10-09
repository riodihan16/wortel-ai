import { useCallback, useEffect, useRef, useState } from "react";
import { EventsOff, EventsOn } from "../../wailsjs/runtime/runtime";
import {
  CancelGeneration,
  DeleteConversation,
  History,
  ListConversations,
  RenameConversation,
  SendMessage,
} from "../../wailsjs/go/main/App";
import type { ChatEvent, ChatMessage, ChatStatus } from "../types/chat";

export interface ConversationItem {
  id: string;
  title: string;
}

// useChat owns conversation list + send/cancel + streaming state.
// Matches deltas by requestId.
export function useChat() {
  const [conversations, setConversations] = useState<ConversationItem[]>([]);
  const [conversationId, setConversationId] = useState<string>("default");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ChatStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const activeRequest = useRef<string | null>(null);
  // Guard sync: activeRequest baru terisi setelah SendMessage resolve,
  // klik ganda sebelum itu harus ditolak di sini.
  const sendingRef = useRef(false);
  const convRef = useRef(conversationId);
  convRef.current = conversationId;

  const refreshList = useCallback(async () => {
    try {
      const rows = (await ListConversations()) as ConversationItem[] | null;
      setConversations(rows ?? []);
    } catch {
      // List gagal: biarkan daftar lama.
    }
  }, []);

  const loadConversation = useCallback(async (id: string) => {
    setConversationId(id);
    setError(null);
    setStatus("idle");
    activeRequest.current = null;
    sendingRef.current = false;
    try {
      const rows = (await History(id, 100)) as
        | { id: string; role: string; content: string; status: string }[]
        | null;
      setMessages(
        (rows ?? [])
          .filter((r) => r.role === "user" || r.role === "assistant")
          .map((r) => ({
            id: r.id,
            role: r.role as "user" | "assistant",
            content: r.content,
            status: (r.status === "cancelled"
              ? "cancelled"
              : r.status === "failed" || r.status === "error"
                ? "error"
                : "complete") as ChatStatus,
          }))
      );
    } catch {
      setMessages([]);
    }
  }, []);

  useEffect(() => {
    refreshList().then(() => loadConversation("default"));
  }, [refreshList, loadConversation]);

  useEffect(() => {
    const handler = (ev: ChatEvent) => {
      if (ev.conversationId !== convRef.current) return;
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
          refreshList();
          break;
        case "cancelled":
          setStatus("cancelled");
          setMessages((prev) =>
            prev.map((m) => (m.id === ev.messageId ? { ...m, status: "cancelled" } : m))
          );
          activeRequest.current = null;
          sendingRef.current = false;
          refreshList();
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
  }, [refreshList]);

  const send = useCallback(
    async (text: string) => {
      const trimmed = text.trim();
      if (!trimmed || sendingRef.current || activeRequest.current) return;
      sendingRef.current = true;
      setError(null);
      setStatus("sending");
      const conv = convRef.current;
      setMessages((prev) => [
        ...prev,
        { id: `user-${Date.now()}`, role: "user", content: trimmed, status: "complete" },
      ]);
      try {
        const requestId = (await SendMessage(conv, trimmed)) as string;
        activeRequest.current = requestId;
      } catch (e) {
        setStatus("error");
        setError(e instanceof Error ? e.message : "Gagal mengirim.");
        sendingRef.current = false;
      }
    },
    []
  );

  const cancel = useCallback(async () => {
    // Saat sending, requestId belum ada: batalkan via backend lalu lepas guard.
    await CancelGeneration(convRef.current);
    sendingRef.current = false;
  }, []);

  const newConversation = useCallback(async () => {
    const id =
      typeof crypto !== "undefined" && "randomUUID" in crypto
        ? crypto.randomUUID()
        : `conv-${Date.now()}`;
    await loadConversation(id);
    await refreshList();
  }, [loadConversation, refreshList]);

  const switchConversation = useCallback(
    async (id: string) => {
      if (id === convRef.current) return;
      await loadConversation(id);
    },
    [loadConversation]
  );

  const renameConversation = useCallback(
    async (id: string, title: string) => {
      await RenameConversation(id, title);
      await refreshList();
    },
    [refreshList]
  );

  const deleteConversation = useCallback(
    async (id: string) => {
      await DeleteConversation(id);
      await refreshList();
      if (id === convRef.current) {
        await loadConversation("default");
      }
    },
    [loadConversation, refreshList]
  );

  const busy = status === "sending" || status === "streaming";

  return {
    messages,
    status,
    error,
    busy,
    send,
    cancel,
    conversationId,
    conversations,
    newConversation,
    switchConversation,
    renameConversation,
    deleteConversation,
    refreshList,
  };
}
