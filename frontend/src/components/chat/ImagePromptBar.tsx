// ComposerPlusMenu is the composer's "+" button: a single entry point whose
// menu offers "Upload file" (the attachment picker) and "Buat gambar" (toggle
// image mode). It replaces the two separate clip + image buttons so the
// composer stays uncluttered. Same menu contract as ModelPicker: Escape +
// click-outside close, focus returns to the trigger, no global handlers while
// closed. The image-mode check mark shows whether the composer sends chat or
// image prompts.
import { useEffect, useRef, useState } from "react";
import type { JSX } from "react";
import { IconCheck, IconModel, IconPaperclip, IconPlus } from "../ui/Icon";
import { moveMenuFocus } from "./menuKeys";

interface Props {
  /** Attachment picking is unavailable (busy or no provider). */
  attachDisabled: boolean;
  /** Image mode is on. */
  imageActive: boolean;
  /** A model image is configured. */
  hasImageModel: boolean;
  onUpload: () => void;
  onToggleImage: () => void;
}

export function ComposerPlusMenu({
  attachDisabled,
  imageActive,
  hasImageModel,
  onUpload,
  onToggleImage,
}: Props): JSX.Element {
  const [open, setOpen] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    menuRef.current?.querySelector<HTMLElement>('[role="menuitem"]')?.focus();
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
  }, [open]);

  const close = () => {
    setOpen(false);
    triggerRef.current?.focus();
  };

  return (
    <div
      className="composer-plus"
      ref={boxRef}
      onBlur={(e) => {
        if (open && !boxRef.current?.contains(e.relatedTarget as Node | null)) setOpen(false);
      }}
    >
      <button
        ref={triggerRef}
        type="button"
        id="btn-plus"
        className={imageActive ? "btn btn--icon btn--primary" : "btn btn--icon"}
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label="Tambah lampiran atau buat gambar"
        title="Tambah"
      >
        <IconPlus size={15} />
      </button>
      {open && (
        <div
          className="menu plus-menu"
          role="menu"
          ref={menuRef}
          tabIndex={-1}
          onKeyDown={(e) => moveMenuFocus(e, menuRef.current)}
        >
          <button
            type="button"
            role="menuitem"
            className="menu-option"
            disabled={attachDisabled}
            title={attachDisabled ? "Tidak tersedia saat ini" : "Lampirkan gambar, PDF, atau teks"}
            onClick={() => {
              close();
              onUpload();
            }}
          >
            <IconPaperclip size={14} />
            <span className="menu-option__label">Upload file</span>
          </button>
          <button
            type="button"
            role="menuitem"
            className={imageActive ? "menu-option menu-option--on" : "menu-option"}
            title={hasImageModel ? undefined : "Belum ada model gambar: atur di Pengaturan"}
            onClick={() => {
              close();
              onToggleImage();
            }}
          >
            <IconModel size={14} />
            <span className="menu-option__label">Buat gambar</span>
            {imageActive && <IconCheck size={14} />}
          </button>
        </div>
      )}
    </div>
  );
}
