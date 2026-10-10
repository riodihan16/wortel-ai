// MessageItem renders one turn of the transcript. Replies are spoken by the
// carrot: an avatar whose face follows the reply (thinking while streaming,
// worried on failure) beside unboxed text; the user's turns are neutral
// bubbles. A reply's outcome lives on the reply itself, never in a separate
// status row: a failed reply becomes an inline error card (reason + retry on
// the newest turn), a cancelled reply gets an end-cap, and an empty streaming
// reply shows typing dots until the first delta lands.
import type { JSX } from "react";
import type { ChatMessage } from "../../types/chat";
import { Carrot, type CarrotMood } from "../ui/Carrot";
import { IconAlert } from "../ui/Icon";
import { SafeMarkdown } from "./SafeMarkdown";

interface Props {
  message: ChatMessage;
  /** Newest message: the only one whose failure can be retried. */
  isLast: boolean;
  busy: boolean;
  onRetry: () => void;
}

export function MessageItem({ message: m, isLast, busy, onRetry }: Props): JSX.Element {
  const failed = m.role === "assistant" && m.status === "error";
  const cancelled = m.role === "assistant" && m.status === "cancelled";
  const waiting = m.role === "assistant" && m.status === "streaming" && !m.content;
  const showBody = !(failed && !m.content);
  const mood: CarrotMood = failed ? "error" : m.status === "streaming" ? "thinking" : "idle";

  return (
    <div className={`msg msg-${m.role}`}>
      {m.role === "assistant" && <Carrot size={24} mood={mood} className="msg-avatar" />}
      <div className="msg-main">
        {showBody && (
          <div className="msg-body">
            {waiting ? (
              <span className="msg-typing" role="img" aria-label="Wortel sedang menulis">
                <span />
                <span />
                <span />
              </span>
            ) : m.role === "assistant" ? (
              <SafeMarkdown text={m.content} />
            ) : (
              m.content
            )}
            {cancelled && (
              <span className="msg-endcap">{m.content ? "Dihentikan" : "Dihentikan sebelum ada balasan"}</span>
            )}
          </div>
        )}
        {failed && (
          <div className="msg-failed" role={isLast ? "alert" : undefined}>
            <IconAlert size={14} />
            <span className="msg-failed__text">{m.errorText || "Balasan ini gagal."}</span>
            {isLast && !busy && (
              <button type="button" className="btn btn--sm" onClick={onRetry}>
                Coba lagi
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
