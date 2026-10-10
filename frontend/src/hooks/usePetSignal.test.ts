import { describe, expect, it } from "vitest";
import { renderHook } from "@testing-library/react";
import type { ChatStatus } from "../types/chat";
import { usePetSignal, type ViewMode } from "./usePetSignal";

function run(start: { mode: ViewMode; status: ChatStatus }) {
  return renderHook((p: { mode: ViewMode; status: ChatStatus }) => usePetSignal(p.mode, p.status), { initialProps: start });
}

describe("usePetSignal", () => {
  it("flags a reply that finished while the user was on the pet", () => {
    const r = run({ mode: "pet", status: "streaming" });
    r.rerender({ mode: "pet", status: "complete" });
    expect(r.result.current.replyReady).toBe(true);
  });

  it("clears the flag once chat is opened", () => {
    const r = run({ mode: "pet", status: "streaming" });
    r.rerender({ mode: "pet", status: "complete" });
    r.rerender({ mode: "chat", status: "complete" });
    r.rerender({ mode: "pet", status: "complete" });
    expect(r.result.current.replyReady).toBe(false);
  });

  it("does not flag replies the user watched finish in chat", () => {
    const r = run({ mode: "chat", status: "streaming" });
    r.rerender({ mode: "chat", status: "complete" });
    r.rerender({ mode: "pet", status: "complete" });
    expect(r.result.current.replyReady).toBe(false);
  });

  it("keeps the pet worried until the error has been seen in chat", () => {
    const r = run({ mode: "pet", status: "streaming" });
    r.rerender({ mode: "pet", status: "error" });
    expect(r.result.current.petStatus).toBe("error");
    r.rerender({ mode: "chat", status: "error" });
    r.rerender({ mode: "pet", status: "error" });
    expect(r.result.current.petStatus).toBe("idle");
  });

  it("worries again on the next failure", () => {
    const r = run({ mode: "chat", status: "error" });
    r.rerender({ mode: "pet", status: "error" });
    expect(r.result.current.petStatus).toBe("idle");
    r.rerender({ mode: "pet", status: "sending" });
    r.rerender({ mode: "pet", status: "error" });
    expect(r.result.current.petStatus).toBe("error");
  });
});

describe("usePetSignal error seen", () => {
  it("does not count Settings as having seen the error", () => {
    const r = run({ mode: "pet", status: "error" });
    r.rerender({ mode: "settings", status: "error" });
    r.rerender({ mode: "pet", status: "error" });
    expect(r.result.current.petStatus).toBe("error");
  });
});

describe("usePetSignal reply ready via Settings", () => {
  it("keeps the bubble when only Settings was opened", () => {
    const r = run({ mode: "pet", status: "streaming" });
    r.rerender({ mode: "pet", status: "complete" });
    r.rerender({ mode: "settings", status: "complete" });
    r.rerender({ mode: "pet", status: "complete" });
    expect(r.result.current.replyReady).toBe(true);
  });
});
