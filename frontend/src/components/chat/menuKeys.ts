import type { KeyboardEvent } from "react";

// Shared keyboard model for the role="menu" pickers (conversation, model):
// ArrowUp/Down wrap, Home/End jump. Escape and Tab-out are owned by each
// picker because they also close it.
export function moveMenuFocus(e: KeyboardEvent<HTMLElement>, menu: HTMLElement | null): void {
  const opts = Array.from(menu?.querySelectorAll<HTMLElement>('[role="menuitemradio"]') ?? []);
  if (opts.length === 0) return;
  const at = opts.indexOf(document.activeElement as HTMLElement);
  const next: Record<string, number> = {
    ArrowDown: (at + 1) % opts.length,
    ArrowUp: at <= 0 ? opts.length - 1 : at - 1,
    Home: 0,
    End: opts.length - 1,
  };
  if (!(e.key in next)) return;
  e.preventDefault();
  opts[next[e.key]].focus();
}
