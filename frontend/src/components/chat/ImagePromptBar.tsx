// ImagePromptBar owns the image-generation row above the chat composer:
// prompt entry, image-model display, generate/cancel, a distinct "drawing"
// state, and inline errors. It never renders a provider URL; the finished
// image appears in the transcript via MessageItem. Generation is billed from
// PAYG balance, not the chat plan (Kenari docs §12.4), so the bar says so.
import { useEffect, useRef, useState } from "react";
import type { JSX, KeyboardEvent } from "react";
import { IconAlert, IconModel, IconStop, IconX } from "../ui/Icon";

interface Props {
  /** Image generation is in flight. */
  pending: boolean;
  /** Why the last generation failed, if it did. */
  error: string | null;
  /** Configured image model id, empty when none is set. */
  imageModel: string | null;
  /** Chat streaming is active; independent of image generation. */
  chatBusy: boolean;
  onGenerate: (prompt: string) => void;
  onCancel: () => void;
  onDismissError: () => void;
  onOpenSettings: () => void;
}

export function ImagePromptBar({
  pending,
  error,
  imageModel,
  chatBusy,
  onGenerate,
  onCancel,
  onDismissError,
  onOpenSettings,
}: Props): JSX.Element {
  const [prompt, setPrompt] = useState("");
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // A finished generation should not leave a stale error on screen.
  useEffect(() => {
    if (pending) return;
  }, [pending]);

  const hasModel = !!imageModel;
  const canGenerate = prompt.trim() !== "" && hasModel && !pending;

  const submit = () => {
    if (!canGenerate) return;
    onGenerate(prompt);
    setPrompt("");
  };

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Escape" && pending) {
      e.preventDefault();
      onCancel();
      return;
    }
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  };

  return (
    <div id="image-prompt">
      <div id="image-prompt__row">
        <textarea
          ref={inputRef}
          className="textarea"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={onKeyDown}
          placeholder={
            !hasModel
              ? "Pilih model gambar di Pengaturan"
              : "Deskripsikan gambar… (Enter buat, Shift+Enter baris baru)"
          }
          rows={1}
          aria-label="Prompt gambar"
          disabled={!hasModel || pending}
        />
        <button
          type="button"
          id="btn-generate-image"
          className="btn btn--icon btn--primary"
          onClick={submit}
          disabled={!canGenerate}
          aria-label="Buat gambar"
          title="Buat gambar (Enter)"
        >
          <IconModel size={15} />
        </button>
      </div>
      {pending && (
        <div id="image-prompt__drawing" role="status" aria-live="polite">
          <span className="image-prompt__spinner" aria-hidden="true" />
          <span className="image-prompt__label">Wortel sedang menggambar…</span>
          <button
            type="button"
            className="btn btn--sm btn--ghost"
            onClick={onCancel}
            aria-label="Hentikan pembuatan gambar"
          >
            <IconStop size={13} />
            Hentikan
          </button>
        </div>
      )}
      {hasModel ? (
        <p className="image-prompt__hint">Dibuat dari saldo PAYG, bukan paket chat.</p>
      ) : (
        <p className="image-prompt__hint">
          Belum ada model gambar.{" "}
          <button type="button" className="btn btn--sm btn--ghost" onClick={onOpenSettings}>
            Atur di Pengaturan
          </button>
        </p>
      )}
      {error && (
        <div className="banner banner--error" role="alert">
          <IconAlert size={14} />
          <span className="banner__text">{error}</span>
          <button type="button" className="icon-btn" onClick={onDismissError} aria-label="Tutup pesan gambar" title="Tutup">
            <IconX size={13} />
          </button>
        </div>
      )}
      {/* Kick off the drawing state without letting a chat stream interfere. */}
      <input type="hidden" value={chatBusy ? "1" : "0"} readOnly />
    </div>
  );
}
