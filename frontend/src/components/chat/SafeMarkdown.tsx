// Markdown renderer for AI output (untrusted input).
//
// Layered defense:
// 1. react-markdown never renders raw HTML (no rehype-raw), so an injected
//    `<img onerror>` or `<script>` stays literal text.
// 2. rehype-sanitize runs the hast tree through a strict allowlist as a
//    second layer, in case a plugin ever adds raw nodes.
// 3. Links are restricted to http/https by the schema's protocols plus a
//    component override that drops everything else. External links open in
//    the default browser via Wails; rel guards the opener.
//
// GFM (tables, strikethrough, task lists, autolinks) is on: AI answers use
// tables often and the previous hand-rolled parser rendered them as noise.
//
// Extra overrides: `table` is wrapped so wide tables scroll inside the bubble;
// `pre` gets a Copy button that falls back silently when clipboard is missing.

// Allow task-list checkboxes (GFM) plus the attributes markdown normally
// emits, nothing more. `defaultSchema` already permits the common inline /
// block tags and drops scripts, styles, iframes, and event handlers.
import { isValidElement, useEffect, useRef, useState } from "react";
import type { JSX, ReactNode } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import { IconCheck, IconCopy } from "../ui/Icon";

const schema = {
  ...defaultSchema,
  tagNames: [...(defaultSchema.tagNames ?? []), "input"],
  attributes: {
    ...defaultSchema.attributes,
    input: [["type", "checkbox"], "checked", "disabled"],
  },
  protocols: {
    ...defaultSchema.protocols,
    href: ["http", "https"],
  },
};

function CodeBlock({ children }: { children: ReactNode }): JSX.Element {
  const [copied, setCopied] = useState(false);
  const timer = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (timer.current !== null) window.clearTimeout(timer.current);
    };
  }, []);

  const copy = async () => {
    try {
      const clip = navigator.clipboard;
      if (!clip) return;
      await clip.writeText(extractText(children));
      setCopied(true);
      if (timer.current !== null) window.clearTimeout(timer.current);
      timer.current = window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard unavailable (non-secure context): stay silent, no throw.
    }
  };

  return (
    <div className="md-code">
      <pre>{children}</pre>
      <button
        type="button"
        className="btn btn--sm btn--ghost md-code__copy"
        onClick={copy}
        aria-label={copied ? "Tersalin" : "Salin kode"}
        title={copied ? "Tersalin" : "Salin kode"}
      >
        {copied ? <IconCheck size={13} /> : <IconCopy size={13} />}
        {copied ? "Tersalin" : "Salin"}
      </button>
    </div>
  );
}

function extractText(node: ReactNode): string {
  if (node === null || node === undefined) return "";
  if (typeof node === "string" || typeof node === "number") return String(node);
  if (Array.isArray(node)) return node.map(extractText).join("");
  if (isValidElement<{ children?: ReactNode }>(node)) {
    return extractText(node.props.children);
  }
  return "";
}

export function SafeMarkdown({ text }: { text: string }) {
  return (
    <div className="md">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[[rehypeSanitize, schema]]}
        components={{
          a: ({ href, children, ...rest }) => {
            const safe = typeof href === "string" && /^https?:\/\//i.test(href);
            if (!safe) {
              // Non-http link (javascript:, data:, file:, relative): text only.
              return <span>{children}</span>;
            }
            return (
              <a href={href} target="_blank" rel="noreferrer noopener" {...rest}>
                {children}
              </a>
            );
          },
          table: ({ children, ...rest }) => (
            <div className="md-table-wrap">
              <table {...rest}>{children}</table>
            </div>
          ),
          pre: ({ children }) => <CodeBlock>{children}</CodeBlock>,
        }}
      >
        {text}
      </ReactMarkdown>
    </div>
  );
}
