import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";

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

// Allow task-list checkboxes (GFM) plus the attributes markdown normally
// emits, nothing more. `defaultSchema` already permits the common inline /
// block tags and drops scripts, styles, iframes, and event handlers.
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
        }}
      >
        {text}
      </ReactMarkdown>
    </div>
  );
}
