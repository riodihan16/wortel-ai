// ModelPicker owns the active-model switcher in #chat-footer.
// Renders null when no provider/models. Dropdown opens upward (footer is
// bottom row). Escape + click-outside close; no global handlers.
import { useEffect, useRef, useState } from "react";
import type { JSX } from "react";
import type { ProviderProfile } from "../../types/provider";
import { IconCheck, IconChevronDown, IconGlobe } from "../ui/Icon";

interface Props {
  provider: ProviderProfile | null;
  busy: boolean;
  onModelChange: (id: string, model: string) => void;
}

export function ModelPicker({ provider, busy, onModelChange }: Props): JSX.Element | null {
  const [open, setOpen] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false);
    };
    const onEsc = (e: globalThis.KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
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
    <div className="model-picker" ref={boxRef}>
      <button
        type="button"
        className="btn model-trigger"
        onClick={() => setOpen((v) => !v)}
        disabled={busy}
        aria-haspopup="menu"
        aria-expanded={open}
        title={`Provider: ${provider.name}`}
      >
        <IconGlobe size={14} />
        <span className="model-trigger__label">{activeLabel}</span>
        <IconChevronDown size={14} />
      </button>
      {open && (
        <div className="menu model-menu" role="menu">
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
