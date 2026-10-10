import { useEffect, useRef, useState } from "react";
import type { ProviderProfile } from "../../types/provider";

interface Props {
  provider: ProviderProfile | null;
  busy: boolean;
  onModelChange: (id: string, model: string) => void;
}

// ModelPicker is the chat-header control for switching the active model.
// Renders nothing when no provider is active or it has a single model.
// The dropdown is a button + list, not a native <select>, so it fits the
// header and stays clickable inside the drag region.
export function ModelPicker({ provider, busy, onModelChange }: Props) {
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
  }, [open]);

  if (!provider || provider.models.length === 0) return null;

  const active = provider.activeModel || provider.model;
  const activeEntry = provider.models.find((m) => m.id === active);
  const activeLabel = (activeEntry && activeEntry.alias) || active;

  return (
    <div id="model-picker" ref={boxRef}>
      <button
        type="button"
        id="btn-model"
        onClick={() => setOpen((v) => !v)}
        disabled={busy}
        title={`Provider: ${provider.name}`}
      >
        {provider.models.length > 1 ? `Model: ${activeLabel} ▾` : `Model: ${activeLabel}`}
      </button>
      {open && (
        <div id="model-menu" role="menu">
          {provider.models.map((m) => (
            <button
              key={m.id}
              type="button"
              role="menuitemradio"
              aria-checked={m.id === active}
              className={`model-option${m.id === active ? " model-option-on" : ""}`}
              onClick={() => {
                setOpen(false);
                if (m.id !== active) onModelChange(provider.id, m.id);
              }}
            >
              {m.alias || m.id}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
