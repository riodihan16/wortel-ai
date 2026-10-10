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

function itemLabel(c: ConversationItem): string {
  return c.title || c.id.slice(0, 8);
}

export function ConversationPicker({ conversations, conversationId, disabled, onSwitch }: Props): JSX.Element {
  const [open, setOpen] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  // A freshly created conversation is not in the list until its first message
  // is saved; keep it visible as the current entry.
  const items: ConversationItem[] = conversations.some((c) => c.id === conversationId)
    ? conversations
    : [{ id: conversationId, title: FRESH_LABEL }, ...conversations];
  const current = items.find((c) => c.id === conversationId);
  const currentLabel = current ? itemLabel(current) : FRESH_LABEL;

  const close = (refocus: boolean) => {
    setOpen(false);
    if (refocus) triggerRef.current?.focus();
  };

  useEffect(() => {
    if (!open) return;
    menuRef.current?.querySelector<HTMLElement>('[aria-checked="true"]')?.focus();
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
        <div className="menu conv-menu" role="menu" ref={menuRef} tabIndex={-1} onKeyDown={(e) => moveMenuFocus(e, menuRef.current)}>
          {items.map((c) => {
            const isOn = c.id === conversationId;
            return (
              <button
                key={c.id}
                type="button"
                role="menuitemradio"
                aria-checked={isOn}
                className={isOn ? "menu-option menu-option--on" : "menu-option"}
                title={itemLabel(c)}
                onClick={() => {
                  close(true);
                  if (!isOn) onSwitch(c.id);
                }}
              >
                <span className="menu-option__label">{itemLabel(c)}</span>
                {isOn && <IconCheck size={14} />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
