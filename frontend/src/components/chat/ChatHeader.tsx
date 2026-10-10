// ChatHeader owns the chat-mode window header: brand mark + wordmark on the
// left, four icon actions on the right. The version lives in Settings > Tentang.
// Invariant: root keeps id="chat-header" + class "drag-handle" so the Wails
// window drags from the header; all buttons opt out via base.css automatically.
import type { JSX } from "react";
import { BrandMark } from "../ui/BrandMark";
import {
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
  return (
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
