import { useCallback, useEffect, useRef, useState } from "react";
import { EventsOff, EventsOn } from "../../wailsjs/runtime/runtime";
import {
  CancelGeneration,
  DeleteConversation,
  History,
  ListConversations,
  RenameConversation,
  RetryMessage,
  SendMessage,
} from "../../wailsjs/go/main/App";
import type { ChatEvent, ChatMessage, ChatStatus } from "../types/chat";

export interface ConversationItem {
  id: string;
  title: string;
}

function freshId(): string {
  return typeof crypto !== "undefined" && "randomUUID" in crypto
    ? crypto.randomUUID()
    : `conv-${Date.now()}`;
}

// useChat owns conversation list + send/cancel + streaming state.
// Matches deltas by requestId. No special "default" id: startup picks
// the most recent conversation, or a fresh empty one when none exist.
export function useChat() {
  const [conversations, setConversations] = useState<ConversationItem[]>([]);
  const [conversationId, setConversationId] = useState<string>("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ChatStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const activeRequest = useRef<string | null>(null);
  // Guard sync: activeRequest baru terisi setelah SendMessage resolve,
  // klik ganda sebelum itu harus ditolak di sini.
  const sendingRef = useRef(false);
  const convRef = useRef(conversationId);
  convRef.current = conversationId;
  const messagesRef = useRef(messages);
  messagesRef.current = messages;

  const refreshList = useCallback(async (): Promise<ConversationItem[]> => {
    try {
      const rows = (await ListConversations()) as ConversationItem[] | null;
      const list = rows ?? [];
      setConversations(list);
      return list;
    } catch {
      // List gagal: biarkan daftar lama.
      return [];
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

  const loadEmpty = useCallback((id: string) => {
    setConversationId(id);
    setMessages([]);
    setError(null);
    setStatus("idle");
    activeRequest.current = null;
    sendingRef.current = false;
  }, []);

  useEffect(() => {
    refreshList().then((rows) => {
      if (rows.length > 0) {
        loadConversation(rows[0].id);
      } else {
        loadEmpty(freshId());
      }
    });
  }, [refreshList, loadConversation, loadEmpty]);

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

  // begin hands one turn to the backend. The requestId arrives after the
  // backend's "started" event may already have been adopted (see handler).
  const begin = useCallback(
    async (conv: string, text: string, call: (c: string, t: string) => Promise<string>) => {
      try {
        const requestId = await call(conv, text);
        activeRequest.current = requestId;
      } catch (e) {
        setStatus("error");
        setError(e instanceof Error ? e.message : "Gagal mengirim.");
        sendingRef.current = false;
      }
    },
    []
  );

  const send = useCallback(async (text: string) => {
    const trimmed = text.trim();
    if (!trimmed || sendingRef.current || activeRequest.current) return;
    let conv = convRef.current;
    if (!conv) {
      conv = freshId();
      convRef.current = conv;
      setConversationId(conv);
    }
    sendingRef.current = true;
    setError(null);
    setStatus("sending");
    setMessages((prev) => [
      ...prev,
      { id: `user-${Date.now()}`, role: "user", content: trimmed, status: "complete" },
    ]);
    await begin(conv, trimmed, SendMessage);
  }, [begin]);

  // retry re-runs the newest user turn in place: the failed reply bubble is
  // dropped and the backend reuses the stored user row, so neither the UI
  // nor history shows the question twice.
  const retry = useCallback(async () => {
    if (sendingRef.current || activeRequest.current) return;
    const conv = convRef.current;
    const msgs = messagesRef.current;
    let lastUser = -1;
    for (let i = msgs.length - 1; i >= 0; i--) {
      if (msgs[i].role === "user") {
        lastUser = i;
        break;
      }
    }
    if (!conv || lastUser < 0) return;
    const text = msgs[lastUser].content;
    sendingRef.current = true;
    setError(null);
    setStatus("sending");
    setMessages((prev) =>
      prev.filter((m, i) => !(i > lastUser && m.role === "assistant" && m.status === "error"))
    );
    await begin(conv, text, RetryMessage);
  }, [begin]);

  const clearError = useCallback(() => setError(null), []);

  const cancel = useCallback(async () => {
    // Saat sending, requestId belum ada: batalkan via backend lalu lepas guard.
    // finally: guard harus lepas walau CancelGeneration reject; kalau tidak,
    // tombol +/rename/delete/send terkunci permanen (busyNow() true selamanya).
    try {
      await CancelGeneration(convRef.current);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal membatalkan.");
    } finally {
      // Lepas kedua guard: sendingRef untuk fase "sending", activeRequest untuk
      // fase "streaming". Kalau salah satu tertinggal, send() dan aksi lain
      // terkunci permanen sampai restart.
      sendingRef.current = false;
      activeRequest.current = null;
      setStatus("idle");
    }
  }, []);

  const busyNow = useCallback(
    () => sendingRef.current || activeRequest.current !== null,
    []
  );

  const newConversation = useCallback(async () => {
    if (busyNow()) return;
    loadEmpty(freshId());
  }, [loadEmpty, busyNow]);

  const switchConversation = useCallback(
    async (id: string) => {
      if (id === convRef.current || busyNow()) return;
      await loadConversation(id);
    },
    [loadConversation, busyNow]
  );

  const renameConversation = useCallback(
    async (id: string, title: string) => {
      try {
        await RenameConversation(id, title);
        await refreshList();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal mengubah judul.");
      }
    },
    [refreshList]
  );

  const deleteConversation = useCallback(
    async (id: string) => {
      try {
        await DeleteConversation(id);
      } catch (e) {
        // Row belum ada (percakapan baru kosong): anggap sudah bersih.
        const msg = e instanceof Error ? e.message : "";
        if (!/no rows|tidak ditemukan/i.test(msg)) {
          setError(msg || "Gagal menghapus percakapan.");
          return;
        }
      }
      const rows = await refreshList();
      if (id === convRef.current) {
        const next = rows.find((c) => c.id !== id);
        if (next) {
          await loadConversation(next.id);
        } else {
          loadEmpty(freshId());
        }
      }
    },
    [loadConversation, loadEmpty, refreshList]
  );

  const busy = status === "sending" || status === "streaming";

  return {
    messages,
    status,
    error,
    busy,
    send,
    retry,
    clearError,
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
