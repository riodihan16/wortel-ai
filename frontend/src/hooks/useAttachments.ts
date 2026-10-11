import { useCallback, useRef, useState } from "react";
import {
  PickAttachments,
  StageAttachmentData,
  StageAttachmentPaths,
  UnstageAttachment,
} from "../../wailsjs/go/main/App";
import type { AttachmentMeta } from "../types/chat";

export const MAX_ATTACHMENTS = 4;
// Largest single file the backend accepts (attach.MaxFileBytes); a bigger
// paste is refused before it is read and base64-encoded.
const MAX_FILE_MB = 15;
const MAX_FILE_BYTES = MAX_FILE_MB * 1024 * 1024;

export interface AttachmentsApi {
  pending: AttachmentMeta[];
  error: string | null;
  pick: () => Promise<void>;
  stagePaths: (paths: string[]) => Promise<void>;
  stageFiles: (files: File[]) => Promise<void>;
  remove: (id: string) => Promise<void>;
  /** Forget the list locally; the backend moved the items to history on send. */
  reset: () => void;
  clearError: () => void;
}

// Wails rejects with the Go error string; tests reject with Error.
function messageOf(e: unknown, fallback: string): string {
  if (typeof e === "string" && e) return e;
  if (e instanceof Error && e.message) return e.message;
  return fallback;
}

function toBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const r = new FileReader();
    r.onerror = () => reject(new Error("Berkas gagal dibaca."));
    r.onload = () => {
      const s = String(r.result);
      resolve(s.slice(s.indexOf(",") + 1));
    };
    r.readAsDataURL(file);
  });
}

// useAttachments owns the composer's pending attachments. Bytes live in the
// Go staging area; this hook holds metadata only. Dialog, drop, and paste all
// end in the same list. A failed binding leaves the list unchanged and
// reports why; the per-message cap is checked here first so a fifth file
// never round-trips to the backend.
export function useAttachments(): AttachmentsApi {
  const [pending, setPending] = useState<AttachmentMeta[]>([]);
  const [error, setError] = useState<string | null>(null);
  const pendingRef = useRef(pending);
  pendingRef.current = pending;

  const stage = useCallback(
    async (count: number, run: () => Promise<AttachmentMeta[] | null>, fallback: string) => {
      if (count === 0) return;
      if (pendingRef.current.length + count > MAX_ATTACHMENTS) {
        setError(`Maksimal ${MAX_ATTACHMENTS} lampiran per pesan.`);
        return;
      }
      setError(null);
      try {
        const added = (await run()) ?? [];
        if (added.length > 0) setPending((prev) => [...prev, ...added]);
      } catch (e) {
        setError(messageOf(e, fallback));
      }
    },
    []
  );

  // The generated bindings type `kind` as string; the backend only ever
  // sends image|pdf|text, so narrow at the boundary.
  const asMetas = (p: Promise<unknown>) => p as Promise<AttachmentMeta[] | null>;

  // The dialog's count is unknown until it closes; 1 checks there is room at
  // all, the backend enforces the cap on the real batch.
  const pick = useCallback(() => stage(1, () => asMetas(PickAttachments()), "Dialog berkas gagal dibuka."), [stage]);

  const stagePaths = useCallback(
    (paths: string[]) => stage(paths.length, () => asMetas(StageAttachmentPaths(paths)), "Berkas gagal dilampirkan."),
    [stage]
  );

  const stageFiles = useCallback(
    async (files: File[]) => {
      const big = files.find((f) => f.size > MAX_FILE_BYTES);
      if (big) {
        setError(`${big.name}: melebihi ${MAX_FILE_MB} MB`);
        return;
      }
      await stage(
        files.length,
        async () => {
          const out: AttachmentMeta[] = [];
          try {
            for (const f of files) {
              out.push((await StageAttachmentData(f.name, f.type, await toBase64(f))) as AttachmentMeta);
            }
          } catch (e) {
            // Keep backend staging in step with the list the user sees.
            for (const m of out) void UnstageAttachment(m.id).catch(() => {});
            throw e;
          }
          return out;
        },
        "Gambar gagal ditempel."
      );
    },
    [stage]
  );

  const remove = useCallback(async (id: string) => {
    setPending((prev) => prev.filter((a) => a.id !== id));
    try {
      await UnstageAttachment(id);
    } catch {
      // The list is what the user sees; the backend drops leftovers at exit.
    }
  }, []);

  const reset = useCallback(() => setPending([]), []);
  const clearError = useCallback(() => setError(null), []);

  return { pending, error, pick, stagePaths, stageFiles, remove, reset, clearError };
}
