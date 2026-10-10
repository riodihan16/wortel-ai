// ConversationPicker owns the conversation switcher in #conv-bar. It replaces
// a native <select>, whose option popup WebView2 draws with system chrome that
// ignores the theme. Same menu contract as ModelPicker (Escape + click-outside
// close, no global handlers while closed), opening downward, plus arrow-key
// focus movement because the list can grow long.
import { useEffect, useRef, useState } from "react";
import type { JSX } from "react";
import type { ConversationItem } from "../../hooks/useChat";
import { IconCheck, IconChevronDown } from "../ui/Icon";
import { moveMenuFocus } from "./menuKeys";

interface Props {
  conversations: ConversationItem[];
  conversationId: string;
  disabled: boolean;
  onSwitch: (id: string) => void;
}

const FRESH_LABEL = "Percakapan baru";

// Past this many entries a search box leads the menu: scanning 30 look-alike
// titles is recall work; typing two letters is recognition.
const SEARCH_FROM = 7;

function itemLabel(c: ConversationItem): string {
  return c.title || c.id.slice(0, 8);
}

export function ConversationPicker({ conversations, conversationId, disabled, onSwitch }: Props): JSX.Element {
  const [open, setOpen] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState("");

  // A freshly created conversation is not in the list until its first message
  // is saved; keep it visible as the current entry.
  const items: ConversationItem[] = conversations.some((c) => c.id === conversationId)
    ? conversations
    : [{ id: conversationId, title: FRESH_LABEL }, ...conversations];
  const current = items.find((c) => c.id === conversationId);
  const currentLabel = current ? itemLabel(current) : FRESH_LABEL;
  const searchable = items.length >= SEARCH_FROM;
  const needle = query.trim().toLowerCase();
  const shown = needle ? items.filter((c) => itemLabel(c).toLowerCase().includes(needle)) : items;

  const pick = (c: ConversationItem) => {
    close(true);
    if (c.id !== conversationId) onSwitch(c.id);
  };

  const close = (refocus: boolean) => {
    setOpen(false);
    if (refocus) triggerRef.current?.focus();
  };

  useEffect(() => {
    if (!open) return;
    setQuery("");
    (searchRef.current ?? menuRef.current?.querySelector<HTMLElement>('[aria-checked="true"]'))?.focus();
    const onDoc = (e: MouseEvent) => {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false);
    };
    const onEsc = (e: globalThis.KeyboardEvent) => {
      if (e.key === "Escape") close(true);
    };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onEsc);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onEsc);
    };
  }, [open]);

  useEffect(() => {
    if (disabled) setOpen(false);
  }, [disabled]);


  return (
    <div
      className="conv-picker"
      ref={boxRef}
      onBlur={(e) => {
        // Tab out of the open menu: close it rather than leave a menu open
        // behind focus. Moves between options stay inside the box.
        if (open && !boxRef.current?.contains(e.relatedTarget as Node | null)) setOpen(false);
      }}
    >
      <button
        ref={triggerRef}
        type="button"
        className="btn conv-trigger"
        onClick={() => setOpen((v) => !v)}
        disabled={disabled}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={`Pilih percakapan: ${currentLabel}`}
        title={currentLabel}
      >
        <span className="conv-trigger__label">{currentLabel}</span>
        <IconChevronDown size={14} />
      </button>
      {open && (
        // tabIndex -1: a press on the menu's padding or scrollbar focuses the
        // menu itself instead of <body>, so the blur rule above keeps it open.
        // The search box and empty note sit outside role="menu", which may
        // only own menu items; arrows from the box are forwarded to the list.
        <div className="menu conv-menu">
          {searchable && (
            <input
              ref={searchRef}
              type="search"
              className="input conv-search"
              aria-label="Cari percakapan"
              placeholder="Cari judul…"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "ArrowDown" || e.key === "ArrowUp") moveMenuFocus(e, menuRef.current);
                // Enter opens the first match only once the user has typed.
                if (e.key === "Enter" && needle && shown.length > 0) {
                  e.preventDefault();
                  pick(shown[0]);
                }
              }}
            />
          )}
          {shown.length === 0 && <p className="conv-empty">Tidak ada percakapan yang cocok.</p>}
          <div className="conv-menu__list" role="menu" ref={menuRef} tabIndex={-1} onKeyDown={(e) => moveMenuFocus(e, menuRef.current)}>
            {shown.map((c) => {
              const isOn = c.id === conversationId;
              return (
                <button
                  key={c.id}
                  type="button"
                  role="menuitemradio"
                  aria-checked={isOn}
                  className={isOn ? "menu-option menu-option--on" : "menu-option"}
                  title={itemLabel(c)}
                  onClick={() => pick(c)}
                >
                  <span className="menu-option__label">{itemLabel(c)}</span>
                  {isOn && <IconCheck size={14} />}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
