/* Hallmark · component: segmented control · genre: modern-minimal · theme: Tanah
 * states: default · hover · focus-visible · active · disabled
 * Equal-width cells; the thumb slides via --seg-index/--seg-count so no
 * measurement is needed and the animation stays on the compositor.
 */
import type { CSSProperties } from "react";

interface Option<T> {
  value: T;
  label: string;
}

interface Props<T> {
  value: T;
  options: readonly Option<T>[];
  onChange: (value: T) => void;
  disabled?: boolean;
  label: string;
}

export function Segmented<T extends string | number>({
  value,
  options,
  onChange,
  disabled,
  label,
}: Props<T>) {
  const index = Math.max(
    0,
    options.findIndex((o) => o.value === value)
  );

  return (
    <div
      className="seg"
      role="radiogroup"
      aria-label={label}
      style={
        {
          "--seg-index": index,
          "--seg-count": options.length,
        } as CSSProperties
      }
    >
      <span className="seg__thumb" aria-hidden="true" />
      {options.map((o) => (
        <button
          key={String(o.value)}
          type="button"
          role="radio"
          aria-checked={o.value === value}
          disabled={disabled}
          className={
            o.value === value ? "seg__item seg__item--on" : "seg__item"
          }
          onClick={() => onChange(o.value)}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
