import { describe, expect, it, vi, beforeEach } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { AttachmentMeta } from "../types/chat";

function meta(id: string): AttachmentMeta {
  return { id, name: id + ".png", mime: "image/png", size: 3, kind: "image" };
}

const backend = {
  pick: vi.fn(async (): Promise<AttachmentMeta[]> => []),
  paths: vi.fn(async (_p: string[]): Promise<AttachmentMeta[]> => []),
  data: vi.fn(async (_n: string, _m: string, _d: string): Promise<AttachmentMeta> => meta("x")),
  unstage: vi.fn(async (_id: string) => {}),
};

vi.mock("../../wailsjs/go/main/App", () => ({
  PickAttachments: () => backend.pick(),
  StageAttachmentPaths: (p: string[]) => backend.paths(p),
  StageAttachmentData: (n: string, m: string, d: string) => backend.data(n, m, d),
  UnstageAttachment: (id: string) => backend.unstage(id),
}));

import { useAttachments, MAX_ATTACHMENTS } from "./useAttachments";

beforeEach(() => {
  vi.clearAllMocks();
  backend.pick.mockResolvedValue([]);
  backend.paths.mockResolvedValue([]);
  backend.unstage.mockResolvedValue(undefined);
});

describe("useAttachments", () => {
  it("adds what the dialog returns", async () => {
    backend.pick.mockResolvedValueOnce([meta("a"), meta("b")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    expect(result.current.pending.map((a) => a.id)).toEqual(["a", "b"]);
    expect(result.current.error).toBeNull();
  });

  it("keeps the list and reports the reason when the backend refuses", async () => {
    backend.paths.mockRejectedValueOnce("foto.exe: tipe berkas tidak didukung");
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.stagePaths(["C:/x/foto.exe"]));
    expect(result.current.pending).toEqual([]);
    expect(result.current.error).toBe("foto.exe: tipe berkas tidak didukung");
  });

  it("refuses a fifth attachment locally", async () => {
    backend.pick.mockResolvedValueOnce([meta("a"), meta("b"), meta("c"), meta("d")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    expect(result.current.pending).toHaveLength(MAX_ATTACHMENTS);
    await act(() => result.current.stagePaths(["x.png"]));
    expect(backend.paths).not.toHaveBeenCalled();
    expect(result.current.error).toContain("Maksimal 4");
    expect(result.current.pending).toHaveLength(MAX_ATTACHMENTS);
  });

  it("reads pasted files as base64 and stages each", async () => {
    backend.data.mockResolvedValueOnce(meta("p1"));
    const file = new File([new Uint8Array([1, 2, 3])], "tempel.png", { type: "image/png" });
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.stageFiles([file]));
    expect(backend.data).toHaveBeenCalledWith("tempel.png", "image/png", "AQID");
    expect(result.current.pending[0].id).toBe("p1");
  });

  it("refuses a pasted file over 15 MB without reading it", async () => {
    const small = new File([new Uint8Array([1])], "kecil.png", { type: "image/png" });
    const big = new File([new Uint8Array([1])], "besar.png", { type: "image/png" });
    Object.defineProperty(big, "size", { value: 15 * 1024 * 1024 + 1 });
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.stageFiles([small, big]));
    expect(backend.data).not.toHaveBeenCalled();
    expect(result.current.error).toBe("besar.png: melebihi 15 MB");
    expect(result.current.pending).toEqual([]);
  });

  it("unstages already-staged files when a later paste fails", async () => {
    backend.data.mockResolvedValueOnce(meta("p1")).mockRejectedValueOnce("rusak");
    const f = (n: string) => new File([new Uint8Array([1])], n, { type: "image/png" });
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.stageFiles([f("a.png"), f("b.png")]));
    expect(result.current.pending).toEqual([]);
    expect(backend.unstage).toHaveBeenCalledWith("p1");
    expect(result.current.error).toBe("rusak");
  });

  it("remove drops the item and tells the backend", async () => {
    backend.pick.mockResolvedValueOnce([meta("a"), meta("b")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    await act(() => result.current.remove("a"));
    expect(result.current.pending.map((a) => a.id)).toEqual(["b"]);
    expect(backend.unstage).toHaveBeenCalledWith("a");
  });

  it("reset clears the list without touching the backend", async () => {
    backend.pick.mockResolvedValueOnce([meta("a")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    act(() => result.current.reset());
    expect(result.current.pending).toEqual([]);
    expect(backend.unstage).not.toHaveBeenCalled();
  });
});
