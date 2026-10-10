// ChatHeader owns the chat-mode window header: brand mark + wordmark on the
// left, four icon actions on the right. The version lives in Settings > Tentang.
// Keluar asks inline first (ADR-050): it sits next to Sembunyikan, and
// quitting stops any reply in flight.
// Invariant: root keeps id="chat-header" + class "drag-handle" so the Wails
// window drags from the header; all buttons opt out via base.css automatically.
import { useRef, useState } from "react";
import type { JSX } from "react";
import { BrandMark } from "../ui/BrandMark";
import {
  IconAlert,
  IconEyeOff,
  IconMinimize,
  IconPower,
  IconSliders,
} from "../ui/Icon";

interface Props {
  onSettings: () => void;
  onPet: () => void;
  onHide: () => void;
  onQuit: () => void;
}

export function ChatHeader({
  onSettings,
  onPet,
  onHide,
  onQuit,
}: Props): JSX.Element {
  const [confirmQuit, setConfirmQuit] = useState(false);
  const quitRef = useRef<HTMLButtonElement>(null);
  const cancelQuit = () => {
    setConfirmQuit(false);
    quitRef.current?.focus();
  };

  return (
    <>
      <div id="chat-header" className="drag-handle">
        <div className="brand">
          <BrandMark />
          <span className="brand__name">Wortel AI</span>
        </div>
        <div className="header-actions">
          <button
            type="button"
            className="icon-btn"
            onClick={onSettings}
            aria-label="Pengaturan"
            title="Pengaturan"
          >
            <IconSliders size={15} />
          </button>
          <button
            type="button"
            className="icon-btn"
            onClick={onPet}
            aria-label="Kecilkan jadi wortel"
            title="Kecilkan jadi wortel"
          >
            <IconMinimize size={15} />
          </button>
          <button
            type="button"
            className="icon-btn"
            onClick={onHide}
            aria-label="Sembunyikan"
            title="Sembunyikan (buka lagi dari ikon Wortel di tray)"
          >
            <IconEyeOff size={15} />
          </button>
          <button
            ref={quitRef}
            type="button"
            className="icon-btn icon-btn--danger"
            onClick={() => setConfirmQuit(true)}
            aria-label="Keluar"
            aria-expanded={confirmQuit}
            title="Keluar dari Wortel AI"
          >
            <IconPower size={15} />
          </button>
        </div>
      </div>
      {confirmQuit && (
        <div
          id="quit-confirm"
          className="confirm-row"
          role="group"
          aria-label="Konfirmasi keluar"
          onKeyDown={(e) => {
            if (e.key === "Escape") cancelQuit();
          }}
        >
          <IconAlert size={14} />
          <span className="confirm-row__msg">Keluar dari Wortel AI?</span>
          <button type="button" className="btn btn--sm btn--danger" onClick={onQuit}>
            Keluar
          </button>
          {/* Focused first: Enter on the armed row must not quit. */}
          <button type="button" className="btn btn--sm" onClick={cancelQuit} autoFocus>
            Batal
          </button>
        </div>
      )}
    </>
  );
}
