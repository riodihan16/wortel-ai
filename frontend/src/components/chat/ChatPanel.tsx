// ChatPanel owns #chat-main: conv-bar, list, error, composer, footer.
// Invariants: prop interface frozen; draft clears only on success (submit);
// error never clears draft. Rename and delete-confirm rows are mutually
// exclusive, Escape cancels either; no window-level Escape. Delete asks inline,
// never through window.confirm: WebView2 draws that natively, outside the theme.
import { useEffect, useRef, useState } from "react";
import type { JSX, KeyboardEvent, ReactNode } from "react";
import type { ChatMessage, ChatStatus } from "../../types/chat";
import type { ConversationItem } from "../../hooks/useChat";
import { ConversationPicker } from "./ConversationPicker";
import { SafeMarkdown } from "./SafeMarkdown";
import {
  IconAlert,
  IconArrowUp,
  IconPencil,
  IconPlus,
  IconSparkle,
  IconStop,
  IconTrash,
} from "../ui/Icon";

interface Props {
  messages: ChatMessage[];
  status: ChatStatus;
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
  modelPicker: ReactNode;
  onOpenSettings: () => void;
  hasProvider: boolean;
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

function statusDot(s: ChatStatus): string {
  if (s === "sending" || s === "streaming") return "status-dot status-dot--busy";
  if (s === "error") return "status-dot status-dot--error";
  if (s === "complete") return "status-dot status-dot--ok";
  return "status-dot";
}

export function ChatPanel({
  messages,
  status,
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
  modelPicker,
  onOpenSettings,
  hasProvider,
}: Props): JSX.Element {
  const [renaming, setRenaming] = useState<string | null>(null);
  const [renameDraft, setRenameDraft] = useState("");
  const [deleting, setDeleting] = useState<string | null>(null);
  const deleteBtnRef = useRef<HTMLButtonElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const composerRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const el = listRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages]);

  useEffect(() => {
    composerRef.current?.focus();
  }, []);

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

  const cancelDelete = () => {
    setDeleting(null);
    deleteBtnRef.current?.focus();
  };

  const confirmDelete = (id: string) => {
    setDeleting(null);
    onDelete(id);
    // The focused Hapus button unmounts with the row; hand focus to the
    // composer instead of letting it fall to <body>.
    composerRef.current?.focus();
  };

  // The row names one conversation; switching away must not leave it armed
  // against a conversation the user is no longer looking at. `armed` also
  // guards the one render between the switch and this effect's reset.
  useEffect(() => {
    setDeleting(null);
  }, [conversationId]);
  const armed = deleting === conversationId ? deleting : null;

  const cur = conversations.find((c) => c.id === conversationId);
  const curTitle = cur?.title || conversationId.slice(0, 8);
  const label = statusLabel(status);

  return (
    <div id="chat-main">
      <div id="conv-bar">
        <ConversationPicker
          conversations={conversations}
          conversationId={conversationId}
          disabled={busy}
          onSwitch={onSwitch}
        />
        <button type="button" className="icon-btn" onClick={onNew} title="Percakapan baru" aria-label="Percakapan baru" disabled={busy}>
          <IconPlus size={15} />
        </button>
        <button
          type="button"
          className="icon-btn"
          onClick={() => {
            setDeleting(null);
            setRenameDraft(cur?.title ?? "");
            setRenaming(conversationId);
          }}
          title="Ubah judul"
          aria-label="Ubah judul"
          disabled={busy || !conversationId}
        >
          <IconPencil size={15} />
        </button>
        <button
          ref={deleteBtnRef}
          type="button"
          className="icon-btn icon-btn--danger"
          onClick={() => {
            setRenaming(null);
            setDeleting(conversationId);
          }}
          title="Hapus percakapan"
          aria-label="Hapus percakapan"
          aria-expanded={armed !== null}
          disabled={busy || !conversationId}
        >
          <IconTrash size={15} />
        </button>
      </div>
      {armed && (
        <div
          id="conv-delete"
          className="confirm-row"
          role="group"
          aria-label="Konfirmasi hapus percakapan"
          onKeyDown={(e) => {
            if (e.key === "Escape") cancelDelete();
          }}
        >
          <IconAlert size={14} />
          <span className="confirm-row__msg">
            <span>Hapus permanen “</span>
            <span className="confirm-row__title">{curTitle}</span>
            <span>”?</span>
          </span>
          <button type="button" className="btn btn--sm btn--danger" onClick={() => confirmDelete(armed)} disabled={busy}>
            Hapus
          </button>
          {/* Focused first: Enter on an armed row must not delete. */}
          <button type="button" className="btn btn--sm" onClick={cancelDelete} autoFocus>
            Batal
          </button>
        </div>
      )}
      {renaming && (
        <div id="conv-rename">
          <input
            className="input"
            value={renameDraft}
            onChange={(e) => setRenameDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Escape") setRenaming(null);
              if (e.key === "Enter") commitRename(renaming);
            }}
            placeholder="Judul baru…"
            maxLength={120}
            aria-label="Judul baru"
          />
          <button type="button" className="btn btn--sm" onClick={() => commitRename(renaming)}>
            OK
          </button>
          <button type="button" className="btn btn--sm" onClick={() => setRenaming(null)}>
            Batal
          </button>
        </div>
      )}
      <div id="chat-list" ref={listRef}>
        {messages.length === 0 && !hasProvider && (
          <div className="chat-empty">
            <IconSparkle size={18} />
            <p className="chat-empty__heading">Belum ada provider</p>
            <p className="chat-empty__hint">
              Tambahkan base URL, daftar model, dan API key lebih dulu. Chat aktif
              setelah ada provider yang dipilih.
            </p>
            <button type="button" className="btn btn--primary" onClick={onOpenSettings}>
              Buka Pengaturan
            </button>
          </div>
        )}
        {messages.length === 0 && hasProvider && (
          <div className="chat-empty">
            <IconSparkle size={18} />
            <p className="chat-empty__heading">Belum ada percakapan</p>
            <p className="chat-empty__hint">Mulai dengan satu pertanyaan singkat.</p>
            <p className="chat-empty__example-label">Contoh prompt</p>
            <p className="chat-empty__example">Ringkas catatan ini jadi 3 poin.</p>
          </div>
        )}
        {messages.map((m) => (
          <div key={m.id} className={`msg msg-${m.role}`}>
            <div className="msg-body">
              {m.role === "assistant" ? <SafeMarkdown text={m.content} /> : m.content}
            </div>
            {statusLabel(m.status) ? (
              <div className="msg-status">{statusLabel(m.status)}</div>
            ) : (
              <div className="msg-status" aria-hidden="true" />
            )}
          </div>
        ))}
      </div>
      {error && (
        <div id="chat-error" className="banner banner--error" role="alert">
          <IconAlert size={15} />
          <span className="banner__text">{error}</span>
          <button type="button" className="btn btn--sm" onClick={onRetry}>
            Coba lagi
          </button>
        </div>
      )}
      <div id="chat-composer">
        <textarea
          ref={composerRef}
          className="textarea"
          value={draft}
          onChange={(e) => onDraftChange(e.target.value)}
          onKeyDown={keyDown}
          placeholder="Tulis pesan… (Enter kirim, Shift+Enter baris baru)"
          rows={2}
          aria-label="Tulis pesan"
        />
        <button
          type="button"
          id="btn-send"
          className={busy ? "btn btn--icon btn--danger" : "btn btn--icon btn--primary"}
          onClick={() => {
            if (busy) onCancel();
            else submit();
          }}
          disabled={!draft.trim() && !busy}
          aria-label={busy ? "Hentikan" : "Kirim pesan"}
          title={busy ? "Hentikan" : "Kirim pesan"}
        >
          {busy ? <IconStop size={14} /> : <IconArrowUp size={14} />}
        </button>
      </div>
      <div id="chat-footer">
        <div id="chat-footer-model">{modelPicker}</div>
        <span id="chat-state">
          <span className={statusDot(status)} aria-hidden="true" />
          {label && <span className="chat-state__label">{label}</span>}
        </span>
      </div>
    </div>
  );
}
