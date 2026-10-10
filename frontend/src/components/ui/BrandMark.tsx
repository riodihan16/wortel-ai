// BrandMark is the 18px faceless carrot in window headers. Same fixed
// --carrot-* palette as the character, so it matches the pet in both themes.
import type { JSX } from "react";

export function BrandMark(): JSX.Element {
  return (
    <svg className="brand__mark" width="18" height="18" viewBox="0 0 18 18" aria-hidden="true" focusable="false">
      <path d="M9 5.2 C6.6 4.2 5.4 2.4 5.2 0.6 C7 1 8.6 2.8 9 5.2 Z" fill="var(--carrot-leaf)" />
      <path d="M9 5.2 C9 3.4 9.2 1.7 10 0.2 C10.9 1.8 10.4 3.8 9 5.2 Z" fill="var(--carrot-leaf)" />
      <path d="M9 5.2 C11 4.4 12.4 2.9 12.7 1.1 C11 1.4 9.5 3 9 5.2 Z" fill="var(--carrot-leaf)" />
      <path
        d="M4.8 4.6 L13.2 4.6 C13.2 8.6 11.4 12.6 9.3 17.1 C9.1 17.6 8.9 17.6 8.7 17.1 C6.6 12.6 4.8 8.6 4.8 4.6 Z"
        fill="var(--carrot-body)"
      />
    </svg>
  );
}
