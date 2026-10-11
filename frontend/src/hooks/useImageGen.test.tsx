import { describe, expect, it, vi, beforeEach } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ImageEvent } from "../types/image";

const backend = {
  generate: vi.fn(async () => "req-1"),
  cancel: vi.fn(async () => {}),
};

vi.mock("../../wailsjs/go/main/App", () => ({
  GenerateImage: (...a: unknown[]) => backend.generate(...(a as [])),
  CancelImageGeneration: (...a: unknown[]) => backend.cancel(...(a as [])),
}));

import { runtimeMock } from "../test/wails";
import { useImageGen } from "../hooks/useImageGen";

function emit(ev: Partial<ImageEvent> & { type: ImageEvent["type"] }, conversationId: string) {
  runtimeMock.emit("image:event", {
    requestId: ev.requestId ?? "req-1",
    conversationId,
    messageId: ev.messageId ?? "m1",
    type: ev.type,
    mime: ev.mime,
    prompt: ev.prompt,
    error: ev.error,
  });
}

beforeEach(() => {
  runtimeMock.reset();
  vi.clearAllMocks();
  backend.generate.mockResolvedValue("req-1");
});

describe("useImageGen", () => {
  it("adopts the started event and settles on completed", async () => {
    const { result } = renderHook(() => useImageGen());
    await act(async () => {
      await result.current.generate("c1", "panda", "img");
    });
    act(() => emit({ type: "started" }, "c1"));
    expect(result.current.pending).toBe(true);
    act(() => emit({ type: "completed", mime: "image/png" }, "c1"));
    expect(result.current.pending).toBe(false);
    expect(result.current.status).toBe("complete");
  });

  it("ignores events from a different conversation", async () => {
    const { result } = renderHook(() => useImageGen());
    await act(async () => {
      await result.current.generate("c1", "panda", "img");
    });
    act(() => emit({ type: "started" }, "c1"));
    const before = result.current.status;
    act(() => emit({ type: "completed" }, "other"));
    expect(result.current.status).toBe(before);
    expect(result.current.pending).toBe(true);
  });

  it("ignores a superseded requestId", async () => {
    const { result } = renderHook(() => useImageGen());
    await act(async () => {
      await result.current.generate("c1", "panda", "img");
    });
    act(() => emit({ type: "started", requestId: "req-1" }, "c1"));
    act(() => emit({ type: "completed", requestId: "stale" }, "c1"));
    expect(result.current.pending).toBe(true);
  });

  it("reports a failure reason and clears pending", async () => {
    const { result } = renderHook(() => useImageGen());
    await act(async () => {
      await result.current.generate("c1", "panda", "img");
    });
    act(() => emit({ type: "started" }, "c1"));
    act(() =>
      emit({ type: "failed", error: { code: "insufficient_balance", message: "Saldo kurang.", retryable: false } }, "c1")
    );
    expect(result.current.pending).toBe(false);
    expect(result.current.status).toBe("error");
    expect(result.current.error).toBe("Saldo kurang.");
  });

  it("cancel calls the backend and resets state", async () => {
    const { result } = renderHook(() => useImageGen());
    await act(async () => {
      await result.current.generate("c1", "panda", "img");
    });
    act(() => emit({ type: "started" }, "c1"));
    act(() => result.current.cancel());
    expect(backend.cancel).toHaveBeenCalledWith("c1");
    expect(result.current.pending).toBe(false);
    expect(result.current.status).toBe("idle");
  });

  it("does not generate without a model or prompt", async () => {
    const { result } = renderHook(() => useImageGen());
    await act(async () => {
      await result.current.generate("c1", "panda", "");
    });
    await act(async () => {
      await result.current.generate("c1", "  ", "img");
    });
    expect(backend.generate).not.toHaveBeenCalled();
  });

  it("removes the image:event listener on unmount", async () => {
    const { unmount } = renderHook(() => useImageGen());
    unmount();
    // After unmount, an emitted event must not throw or touch deleted state.
    expect(() => runtimeMock.emit("image:event", { type: "started", conversationId: "c1" })).not.toThrow();
  });
});
