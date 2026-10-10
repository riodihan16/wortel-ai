// ModelPicker owns the active-model switcher in #chat-footer.
// Renders null when no provider/models. Dropdown opens upward (footer is
// bottom row). Same menu contract as ConversationPicker: focus lands on the
// active model, arrows move, Escape returns focus to the trigger, Tab-out and
// click-outside close; no global handlers while closed.
import { useEffect, useRef, useState } from "react";
import type { JSX } from "react";
import type { ProviderProfile } from "../../types/provider";
import { IconCheck, IconChevronDown, IconModel } from "../ui/Icon";
import { moveMenuFocus } from "./menuKeys";

interface Props {
  provider: ProviderProfile | null;
  busy: boolean;
  onModelChange: (id: string, model: string) => void;
}

export function ModelPicker({ provider, busy, onModelChange }: Props): JSX.Element | null {
  const [open, setOpen] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    menuRef.current?.querySelector<HTMLElement>('[aria-checked="true"]')?.focus();
    const onDoc = (e: MouseEvent) => {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false);
    };
    const onEsc = (e: globalThis.KeyboardEvent) => {
      if (e.key === "Escape") {
        setOpen(false);
        triggerRef.current?.focus();
      }
    };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onEsc);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onEsc);
    };
  }, [open ]);

  if (!provider || provider.models.length === 0) return null;

  const active = provider.activeModel || provider.model;
  const activeEntry = provider.models.find((m) => m.id === active);
  const activeLabel = (activeEntry && activeEntry.alias) || active;

  return (
    <div
      className="model-picker"
      ref={boxRef}
      onBlur={(e) => {
        if (open && !boxRef.current?.contains(e.relatedTarget as Node | null)) setOpen(false);
      }}
    >
      <button
        ref={triggerRef}
        type="button"
        className="btn model-trigger"
        onClick={() => setOpen((v) => !v)}
        disabled={busy}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={`Model ${activeLabel}, provider ${provider.name}`}
        title={`Provider: ${provider.name}`}
      >
        <IconModel size={14} />
        <span className="model-trigger__label">{activeLabel}</span>
        <IconChevronDown size={14} />
      </button>
      {open && (
        <div
          className="menu model-menu"
          role="menu"
          ref={menuRef}
          tabIndex={-1}
          onKeyDown={(e) => moveMenuFocus(e, menuRef.current)}
        >
          {provider.models.map((m) => {
            const isOn = m.id === active;
            return (
              <button
                key={m.id}
                type="button"
                role="menuitemradio"
                aria-checked={isOn}
                className={isOn ? "menu-option menu-option--on" : "menu-option"}
                title={m.id}
                disabled={busy}
                onClick={() => {
                  setOpen(false);
                  triggerRef.current?.focus();
                  if (!isOn) onModelChange(provider.id, m.id);
                }}
              >
                <span className="menu-option__label">{m.alias || m.id}</span>
                {isOn && <IconCheck size={14} />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
