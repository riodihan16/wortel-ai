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
};

vi.mock("../../wailsjs/go/main/App", () => ({
  ListProviders: () => backend.list(),
  ActiveProvider: () => backend.active(),
  DeleteProvider: (...a: unknown[]) => backend.remove(...(a as [])),
  SetActiveProvider: (...a: unknown[]) => backend.activate(...(a as [])),
  SaveProvider: vi.fn(),
  SetActiveModel: vi.fn(),
  TestProvider: vi.fn(),
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
