import { useEffect, useRef, useState } from "react";
import type { ChatStatus } from "../types/chat";

export type ViewMode = "pet" | "chat" | "settings";

// usePetSignal turns chat status into what the desktop pet should say.
// - replyReady: a reply finished while the user was on the pet; it stays
//   until chat is opened (the pet's "Sudah!" bubble).
// - petStatus: the chat status, except a failure the user has already seen
//   in chat no longer keeps the pet worried; the next failure does again.
export function usePetSignal(mode: ViewMode, status: ChatStatus): { petStatus: ChatStatus; replyReady: boolean } {
  const [replyReady, setReplyReady] = useState(false);
  const [errorSeen, setErrorSeen] = useState(false);
  const prevStatus = useRef(status);

  useEffect(() => {
    const was = prevStatus.current;
    prevStatus.current = status;
    if (mode === "pet" && status === "complete" && was !== "complete") setReplyReady(true);
  }, [mode, status]);

  useEffect(() => {
    if (mode !== "pet") setReplyReady(false);
  }, [mode]);

  useEffect(() => {
    if (status !== "error") setErrorSeen(false);
    else if (mode !== "pet") setErrorSeen(true);
  }, [mode, status]);

  return {
    petStatus: status === "error" && errorSeen ? "idle" : status,
    replyReady,
  };
}
