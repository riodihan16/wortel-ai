import { afterEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useCopy } from "./useCopy";

function clipboard(writeText: ((t: string) => Promise<void>) | undefined) {
  Object.defineProperty(navigator, "clipboard", { value: writeText ? { writeText } : undefined, configurable: true });
}

afterEach(() => {
  vi.useRealTimers();
});

describe("useCopy", () => {
  it("copies, confirms, then settles back", async () => {
    vi.useFakeTimers();
    const write = vi.fn(async () => {});
    clipboard(write);
    const { result } = renderHook(() => useCopy());
    await act(() => result.current.copy("halo"));
    expect(write).toHaveBeenCalledWith("halo");
    expect(result.current.state).toBe("copied");
    act(() => vi.advanceTimersByTime(1600));
    expect(result.current.state).toBe("idle");
  });

  it("says so when the clipboard is unavailable", async () => {
    clipboard(undefined);
    const { result } = renderHook(() => useCopy());
    await act(() => result.current.copy("halo"));
    expect(result.current.state).toBe("failed");
  });

  it("says so when the write is refused", async () => {
    clipboard(vi.fn(async () => { throw new Error("denied"); }));
    const { result } = renderHook(() => useCopy());
    await act(() => result.current.copy("halo"));
    expect(result.current.state).toBe("failed");
  });
});
