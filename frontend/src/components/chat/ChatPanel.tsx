import { useEffect, useRef, useState } from "react";
import type { KeyboardEvent } from "react";
import type { ChatMessage, ChatStatus } from "../../types/chat";
import type { ConversationItem } from "../../hooks/useChat";
import { SafeMarkdown } from "./SafeMarkdown";

interface Props {
  messages: ChatMessage[];
  error: string | null;
  busy: boolean;
  conversations: ConversationItem[];
  conversationId: string;
  onSend: (text: string) => void;
  onCancel: () => void;
  onRetry: () => void;
  onNew: () => void;
  onSwitch: (id: string) => void;
  onRename: (id: string, title: string) => void;
  onDelete: (id: string) => void;
  draft: string;
  onDraftChange: (v: string) => void;
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

export function ChatPanel({
  messages,
  error,
  busy,
  conversations,
  conversationId,
  onSend,
  onCancel,
  onRetry,
  onNew,
  onSwitch,
  onRename,
  onDelete,
  draft,
  onDraftChange,
}: Props) {
  const [renaming, setRenaming] = useState<string | null>(null);
  const [renameDraft, setRenameDraft] = useState("");
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = listRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages]);

  const submit = () => {
    if (!draft.trim() || busy) return;
    onSend(draft);
    onDraftChange("");
  };

  const keyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  };

  const commitRename = (id: string) => {
    const t = renameDraft.trim();
    if (t) onRename(id, t);
    setRenaming(null);
  };

  return (
    <div id="chat-main">
      <div id="conv-bar">
        <select
          id="conv-select"
          value={conversationId}
          onChange={(e) => onSwitch(e.target.value)}
          title="Pilih percakapan"
          disabled={busy}
        >
          {/* A freshly created conversation has an id that is not in the list
              yet. Without a matching option the browser silently forces the
              select to the first entry, and picking that entry then fires no
              change event. Keep a placeholder option for the current id. */}
          {!conversations.some((c) => c.id === conversationId) && (
            <option value={conversationId}>Percakapan baru</option>
          )}
          {conversations.map((c) => (
            <option key={c.id} value={c.id}>
              {c.title || c.id.slice(0, 8)}
            </option>
          ))}
        </select>
        <button onClick={onNew} title="Percakapan baru" disabled={busy}>
          +
        </button>
        <button
          onClick={() => {
            const cur = conversations.find((c) => c.id === conversationId);
            setRenameDraft(cur?.title ?? "");
            setRenaming(conversationId);
          }}
          title="Ubah judul"
          disabled={busy || !conversationId}
        >
          ✎
        </button>
        <button
          onClick={() => {
            if (window.confirm("Hapus percakapan ini?")) onDelete(conversationId);
          }}
          title="Hapus percakapan"
          disabled={busy || !conversationId}
        >
          🗑
        </button>
      </div>
      {renaming && (
        <div id="conv-rename">
          <input
            value={renameDraft}
            onChange={(e) => setRenameDraft(e.target.value)}
            placeholder="Judul baru…"
            maxLength={120}
          />
          <button onClick={() => commitRename(renaming)}>OK</button>
          <button onClick={() => setRenaming(null)}>Batal</button>
        </div>
      )}
      <div id="chat-list" ref={listRef}>
        {messages.length === 0 && (
          <p id="chat-empty">Halo! Tambahkan provider di Settings, lalu kirim pesan.</p>
        )}
        {messages.map((m) => (
          <div key={m.id} className={`msg msg-${m.role}`}>
            <div className="msg-body">
              {m.role === "assistant" ? <SafeMarkdown text={m.content} /> : m.content}
            </div>
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
          onChange={(e) => onDraftChange(e.target.value)}
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
    </div>
  );
}
