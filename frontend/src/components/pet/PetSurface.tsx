// PetSurface owns the 180x180 always-on-top companion: a carrot button that
// opens chat, plus a hover/keyboard-revealed hide/quit rail.
// Invariant: the rail + hint reveal is pure CSS (#pet:hover / :focus-within),
// so there is deliberately no React hover state and the buttons stay tabbable.
//
// Status indicator (docs 06 §3): one 14px badge, bottom-right, no text on the
// desktop. Shape carries the meaning, not just colour — hollow ring = idle,
// filled disc + pulse = thinking, triangle = error (WCAG 1.4.1). The state is
// also announced to screen readers through .sr-only text.
import type { JSX } from "react";
import type { ChatStatus } from "../../types/chat";
import type { PetDragHandlers } from "../../hooks/usePetDrag";
import { IconAlert, IconEyeOff, IconPower } from "../ui/Icon";

type PetStatus = "idle" | "thinking" | "error";

function petStatus(status: ChatStatus): PetStatus {
  if (status === "sending" || status === "streaming") return "thinking";
  if (status === "error") return "error";
  return "idle";
}

const STATUS_TEXT: Record<PetStatus, string> = {
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
  const state = petStatus(status);

  return (
    <div id="pet">
      <span className="pet__hint" aria-hidden="true">
        Klik untuk buka chat
      </span>
      <span className={`pet__status pet__status--${state}`} role="status">
        <span className="pet__mark" aria-hidden="true">
          {state === "error" ? <IconAlert size={10} /> : null}
        </span>
        <span className="sr-only">{STATUS_TEXT[state]}</span>
      </span>
      <button
        type="button"
        className={state === "thinking" ? "pet__btn pet__btn--busy" : "pet__btn"}
        {...drag}
        aria-label="Buka chat"
        title="Klik: buka chat. Geser: pindah"
      >
        <span className="pet__float">
          <svg
            width="116"
            height="116"
            viewBox="0 0 116 116"
            aria-hidden="true"
          >
            {/* Leaves: three brand-leaf shoots rising from the crown. */}
            <path
              d="M58 34 C48 28 42 18 42 8 C52 10 58 21 58 34 Z"
              fill="var(--color-brand-leaf)"
            />
            <path
              d="M58 32 C57 22 57 12 60 4 C64 12 63 24 58 32 Z"
              fill="var(--color-brand-leaf)"
            />
            <path
              d="M58 34 C66 30 72 22 73 13 C64 14 58 24 58 34 Z"
              fill="var(--color-brand-leaf)"
            />
            {/* Body: tapered root, wide crown narrowing to a tip. */}
            <path
              d="M38 28 L78 28 C78 50 68 74 59 99 C58.4 101 57.6 101 57 99 C48 74 38 50 38 28 Z"
              fill="var(--color-brand)"
            />
            {/* Shading sliver down the left flank for a hint of roundness. */}
            <path
              d="M44 28 L51 28 C49 50 48 70 50 88 C44 72 42 48 44 28 Z"
              fill="var(--color-ink)"
              opacity="0.12"
            />
            {/* Ridge detail strokes in brand at lower opacity; they read as
               slightly lighter lines where they cross the shaded flank. */}
            <path
              d="M47 68 Q58 72 69 68"
              fill="none"
              stroke="var(--color-brand)"
              strokeWidth="2"
              strokeLinecap="round"
              opacity="0.55"
            />
            <path
              d="M50 80 Q58 83 66 80"
              fill="none"
              stroke="var(--color-brand)"
              strokeWidth="2"
              strokeLinecap="round"
              opacity="0.55"
            />
            {/* Eyes in --color-ink: near-white on dark theme, near-black on
               light theme — both sit legibly on the brand-orange body. */}
            <circle cx="50" cy="46" r="3.2" fill="var(--color-ink)" />
            <circle cx="66" cy="46" r="3.2" fill="var(--color-ink)" />
            <path
              d="M52 56 Q58 61 64 56"
              fill="none"
              stroke="var(--color-ink)"
              strokeWidth="2"
              strokeLinecap="round"
            />
          </svg>
        </span>
      </button>
      <div className="pet__actions">
        <button
          type="button"
          className="icon-btn"
          onClick={onHide}
          aria-label="Sembunyikan"
          title="Sembunyikan"
        >
          <IconEyeOff />
        </button>
        <button
          type="button"
          className="icon-btn icon-btn--danger"
          onClick={onQuit}
          aria-label="Keluar"
          title="Keluar"
        >
          <IconPower />
        </button>
      </div>
    </div>
  );
}
