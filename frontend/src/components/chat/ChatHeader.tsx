// ChatHeader owns the chat-mode window header: brand mark + wordmark +
// version pill on the left, four icon actions on the right.
// Invariant: root keeps id="chat-header" + class "drag-handle" so the Wails
// window drags from the header; all buttons opt out via base.css automatically.
import type { JSX } from "react";
import {
  IconEyeOff,
  IconMinimize,
  IconPower,
  IconSliders,
} from "../ui/Icon";

interface Props {
  version: string;
  onSettings: () => void;
  onPet: () => void;
  onHide: () => void;
  onQuit: () => void;
}

export function ChatHeader({
  version,
  onSettings,
  onPet,
  onHide,
  onQuit,
}: Props): JSX.Element {
  return (
    <div id="chat-header" className="drag-handle">
      <div className="brand">
        <svg
          className="brand__mark"
          width="18"
          height="18"
          viewBox="0 0 18 18"
          aria-hidden="true"
        >
          <path
            d="M9 5.2 C6.6 4.2 5.4 2.4 5.2 0.6 C7 1 8.6 2.8 9 5.2 Z"
            fill="var(--color-brand-leaf)"
          />
          <path
            d="M9 5.2 C9 3.4 9.2 1.7 10 0.2 C10.9 1.8 10.4 3.8 9 5.2 Z"
            fill="var(--color-brand-leaf)"
          />
          <path
            d="M9 5.2 C11 4.4 12.4 2.9 12.7 1.1 C11 1.4 9.5 3 9 5.2 Z"
            fill="var(--color-brand-leaf)"
          />
          <path
            d="M4.8 4.6 L13.2 4.6 C13.2 8.6 11.4 12.6 9.3 17.1 C9.1 17.6 8.9 17.6 8.7 17.1 C6.6 12.6 4.8 8.6 4.8 4.6 Z"
            fill="var(--color-brand)"
          />
        </svg>
        <span className="brand__name">Wortel AI</span>
        <span className="brand__version">v{version}</span>
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
          type="button"
          className="icon-btn icon-btn--danger"
          onClick={onQuit}
          aria-label="Keluar"
          title="Keluar dari Wortel AI"
        >
          <IconPower size={15} />
        </button>
      </div>
    </div>
  );
}
