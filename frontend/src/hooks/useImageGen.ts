import { useCallback, useEffect, useRef, useState } from "react";
import { EventsOff, EventsOn } from "../../wailsjs/runtime/runtime";
import { CancelImageGeneration, GenerateImage } from "../../wailsjs/go/main/App";
import type { ImageEvent, ImageStatus } from "../types/image";

export interface ImageGenApi {
  pending: boolean;
  status: ImageStatus;
  error: string | null;
  /** Generate on the given conversation with the given prompt + image model. */
  generate: (conversationId: string, prompt: string, model: string) => Promise<void>;
  cancel: () => void;
  clearError: () => void;
}

// useImageGen owns one image generation for the active conversation. Events
// are matched by conversationId and requestId so a stale or superseded event
// never mutates local state. Independent of useChat: a running chat stream
// does not block image generation and vice versa (one image per conversation
// is enforced backend-side).
export function useImageGen(): ImageGenApi {
  const [pending, setPending] = useState(false);
  const [status, setStatus] = useState<ImageStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const requestRef = useRef<string | null>(null);
  const convRef = useRef<string>("");
  const sendingRef = useRef(false);

  const finish = useCallback(() => {
    requestRef.current = null;
    sendingRef.current = false;
    setPending(false);
  }, []);

  useEffect(() => {
    const handler = (ev: ImageEvent) => {
      if (ev.conversationId !== convRef.current) return;
      // A started event may arrive before the binding returns the requestId;
      // adopt it. Otherwise only the active request's events apply.
      if (ev.type === "started" && requestRef.current === null) {
        requestRef.current = ev.requestId;
      } else if (ev.requestId !== requestRef.current) {
        return;
      }
      switch (ev.type) {
        case "started":
          setPending(true);
          setStatus("pending");
          break;
        case "completed":
          setPending(false);
          setStatus("complete");
          finish();
          break;
        case "cancelled":
          setPending(false);
          setStatus("cancelled");
          finish();
          break;
        case "failed":
          setPending(false);
          setStatus("error");
          setError(ev.error?.message ?? "Pembuatan gambar gagal.");
          finish();
          break;
      }
    };
    EventsOn("image:event", handler);
    return () => {
      EventsOff("image:event");
    };
  }, [finish]);

  const generate = useCallback(
    async (conversationId: string, prompt: string, model: string) => {
      const trimmed = prompt.trim();
      if (!trimmed || sendingRef.current || requestRef.current || !model) return;
      convRef.current = conversationId;
      sendingRef.current = true;
      setError(null);
      setPending(true);
      setStatus("pending");
      try {
        const requestId = await GenerateImage(conversationId, trimmed, model);
        // Adopt only if a fast terminal event has not already cleared it.
        if (sendingRef.current) requestRef.current = requestId;
      } catch (e) {
        setPending(false);
        setStatus("error");
        setError(e instanceof Error ? e.message : "Gagal membuat gambar.");
        sendingRef.current = false;
        requestRef.current = null;
      }
    },
    []
  );

  const cancel = useCallback(() => {
    // Release guards immediately; the backend's own cancelled event is then
    // dropped by the requestId filter, so settle the state here.
    const conv = convRef.current;
    void CancelImageGeneration(conv).catch(() => {});
    setPending(false);
    setStatus("idle");
    finish();
  }, [finish]);

  const clearError = useCallback(() => setError(null), []);

  return { pending, status, error, generate, cancel, clearError };
}
