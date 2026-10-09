import { useEffect, useRef, useState } from "react";
import type { KeyboardEvent } from "react";
import type { ChatMessage, ChatStatus } from "../../types/chat";

interface Props {
  messages: ChatMessage[];
  status: ChatStatus;
  error: string | null;
  busy: boolean;
  onSend: (text: string) => void;
  onCancel: () => void;
  onRetry: () => void;
}

function statusLabel(s: ChatStatus): string {
  switch (s) {
    case "sending":
      return "mengirim…";
    case "streaming":
      return "mengetik…";
    case "cancelled":
      return "dibatalkan";
    case "error":
      return "gagal";
    case "complete":
      return "";
    default:
      return "";
  }
}

export function ChatPanel({ messages, status, error, busy, onSend, onCancel, onRetry }: Props) {
  const [draft, setDraft] = useState("");
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = listRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages]);

  const submit = () => {
    if (!draft.trim() || busy) return;
    onSend(draft);
    setDraft("");
  };

  const keyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  };

  return (
    <div id="chat-main">
      <div id="chat-list" ref={listRef}>
        {messages.length === 0 && (
          <p id="chat-empty">Halo! Kirim pesan untuk mulai. Mock provider, tanpa biaya.</p>
        )}
        {messages.map((m) => (
          <div key={m.id} className={`msg msg-${m.role}`}>
            <div className="msg-body">{m.content}</div>
            {statusLabel(m.status) && <div className="msg-status">{statusLabel(m.status)}</div>}
          </div>
        ))}
      </div>
      {error && (
        <div id="chat-error" role="alert">
          <span>{error}</span>
          <button onClick={onRetry}>Coba lagi</button>
        </div>
      )}
      <div id="chat-composer">
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={keyDown}
          placeholder="Tulis pesan… (Enter kirim, Shift+Enter baris baru)"
          rows={2}
        />
        {busy ? (
          <button onClick={onCancel} id="btn-stop">
            Stop
          </button>
        ) : (
          <button onClick={submit} id="btn-send" disabled={!draft.trim()}>
            Kirim
          </button>
        )}
      </div>
      <div id="chat-state">status: {status}</div>
    </div>
  );
}
