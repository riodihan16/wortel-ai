// markdownToText turns a reply's markdown into the plain text a user expects
// after "Salin": no ** _ ` # > marks, code kept without fences, links as
// "text (url)", list markers normalised to "- " or "1. ".
//
// What is not markdown stays as written: URLs, inline code and backslash
// escapes are held aside (placeholders) before the emphasis passes, and
// emphasis only matches at word boundaries, so 2*3*4, (1)*(2), snake_case,
// __init__.py and https://x/_a_/ survive.
const HOLD = "\u0000";

export function markdownToText(md: string): string {
  const out: string[] = [];
  let inFence = false;
  for (const raw of md.replace(/\r\n/g, "\n").split("\n")) {
    if (/^\s*(```|~~~)/.test(raw)) {
      inFence = !inFence;
      continue;
    }
    if (inFence) {
      out.push(raw);
      continue;
    }
    const held: string[] = [];
    const hold = (s: string) => `${HOLD}${held.push(s) - 1}${HOLD}`;
    let line = raw
      .replace(/^\s{0,3}#{1,6}\s+/, "")
      .replace(/^\s{0,3}>\s?/, "")
      .replace(/^(\s*)[*+]\s+/, "$1- ");
    line = line
      .replace(/\\([\\`*_{}[\]()#+\-.!~>])/g, (_, c: string) => hold(c))
      .replace(/!\[([^\]]*)\]\([^)]*\)/g, "$1")
      .replace(/\[([^\]]+)\]\(([^)\s]+)[^)]*\)/g, (_, t: string, u: string) => `${t} (${hold(u)})`)
      .replace(/https?:\/\/[^\s)]+/g, (u) => hold(u))
      .replace(/`([^`]+)`/g, (_, c: string) => hold(c));
    line = line
      .replace(/(^|[^\w*])\*\*(?=\S)([\s\S]*?\S)\*\*(?![\w*])/g, "$1$2")
      .replace(/(^|[^\w_])__(?=\S)([\s\S]*?\S)__(?![\w_.])/g, "$1$2")
      .replace(/(^|[^\w*)])\*(?=[^\s*])([^*]*?[^\s*])\*(?![\w(*])/g, "$1$2")
      .replace(/(^|[^\w_])_(?=[^\s_])([^_]*?[^\s_])_(?![\w_])/g, "$1$2")
      .replace(/~~(?=\S)([\s\S]*?\S)~~/g, "$1");
    out.push(line.replace(/\u0000(\d+)\u0000/g, (_, i: string) => held[Number(i)]));
  }
  return out.join("\n").replace(/\n{3,}/g, "\n\n").trim();
}
