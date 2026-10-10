import { beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ProviderProfile } from "../types/provider";

const profile = (id: string): ProviderProfile => ({
  id,
  name: id,
  baseURL: "https://api.example.com/v1",
  model: "m",
  models: [{ id: "m", alias: "" }],
  activeModel: "m",
  credentialRef: "",
  hasCredential: true,
  enabled: true,
});

const backend = {
  list: vi.fn(async () => [] as ProviderProfile[]),
  active: vi.fn(async () => null as ProviderProfile | null),
  remove: vi.fn(async () => {}),
  activate: vi.fn(async () => {}),
  test: vi.fn(async () => ({ ok: false, message: "401" })),
};

vi.mock("../../wailsjs/go/main/App", () => ({
  ListProviders: () => backend.list(),
  ActiveProvider: () => backend.active(),
  DeleteProvider: (...a: unknown[]) => backend.remove(...(a as [])),
  SetActiveProvider: (...a: unknown[]) => backend.activate(...(a as [])),
  SaveProvider: vi.fn(),
  SetActiveModel: vi.fn(),
  TestProvider: (...a: unknown[]) => backend.test(...(a as [])),
}));

import { useProviders } from "./useProviders";

beforeEach(() => {
  vi.clearAllMocks();
  backend.list.mockResolvedValue([profile("a"), profile("b")]);
  backend.active.mockResolvedValue(profile("b"));
});

describe("useProviders active provider", () => {
  it("marks the backend's active provider on first load", async () => {
    const { result } = renderHook(() => useProviders());
    await waitFor(() => expect(result.current.activeId).toBe("b"));
  });

  it("clears the active id when the backend reports none", async () => {
    backend.active.mockResolvedValue(null);
    const { result } = renderHook(() => useProviders());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.activeId).toBe("");
  });

  it("takes the active id from the backend after switching", async () => {
    const { result } = renderHook(() => useProviders());
    await waitFor(() => expect(result.current.activeId).toBe("b"));

    backend.active.mockResolvedValue(profile("a"));
    await act(() => result.current.activate("a"));

    expect(backend.activate).toHaveBeenCalledWith("a");
    expect(result.current.activeId).toBe("a");
  });

  it("keeps the active id when the backend refuses a switch", async () => {
    const { result } = renderHook(() => useProviders());
    await waitFor(() => expect(result.current.activeId).toBe("b"));

    backend.activate.mockRejectedValueOnce(new Error("provider tidak ditemukan"));
    await act(() => result.current.activate("zzz"));

    expect(result.current.activeId).toBe("b");
    expect(result.current.error).toBe("provider tidak ditemukan");
  });
});

describe("useProviders test result", () => {
  it("clearTest drops the last result", async () => {
    const { result } = renderHook(() => useProviders());
    await act(() => result.current.runTest({} as never));
    expect(result.current.test?.message).toBe("401");
    act(() => result.current.clearTest());
    expect(result.current.test).toBeNull();
  });

  it("ignores a test that finishes after the form it belonged to was closed", async () => {
    let finish: (v: { ok: boolean; message: string }) => void = () => {};
    backend.test.mockImplementationOnce(() => new Promise((r) => { finish = r; }));
    const { result } = renderHook(() => useProviders());
    let pending: Promise<unknown> = Promise.resolve();
    act(() => { pending = result.current.runTest({} as never); });
    act(() => result.current.clearTest());
    await act(async () => { finish({ ok: false, message: "basi" }); await pending; });
    expect(result.current.test).toBeNull();
    expect(result.current.testing).toBe(false);
  });
});
