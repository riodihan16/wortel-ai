/* Hallmark · component: icon set · genre: modern-minimal · theme: Tanah
 * states: n/a (non-interactive glyphs) — colour comes from currentColor so
 * every icon inherits its button's ink token, including hover/disabled.
 * One stroke language: 24px box, 1.75 stroke, round caps and joins.
 */
import type { SVGProps } from "react";

type IconProps = SVGProps<SVGSVGElement> & { size?: number };

function Svg({ size = 16, children, ...rest }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.75}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
      {...rest}
    >
      {children}
    </svg>
  );
}

export const IconSliders = (p: IconProps) => (
  <Svg {...p}>
    <path d="M5 20v-6M5 10V4M12 20v-9M12 7V4M19 20v-3M19 13V4" />
    <circle cx="5" cy="12" r="1.9" />
    <circle cx="12" cy="9" r="1.9" />
    <circle cx="19" cy="15" r="1.9" />
  </Svg>
);

export const IconEyeOff = (p: IconProps) => (
  <Svg {...p}>
    <path d="M9.9 5.2A9.6 9.6 0 0 1 12 5c5 0 8.5 4.2 9.4 6.2.2.5.2 1.1 0 1.6a13 13 0 0 1-2.3 3.3" />
    <path d="M6.6 6.7A13.4 13.4 0 0 0 2.6 12.8c.9 2 4.4 6.2 9.4 6.2 1.7 0 3.2-.5 4.5-1.2" />
    <path d="M10 10a2.8 2.8 0 0 0 4 4" />
    <path d="M3.5 3.5l17 17" />
  </Svg>
);

export const IconPower = (p: IconProps) => (
  <Svg {...p}>
    <path d="M12 3.5v7.5" />
    <path d="M6.9 6.9a7.2 7.2 0 1 0 10.2 0" />
  </Svg>
);

export const IconPlus = (p: IconProps) => (
  <Svg {...p}>
    <path d="M12 5.5v13M5.5 12h13" />
  </Svg>
);

export const IconPencil = (p: IconProps) => (
  <Svg {...p}>
    <path d="M4 20h4l10.5-10.5a2.1 2.1 0 0 0-3-3L5 17v3Z" />
    <path d="M14.5 6.5l3 3" />
  </Svg>
);

export const IconTrash = (p: IconProps) => (
  <Svg {...p}>
    <path d="M4.5 6.5h15" />
    <path d="M9.5 6.5V5a1.5 1.5 0 0 1 1.5-1.5h2A1.5 1.5 0 0 1 14.5 5v1.5" />
    <path d="M6.5 6.5 7.4 19a1.6 1.6 0 0 0 1.6 1.5h6a1.6 1.6 0 0 0 1.6-1.5l.9-12.5" />
    <path d="M10.5 10.5v6M13.5 10.5v6" />
  </Svg>
);

export const IconChevronDown = (p: IconProps) => (
  <Svg {...p}>
    <path d="M6.5 9.5 12 15l5.5-5.5" />
  </Svg>
);

export const IconCheck = (p: IconProps) => (
  <Svg {...p}>
    <path d="M5 12.5 9.5 17 19 7" />
  </Svg>
);

export const IconX = (p: IconProps) => (
  <Svg {...p}>
    <path d="M6.5 6.5l11 11M17.5 6.5l-11 11" />
  </Svg>
);

export const IconArrowUp = (p: IconProps) => (
  <Svg {...p}>
    <path d="M12 19V5.5" />
    <path d="M6.5 11 12 5.5 17.5 11" />
  </Svg>
);

export const IconStop = (p: IconProps) => (
  <Svg {...p}>
    <rect x="7" y="7" width="10" height="10" rx="2.5" fill="currentColor" stroke="none" />
  </Svg>
);

export const IconAlert = (p: IconProps) => (
  <Svg {...p}>
    <path d="M12 4.5 2.9 20h18.2L12 4.5Z" />
    <path d="M12 10v4.2M12 17.2v.1" />
  </Svg>
);

export const IconCopy = (p: IconProps) => (
  <Svg {...p}>
    <rect x="9" y="9" width="11" height="11" rx="2.5" />
    <path d="M15 9V6.5A2.5 2.5 0 0 0 12.5 4h-6A2.5 2.5 0 0 0 4 6.5v6A2.5 2.5 0 0 0 6.5 15H9" />
  </Svg>
);

export const IconKey = (p: IconProps) => (
  <Svg {...p}>
    <circle cx="8" cy="15" r="3.5" />
    <path d="M10.6 12.6 19 4.2" />
    <path d="M16 7.2l2 2M13.6 9.6l2 2" />
  </Svg>
);

export const IconRefresh = (p: IconProps) => (
  <Svg {...p}>
    <path d="M20 12a8 8 0 1 1-2.6-5.9" />
    <path d="M20 4v4.5h-4.5" />
  </Svg>
);

export const IconExternal = (p: IconProps) => (
  <Svg {...p}>
    <path d="M13.5 5.5H19V11" />
    <path d="M19 5.5 11 13.5" />
    <path d="M18 14.5v3A2.5 2.5 0 0 1 15.5 20h-9A2.5 2.5 0 0 1 4 17.5v-9A2.5 2.5 0 0 1 6.5 6h3" />
  </Svg>
);

/* Two arrows pointing inward: shrink the panel back to the pet. */
export const IconMinimize = (p: IconProps) => (
  <Svg {...p}>
    <path d="M4.5 14H10v5.5" />
    <path d="M19.5 10H14V4.5" />
    <path d="M10 14 4 20" />
    <path d="M14 10l6-6" />
  </Svg>
);

/* A small cube: "the model", without the "language/web" read of a globe. */
export const IconModel = (p: IconProps) => (
  <Svg {...p}>
    <path d="M12 3.5 19.5 7.75v8.5L12 20.5 4.5 16.25v-8.5Z" />
    <path d="M4.5 7.75 12 12l7.5-4.25" />
    <path d="M12 12v8.5" />
  </Svg>
);

export const IconPaperclip = (p: IconProps) => (
  <Svg {...p}>
    <path d="m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
  </Svg>
);

export const IconFile = (p: IconProps) => (
  <Svg {...p}>
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
    <path d="M14 2v6h6" />
    <path d="M16 13H8" />
    <path d="M16 17H8" />
  </Svg>
);
