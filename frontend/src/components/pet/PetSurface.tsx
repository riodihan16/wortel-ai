// PetSurface owns the 180x180 always-on-top companion: a carrot button that
// opens chat, plus a hover/keyboard-revealed hide/quit rail.
// Invariant: the rail + hint reveal is pure CSS (#pet:hover / :focus-within),
// so there is deliberately no React hover state and the buttons stay tabbable.
//
// Status (docs 06 §3): the carrot itself carries it, no separate pip and no
// permanent text on the desktop. Shape, not colour, tells the states apart
// (WCAG 1.4.1): idle smiles; thinking looks up with a "…" thought bubble;
// error frowns with drooping leaves and a "!" badge. Screen readers get the
// same state through the .sr-only status text.
import { useEffect, useRef, useState } from "react";
import type { JSX } from "react";
import type { ChatStatus } from "../../types/chat";
import type { PetDragHandlers } from "../../hooks/usePetDrag";
import { Carrot, type CarrotMood } from "../ui/Carrot";
import { IconAlert, IconEyeOff, IconPower } from "../ui/Icon";

function moodFor(status: ChatStatus): CarrotMood {
  if (status === "sending" || status === "streaming") return "thinking";
  if (status === "error") return "error";
  return "idle";
}

const STATUS_TEXT: Record<CarrotMood, string> = {
  idle: "Siap",
  thinking: "Sedang menulis balasan",
  error: "Permintaan terakhir gagal",
};

export function PetSurface(props: {
  status: ChatStatus;
  drag: PetDragHandlers;
  onHide: () => void;
  onQuit: () => void;
}): JSX.Element {
  const { status, drag, onHide, onQuit } = props;
  const mood = moodFor(status);
  // Keluar asks first, inside the rail (ADR-050). Focus on Batal keeps the
  // rail open through :focus-within even after the pointer leaves.
  const [confirmQuit, setConfirmQuit] = useState(false);
  const quitRef = useRef<HTMLButtonElement>(null);
  const refocusQuit = useRef(false);
  useEffect(() => {
    if (!confirmQuit && refocusQuit.current) {
      refocusQuit.current = false;
      quitRef.current?.focus();
    }
  }, [confirmQuit]);
  const cancelQuit = () => {
    refocusQuit.current = true;
    setConfirmQuit(false);
  };

  return (
    <div id="pet" className={`pet--${mood}`}>
      <span className="pet__hint" aria-hidden="true">
        Klik: chat · Geser: pindah
      </span>
      <span className="sr-only" role="status">{STATUS_TEXT[mood]}</span>
      <button
        type="button"
        className="pet__btn"
        {...drag}
        aria-label="Buka chat"
        title="Klik untuk buka chat, geser untuk memindahkan"
      >
        <span className="pet__float">
          <Carrot size={116} mood={mood} />
          {mood === "thinking" && (
            <span className="pet__bubble" aria-hidden="true">
              <i />
              <i />
              <i />
            </span>
          )}
          {mood === "error" && (
            <span className="pet__alert" aria-hidden="true">
              <IconAlert size={11} strokeWidth={2.5} />
            </span>
          )}
        </span>
      </button>
      {confirmQuit ? (
        <div
          className="pet__actions pet__actions--confirm"
          role="group"
          aria-label="Konfirmasi keluar"
          onKeyDown={(e) => {
            if (e.key === "Escape") cancelQuit();
          }}
        >
          <span className="pet__confirm-msg">Keluar?</span>
          <button type="button" className="btn btn--sm btn--danger" onClick={onQuit}>
            Keluar
          </button>
          <button type="button" className="btn btn--sm" onClick={cancelQuit} autoFocus>
            Batal
          </button>
        </div>
      ) : (
        <div className="pet__actions">
          <button
            type="button"
            className="icon-btn"
            onClick={onHide}
            aria-label="Sembunyikan"
            title="Sembunyikan (buka lagi dari ikon Wortel di tray)"
          >
            <IconEyeOff />
          </button>
          <button
            ref={quitRef}
            type="button"
            className="icon-btn icon-btn--danger"
            onClick={() => setConfirmQuit(true)}
            aria-label="Keluar"
            title="Keluar dari Wortel AI"
          >
            <IconPower />
          </button>
        </div>
      )}
    </div>
  );
}
