// Carrot is the companion's one drawing, shared by the pet window, chat empty
// states and the assistant's mark in the transcript. Its colours are fixed
// --carrot-* tokens, never theme tokens: the pet sits on the desktop wallpaper,
// not on the app's paper, so it must not change with Gelap/Terang. The dark
// outline keeps the shape legible on any wallpaper.
//
// Mood is carried by the face and leaves, not by colour alone:
// idle = smile; thinking = eyes up, small mouth; error = worried brows,
// frown, drooping leaves.
import type { JSX } from "react";

export type CarrotMood = "idle" | "thinking" | "error";

interface Props {
  size?: number;
  mood?: CarrotMood;
  className?: string;
}

const EYES: Record<CarrotMood, { y: number; dx: number; r: number }> = {
  idle: { y: 46, dx: 0, r: 3.2 },
  thinking: { y: 43.5, dx: 1.2, r: 3 },
  error: { y: 47, dx: 0, r: 2.8 },
};

const MOUTH: Record<CarrotMood, string> = {
  idle: "M52 56 Q58 61.5 64 56",
  thinking: "M55.5 58 Q58 56.6 60.5 58",
  error: "M52.5 59.5 Q58 55 63.5 59.5",
};

export function Carrot({ size = 116, mood = "idle", className }: Props): JSX.Element {
  const eye = EYES[mood];
  return (
    <svg
      className={["carrot", `carrot--${mood}`, className].filter(Boolean).join(" ")}
      width={size}
      height={size}
      viewBox="0 0 116 116"
      aria-hidden="true"
      focusable="false"
    >
      <g className="carrot__leaves" stroke="var(--carrot-ink)" strokeWidth="2.2" strokeLinejoin="round">
        <path d="M58 34 C48 28 42 18 42 8 C52 10 58 21 58 34 Z" fill="var(--carrot-leaf)" />
        <path d="M58 32 C57 22 57 12 60 4 C64 12 63 24 58 32 Z" fill="var(--carrot-leaf)" />
        <path d="M58 34 C66 30 72 22 73 13 C64 14 58 24 58 34 Z" fill="var(--carrot-leaf)" />
      </g>
      <path
        className="carrot__body"
        d="M38 28 L78 28 C78 50 68 74 59 99 C58.4 101 57.6 101 57 99 C48 74 38 50 38 28 Z"
        fill="var(--carrot-body)"
        stroke="var(--carrot-ink)"
        strokeWidth="2.2"
        strokeLinejoin="round"
      />
      {/* Light from the upper left: a soft shine down the left flank. */}
      <path d="M44 31 L50 31 C48.5 50 47.5 68 49.5 84 C44.5 70 42.5 48 44 31 Z" fill="var(--carrot-shine)" opacity="0.26" />
      <g fill="none" stroke="var(--carrot-shade)" strokeWidth="2" strokeLinecap="round">
        <path d="M63 68 Q67.5 68 70 65.5" />
        <path d="M47 76 Q51 78 55 77.5" />
        <path d="M60 85 Q63 85 65 83" />
      </g>
      <g className="carrot__cheeks" fill="var(--carrot-cheek)" opacity="0.45">
        <ellipse cx="45.5" cy="53" rx="3.4" ry="2" />
        <ellipse cx="70.5" cy="53" rx="3.4" ry="2" />
      </g>
      {mood === "error" && (
        <g stroke="var(--carrot-ink)" strokeWidth="2" strokeLinecap="round">
          <path d="M45.5 41.5 L52.5 39.5" />
          <path d="M63.5 39.5 L70.5 41.5" />
        </g>
      )}
      <g className="carrot__eyes" fill="var(--carrot-ink)">
        <circle cx={50 + eye.dx} cy={eye.y} r={eye.r} />
        <circle cx={66 + eye.dx} cy={eye.y} r={eye.r} />
      </g>
      <path d={MOUTH[mood]} fill="none" stroke="var(--carrot-ink)" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}
