import type { ReactNode } from "react";

// Minimal safe markdown: paragraphs, lists, code blocks, inline code,
// bold, italic, links. Raw HTML never rendered; text always escaped by React.
// Link allowlist: http/https only.

const ALLOWED_SCHEMES = /^(https?:\/\/|#)/i;

interface Block {
  type: "para" | "list" | "code";
  lines: string[];
  lang?: string;
}

function splitBlocks(src: string): Block[] {
  const out: Block[] = [];
  const lines = src.replace(/\r\n/g, "\n").split("\n");
  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (line.trim().startsWith("```")) {
      const lang = line.trim().slice(3).trim();
      const body: string[] = [];
      i++;
      while (i < lines.length && !lines[i].trim().startsWith("```")) {
        body.push(lines[i]);
        i++;
      }
      i++; // skip closing fence
      out.push({ type: "code", lines: body, lang });
      continue;
    }
    if (/^\s*([-∗*]|\d+[.)])\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^\s*([-∗*]|\d+[.)])\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^\s*([-∗*]|\d+[.)])\s+/, ""));
        i++;
      }
      out.push({ type: "list", lines: items });
      continue;
    }
    if (line.trim() === "") {
      i++;
      continue;
    }
    const para: string[] = [];
    while (
      i < lines.length &&
      lines[i].trim() !== "" &&
      !lines[i].trim().startsWith("```") &&
      !/^\s*([-∗*]|\d+[.)])\s+/.test(lines[i])
    ) {
      para.push(lines[i]);
      i++;
    }
    out.push({ type: "para", lines: para });
  }
  return out;
}

// Inline: `code`, **bold**, *italic*, [text](url). Escaping via React nodes.
function renderInline(text: string, keyPrefix: string): ReactNode[] {
  const nodes: ReactNode[] = [];
  const re = /(`[^`]+`|\*\*[^*]+\*\*|\*[^*]+\*|\[[^\]]+\]\([^)]+\))/g;
  let last = 0;
  let k = 0;
  const pushText = (s: string) => {
    if (s) nodes.push(<span key={`${keyPrefix}-t${k++}`}>{s}</span>);
  };
  for (const m of text.matchAll(re)) {
    const idx = m.index ?? 0;
    pushText(text.slice(last, idx));
    const tok = m[0];
    if (tok.startsWith("`")) {
      nodes.push(<code key={`${keyPrefix}-c${k++}`}>{tok.slice(1, -1)}</code>);
    } else if (tok.startsWith("**")) {
      nodes.push(<strong key={`${keyPrefix}-b${k++}`}>{tok.slice(2, -2)}</strong>);
    } else if (tok.startsWith("*")) {
      nodes.push(<em key={`${keyPrefix}-i${k++}`}>{tok.slice(1, -1)}</em>);
    } else {
      const lm = /^\[([^\]]+)\]\(([^)]+)\)$/.exec(tok);
      if (lm && ALLOWED_SCHEMES.test(lm[2].trim())) {
        nodes.push(
          <a key={`${keyPrefix}-a${k++}`} href={lm[2].trim()} target="_blank" rel="noreferrer">
            {lm[1]}
          </a>
        );
      } else if (lm) {
        // Disallowed scheme: render text only, no link.
        nodes.push(<span key={`${keyPrefix}-x${k++}`}>{lm[1]}</span>);
      } else {
        pushText(tok);
      }
    }
    last = idx + tok.length;
  }
  pushText(text.slice(last));
  return nodes;
}

export function SafeMarkdown({ text }: { text: string }) {
  const blocks = splitBlocks(text);
  return (
    <>
      {blocks.map((b, bi) => {
        if (b.type === "code") {
          return (
            <pre key={bi} className="md-code">
              <code>{b.lines.join("\n")}</code>
            </pre>
          );
        }
        if (b.type === "list") {
          return (
            <ul key={bi} className="md-list">
              {b.lines.map((ln, li) => (
                <li key={li}>{renderInline(ln, `b${bi}l${li}`)}</li>
              ))}
            </ul>
          );
        }
        return (
          <p key={bi} className="md-para">
            {b.lines.map((ln, li) => (
              <span key={li}>
                {li > 0 && <br />}
                {renderInline(ln, `b${bi}p${li}`)}
              </span>
            ))}
          </p>
        );
      })}
    </>
  );
}
