import { describe, expect, it } from "vitest";
import { markdownToText } from "./markdownToText";

describe("markdownToText", () => {
  it("drops emphasis and inline code marks", () => {
    expect(markdownToText("Deploy **Kamis**, _pagi_, pakai `git`.")).toBe("Deploy Kamis, pagi, pakai git.");
  });
  it("keeps list markers and paragraphs", () => {
    expect(markdownToText("# Judul\n\n1. **Satu**\n2. Dua\n\n- a\n* b")).toBe("Judul\n\n1. Satu\n2. Dua\n\n- a\n- b");
  });
  it("keeps code block content without fences", () => {
    expect(markdownToText("Jalankan:\n\n```bash\ngit checkout -b x\n```")).toBe("Jalankan:\n\ngit checkout -b x");
  });
  it("turns links into text with the address", () => {
    expect(markdownToText("Lihat [dok](https://x.dev) ya")).toBe("Lihat dok (https://x.dev) ya");
  });
  it("drops quote markers and stray emphasis inside words safely", () => {
    expect(markdownToText("> Catatan **penting**\nsnake_case_name")).toBe("Catatan penting\nsnake_case_name");
  });
});
