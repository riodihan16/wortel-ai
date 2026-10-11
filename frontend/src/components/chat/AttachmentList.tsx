// AttachmentList shows a sent message's attachments: image thumbnails that
// toggle to full width on click, and name+size chips for files. Images come
// from the local asset route, never from provider output.
import { useState } from "react";
import type { JSX } from "react";
import type { AttachmentMeta } from "../../types/chat";
import { IconFile } from "../ui/Icon";
import { formatSize } from "../../lib/formatSize";

export function AttachmentList({ items }: { items: AttachmentMeta[] }): JSX.Element {
  const [open, setOpen] = useState<string | null>(null);
  return (
    <div className="att-list" role="group" aria-label={`Lampiran: ${items.length} berkas`}>
      {items.map((a) =>
        a.kind === "image" ? (
          <button
            key={a.id}
            type="button"
            className={open === a.id ? "att-thumb att-thumb--open" : "att-thumb"}
            onClick={() => setOpen(open === a.id ? null : a.id)}
            aria-pressed={open === a.id}
            aria-label={a.name}
            title={open === a.id ? "Perkecil" : "Perbesar"}
          >
            <img src={`/attachments/${a.id}`} alt={a.name} />
          </button>
        ) : (
          <span key={a.id} className="att-chip">
            <IconFile size={14} />
            <span className="att-chip__name">{a.name}</span>
            <span className="att-chip__size">{formatSize(a.size)}</span>
          </span>
        )
      )}
    </div>
  );
}
