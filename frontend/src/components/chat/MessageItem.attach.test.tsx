import { describe, expect, it, afterEach } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import type { ChatMessage } from "../../types/chat";
import { MessageItem } from "./MessageItem";

const img = { id: "a", name: "foto.png", mime: "image/png", size: 100, kind: "image" as const };
const pdf = { id: "b", name: "nota.pdf", mime: "application/pdf", size: 3072, kind: "pdf" as const };

function user(over: Partial<ChatMessage> = {}): ChatMessage {
  return { id: "u1", role: "user", content: "lihat", status: "complete", ...over };
}

afterEach(() => cleanup());

describe("MessageItem attachments", () => {
  it("renders image thumbnails and file chips for a user turn", () => {
    render(<MessageItem message={user({ attachments: [img, pdf] })} isLast busy={false} onRetry={() => {}} />);
    const el = screen.getByRole("img", { name: "foto.png" });
    expect(el.getAttribute("src")).toBe("/attachments/a");
    expect(screen.getByText("nota.pdf")).toBeTruthy();
    expect(screen.getByText("3.0 KB")).toBeTruthy();
    expect(screen.getByRole("group", { name: "Lampiran: 2 berkas" })).toBeTruthy();
  });

  it("toggles a thumbnail to full size", () => {
    render(<MessageItem message={user({ attachments: [img] })} isLast busy={false} onRetry={() => {}} />);
    const btn = screen.getByRole("button", { name: "foto.png" });
    expect(btn.getAttribute("aria-pressed")).toBe("false");
    fireEvent.click(btn);
    expect(btn.getAttribute("aria-pressed")).toBe("true");
  });

  it("renders nothing extra without attachments", () => {
    const { container } = render(<MessageItem message={user()} isLast busy={false} onRetry={() => {}} />);
    expect(container.querySelector(".att-list")).toBeNull();
  });

  it("hides the empty bubble when only attachments were sent", () => {
    const { container } = render(<MessageItem message={user({ content: "", attachments: [img] })} isLast busy={false} onRetry={() => {}} />);
    expect(container.querySelector(".msg-body")).toBeNull();
    expect(container.querySelector(".att-list")).not.toBeNull();
  });
});
