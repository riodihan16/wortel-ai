import { describe, expect, it, afterEach } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import type { ChatMessage } from "../../types/chat";
import { MessageItem } from "./MessageItem";

function assistant(over: Partial<ChatMessage> = {}): ChatMessage {
  return { id: "m1", role: "assistant", content: "", status: "complete", ...over };
}

afterEach(() => cleanup());

describe("MessageItem generated image", () => {
  it("renders the image through the local asset route with a non-empty alt", () => {
    render(<MessageItem message={assistant({ imageMime: "image/png" })} isLast busy={false} onRetry={() => {}} />);
    const el = screen.getByRole("img", { name: "Gambar hasil pembuatan" });
    expect(el.getAttribute("src")).toBe("/attachments/m1");
    expect(el.getAttribute("alt")).toBeTruthy();
  });

  it("does not render raw base64 in the DOM", () => {
    const { container } = render(
      <MessageItem message={assistant({ imageMime: "image/png" })} isLast busy={false} onRetry={() => {}} />
    );
    expect(container.innerHTML).not.toContain("base64");
  });

  it("hides the empty text body when the image is the content", () => {
    const { container } = render(
      <MessageItem message={assistant({ imageMime: "image/png" })} isLast busy={false} onRetry={() => {}} />
    );
    expect(container.querySelector(".msg-body")).toBeNull();
  });

  it("renders no image block for a plain assistant reply", () => {
    const { container } = render(
      <MessageItem message={assistant({ content: "halo" })} isLast busy={false} onRetry={() => {}} />
    );
    expect(container.querySelector(".msg-image")).toBeNull();
  });
});
