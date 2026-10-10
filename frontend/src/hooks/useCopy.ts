import { useCallback, useEffect, useRef, useState } from "react";

export type CopyState = "idle" | "copied" | "failed";

// useCopy writes text to the clipboard and reports the outcome for a short
// while, so a copy button can say "Tersalin" or "Gagal menyalin" instead of
// failing silently (WebView2 can refuse clipboard access).
export function useCopy(resetMs = 1500): { state: CopyState; copy: (text: string) => Promise<void> } {
  const [state, setState] = useState<CopyState>("idle");
  const timer = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (timer.current !== null) window.clearTimeout(timer.current);
    };
  }, []);

  const copy = useCallback(
    async (text: string) => {
      let next: CopyState = "failed";
      try {
        const clip = navigator.clipboard;
        if (clip) {
          await clip.writeText(text);
          next = "copied";
        }
      } catch {
        next = "failed";
      }
      setState(next);
      if (timer.current !== null) window.clearTimeout(timer.current);
      timer.current = window.setTimeout(() => setState("idle"), resetMs);
    },
    [resetMs]
  );

  return { state, copy };
}
