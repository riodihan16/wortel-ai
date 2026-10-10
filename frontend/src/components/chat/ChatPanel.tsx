// ChatPanel owns #chat-main: conv-bar, list, error, composer, footer.
// Invariants: draft clears only on success (submit); error never clears
// draft. A failed reply shows its own reason inline (MessageItem); the
// banner carries errors that belong to no reply. Escape in the composer stops a
// running reply. Rename and delete-confirm rows are mutually
// exclusive, Escape cancels either; no window-level Escape. Delete asks inline,
// never through window.confirm: WebView2 draws that natively, outside the theme.
import { useEffect, useRef, useState } from "react";
import type { JSX, KeyboardEvent, ReactNode } from "react";
import type { ChatMessage, ChatStatus } from "../../types/chat";
import type { ConversationItem } from "../../hooks/useChat";
import { ConversationPicker } from "./ConversationPicker";
import { MessageItem } from "./MessageItem";
import { Carrot } from "../ui/Carrot";
import {
  IconAlert,
  IconArrowUp,
  IconChevronDown,
  IconPencil,
  IconPlus,
  IconStop,
  IconTrash,
  IconX,
} from "../ui/Icon";

// Inserted into the composer, never sent: the user still presses Enter.
const EXAMPLE_PROMPT = "Ringkas catatan ini jadi 3 poin:\n";

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
  onDismissError: () => void;
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

// Footer status speaks only while work is in flight; outcomes (failed,
// cancelled) are shown on the reply itself.
function busyLabel(s: ChatStatus): string {
  if (s === "sending") return "Mengirim…";
  if (s === "streaming") return "Wortel mengetik…";
  return "";
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
  onDismissError,
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

  // Follow the reply only while the user is at the bottom; reading earlier
  // text during a stream must not be yanked away. A new conversation and the
  // user's own send always land at the bottom.
  const stick = useRef(true);
  const lastCount = useRef(messages.length);
  const [newBelow, setNewBelow] = useState(false);

  useEffect(() => {
    stick.current = true;
  }, [conversationId]);

  useEffect(() => {
    const el = listRef.current;
    const grew = messages.length > lastCount.current;
    lastCount.current = messages.length;
    if (!el) return;
    if (stick.current || (grew && messages.at(-1)?.role === "user")) {
      el.scrollTop = el.scrollHeight;
      stick.current = true;
      setNewBelow(false);
    } else {
      setNewBelow(true);
    }
  }, [messages]);

  const onListScroll = () => {
    const el = listRef.current;
    if (!el) return;
    const near = el.scrollHeight - el.scrollTop - el.clientHeight < 48;
    stick.current = near;
    if (near) setNewBelow(false);
  };

  const toBottom = () => {
    const el = listRef.current;
    if (el) el.scrollTop = el.scrollHeight;
    stick.current = true;
    setNewBelow(false);
  };

  useEffect(() => {
    composerRef.current?.focus();
  }, []);

  const submit = () => {
    if (!draft.trim() || busy || !hasProvider) return;
    onSend(draft);
    onDraftChange("");
  };

  const keyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Escape" && busy) {
      e.preventDefault();
      onCancel();
      return;
    }
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
  // A new conversation exists only locally until its first message is
  // saved: nothing to rename or delete yet.
  const saved = cur !== undefined;
  const curTitle = cur?.title || conversationId.slice(0, 8);
  const label = busyLabel(status);
  const last = messages.at(-1);
  // SendMessage itself threw: the question is on screen with no reply yet.
  const sendFailed = status === "error" && last?.role === "user";

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
          disabled={busy || !saved}
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
          disabled={busy || !saved}
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
      <div id="chat-list" ref={listRef} role="log" aria-label="Isi percakapan" aria-busy={busy} onScroll={onListScroll}>
        {messages.length === 0 && !hasProvider && (
          <div className="chat-empty">
            <Carrot size={64} />
            <p className="chat-empty__heading">Wortel belum tersambung</p>
            <p className="chat-empty__hint">
              Hubungkan ke OpenAI, OpenRouter, Groq, atau model lokal supaya
              Wortel bisa membalas.
            </p>
            <button type="button" className="btn btn--primary" onClick={onOpenSettings}>
              Hubungkan provider
            </button>
          </div>
        )}
        {messages.length === 0 && hasProvider && (
          <div className="chat-empty">
            <Carrot size={64} />
            <p className="chat-empty__heading">Mau tanya apa?</p>
            <p className="chat-empty__hint">Tulis di bawah, atau mulai dari contoh ini:</p>
            <button
              type="button"
              className="chat-empty__example"
              onClick={() => {
                onDraftChange(EXAMPLE_PROMPT);
                composerRef.current?.focus();
              }}
            >
              Ringkas catatan ini jadi 3 poin
            </button>
          </div>
        )}
        {messages.map((m, i) => (
          <MessageItem
            key={m.id}
            message={m}
            isLast={i === messages.length - 1}
            busy={busy}
            onRetry={onRetry}
          />
        ))}
      </div>
      {newBelow && (
        <button type="button" className="btn btn--sm chat-newbelow" onClick={toBottom} aria-label="Balasan baru di bawah">
          <IconChevronDown size={13} />
          Balasan baru
        </button>
      )}
      {error && (
        <div id="chat-error" className="banner banner--error" role="alert">
          <IconAlert size={15} />
          <span className="banner__text">{error}</span>
          {sendFailed && !busy && (
            <button type="button" className="btn btn--sm" onClick={onRetry}>
              Coba lagi
            </button>
          )}
          <button type="button" className="icon-btn" onClick={onDismissError} aria-label="Tutup pesan error" title="Tutup">
            <IconX size={14} />
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
          placeholder={hasProvider ? "Tulis pesan… (Enter kirim, Shift+Enter baris baru)" : "Hubungkan provider dulu untuk mulai chat"}
          rows={2}
          aria-label="Tulis pesan"
          disabled={!hasProvider}
        />
        <button
          type="button"
          id="btn-send"
          className={busy ? "btn btn--icon btn--stop" : "btn btn--icon btn--primary"}
          onClick={() => {
            if (busy) onCancel();
            else submit();
          }}
          disabled={busy ? false : !hasProvider || !draft.trim()}
          aria-label={busy ? "Hentikan balasan" : "Kirim pesan"}
          title={busy ? "Hentikan balasan (Esc)" : "Kirim pesan (Enter)"}
        >
          {busy ? <IconStop size={14} /> : <IconArrowUp size={14} />}
        </button>
      </div>
      <div id="chat-footer">
        <div id="chat-footer-model">{modelPicker}</div>
        <span id="chat-state" role="status">
          {label && (
            <>
              <Carrot size={26} mood="thinking" className="chat-state__carrot" />
              <span className="chat-state__label">{label}</span>
            </>
          )}
        </span>
      </div>
    </div>
  );
}
