import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { createElement } from "react";
import { SafeMarkdown } from "./SafeMarkdown";

// These guard the security-critical part of rendering untrusted AI output:
// raw HTML must stay inert, and only http/https links become anchors.
describe("SafeMarkdown link and HTML policy", () => {
  it("does not render raw HTML from AI text", () => {
    const { container } = render(
      createElement(SafeMarkdown, { text: '<img src=x onerror=alert(1)><script>alert(2)</script>' })
    );
    // The invariant is structural: no executable element may reach the DOM.
    expect(container.querySelectorAll("img")).toHaveLength(0);
    expect(container.querySelectorAll("script")).toHaveLength(0);
    // "alert" can only survive as inert text, never as a live handler.
    expect(container.innerHTML).toContain("alert");
  });

  it("renders an https link as an anchor with a safe rel", () => {
    render(createElement(SafeMarkdown, { text: "[situs](https://example.com/x)" }));
    const a = screen.getByText("situs").closest("a");
    expect(a).not.toBeNull();
    expect(a!.getAttribute("href")).toBe("https://example.com/x");
    expect(a!.getAttribute("rel")).toContain("noreferrer");
  });

  it("drops a javascript: link to plain text with no anchor", () => {
    const { container } = render(
      createElement(SafeMarkdown, { text: "[jahat](javascript:alert(1))" })
    );
    expect(container.querySelectorAll("a")).toHaveLength(0);
    expect(container.textContent).toContain("jahat");
  });

  it("keeps a data: link non-anchor", () => {
    const { container } = render(
      createElement(SafeMarkdown, { text: "[x](data:text/html,<b>hi</b>)" })
    );
    expect(container.querySelectorAll("a")).toHaveLength(0);
  });
});
