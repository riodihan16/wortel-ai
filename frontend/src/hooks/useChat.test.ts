import { describe, expect, it, vi, beforeEach } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ChatEvent } from "../types/chat";

// Test doubles for the Go bindings. Keep the responses in refs so each test
// can steer the backend without re-mocking the module.
const backend = {
  history: vi.fn(async () => [] as unknown[]),
  list: vi.fn(async () => [] as unknown[]),
  send: vi.fn(async () => "req-1"),
  retry: vi.fn(async () => "req-2"),
  cancel: vi.fn(async () => {}),
  rename: vi.fn(async () => {}),
  remove: vi.fn(async () => {}),
};

vi.mock("../../wailsjs/go/main/App", () => ({
  History: (...a: unknown[]) => backend.history(...(a as [])),
  ListConversations: () => backend.list(),
  SendMessage: (...a: unknown[]) => backend.send(...(a as [])),
  RetryMessage: (...a: unknown[]) => backend.retry(...(a as [])),
  CancelGeneration: (...a: unknown[]) => backend.cancel(...(a as [])),
  RenameConversation: (...a: unknown[]) => backend.rename(...(a as [])),
  DeleteConversation: (...a: unknown[]) => backend.remove(...(a as [])),
}));

import { runtimeMock } from "../test/wails";
import { useChat } from "../hooks/useChat";

function emit(ev: Partial<ChatEvent> & { type: ChatEvent["type"] }, conversationId: string) {
  runtimeMock.emit("chat:event", {
    requestId: ev.requestId ?? "req-1",
    conversationId,
    messageId: ev.messageId ?? "m1",
    type: ev.type,
    delta: ev.delta,
    error: ev.error,
  });
}

async function mountChat() {
  const r = renderHook(() => useChat());
  await waitFor(() => expect(r.result.current.conversationId).not.toBe(""));
  return r;
}

beforeEach(() => {
  runtimeMock.reset();
  vi.clearAllMocks();
  backend.history.mockResolvedValue([]);
  backend.list.mockResolvedValue([]);
  backend.send.mockResolvedValue("req-1");
  backend.retry.mockResolvedValue("req-2");
});

describe("useChat streaming reducer", () => {
  it("adopts the backend requestId from the started event", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;

    act(() => emit({ type: "started", messageId: "m1" }, conv));
    expect(result.current.status).toBe("streaming");

    act(() => emit({ type: "delta", delta: "Ha", messageId: "m1" }, conv));
    act(() => emit({ type: "delta", delta: "lo", messageId: "m1" }, conv));
    expect(result.current.messages.at(-1)?.content).toBe("Halo");
  });

  it("ignores deltas whose requestId does not match the active request", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;

    act(() => emit({ type: "started", requestId: "req-1", messageId: "m1" }, conv));
    act(() => emit({ type: "delta", requestId: "OTHER", delta: "X", messageId: "m1" }, conv));
    expect(result.current.messages.at(-1)?.content).toBe("");
  });

  it("ignores events for a different conversation", async () => {
    const { result } = await mountChat();
    act(() => emit({ type: "started", messageId: "m1" }, "some-other-conv"));
    expect(result.current.status).toBe("idle");
    expect(result.current.messages).toHaveLength(0);
  });

  it("marks the message complete and clears busy on completed", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    act(() => emit({ type: "started", messageId: "m1" }, conv));
    act(() => emit({ type: "delta", delta: "hi", messageId: "m1" }, conv));
    act(() => emit({ type: "completed", messageId: "m1" }, conv));
    expect(result.current.status).toBe("complete");
    expect(result.current.messages.at(-1)?.status).toBe("complete");
  });

  it("keeps partial text and sets cancelled on a cancelled event", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    act(() => emit({ type: "started", messageId: "m1" }, conv));
    act(() => emit({ type: "delta", delta: "partial", messageId: "m1" }, conv));
    act(() => emit({ type: "cancelled", messageId: "m1" }, conv));
    expect(result.current.status).toBe("cancelled");
    expect(result.current.messages.at(-1)?.content).toBe("partial");
  });

  it("surfaces the error message on a failed event", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    act(() => emit({ type: "started", messageId: "m1" }, conv));
    act(() =>
      emit(
        {
          type: "failed",
          messageId: "m1",
          error: { code: "http_401", message: "API key ditolak", retryable: false },
        },
        conv
      )
    );
    expect(result.current.status).toBe("error");
    // The reason belongs to the failed reply; the global error stays free
    // for errors that belong to no reply (rename, delete, send rejection).
    expect(result.current.messages.at(-1)).toMatchObject({ status: "error", errorText: "API key ditolak" });
    expect(result.current.error).toBeNull();
  });
});

describe("useChat send guard", () => {
  it("rejects an empty or whitespace-only message", async () => {
    const { result } = await mountChat();
    await act(async () => {
      await result.current.send("   ");
    });
    expect(backend.send).not.toHaveBeenCalled();
  });

  it("does not fire a second request while one is in flight", async () => {
    const { result } = await mountChat();
    await act(async () => {
      await result.current.send("halo");
    });
    await act(async () => {
      await result.current.send("lagi");
    });
    expect(backend.send).toHaveBeenCalledTimes(1);
  });
});

describe("useChat cancel guard", () => {
  it("releases the busy guard when CancelGeneration rejects", async () => {
    const { result } = await mountChat();
    await act(async () => {
      await result.current.send("halo");
    });
    expect(result.current.busy).toBe(true);

    backend.cancel.mockRejectedValueOnce(new Error("IPC gagal"));
    await act(async () => {
      await result.current.cancel();
    });

    // Guard must be released even though cancel failed, otherwise every
    // action guarded by busyNow() would stay locked forever.
    expect(result.current.busy).toBe(false);
    await act(async () => {
      await result.current.send("lagi");
    });
    expect(backend.send).toHaveBeenCalledTimes(2);
  });

  it("releases both guards when cancel fails during the streaming phase", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    await act(async () => {
      await result.current.send("halo");
    });
    // Backend adopts the requestId only once "started" arrives.
    act(() => emit({ type: "started", messageId: "m1" }, conv));
    expect(result.current.status).toBe("streaming");

    backend.cancel.mockRejectedValueOnce(new Error("IPC gagal"));
    await act(async () => {
      await result.current.cancel();
    });

    expect(result.current.busy).toBe(false);
    await act(async () => {
      await result.current.send("lagi");
    });
    expect(backend.send).toHaveBeenCalledTimes(2);
  });
});

describe("useChat retry", () => {
  async function failOnce() {
    const r = await mountChat();
    const conv = r.result.current.conversationId;
    await act(() => r.result.current.send("halo"));
    act(() => emit({ type: "started", messageId: "m1" }, conv));
    act(() => emit({ type: "failed", messageId: "m1", error: { code: "x", message: "429", retryable: true } }, conv));
    return { ...r, conv };
  }

  it("re-runs the failed turn without a second user bubble", async () => {
    const { result, conv } = await failOnce();
    await act(() => result.current.retry());

    expect(backend.retry).toHaveBeenCalledWith(conv, "halo");
    expect(backend.send).toHaveBeenCalledTimes(1);
    const users = result.current.messages.filter((m) => m.role === "user");
    expect(users).toHaveLength(1);
    expect(result.current.messages.some((m) => m.status === "error")).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it("streams the retried reply into a fresh bubble", async () => {
    const { result, conv } = await failOnce();
    await act(() => result.current.retry());
    act(() => emit({ type: "started", requestId: "req-2", messageId: "m2" }, conv));
    act(() => emit({ type: "delta", requestId: "req-2", delta: "ok", messageId: "m2" }, conv));

    expect(result.current.messages.at(-1)).toMatchObject({ id: "m2", content: "ok" });
  });

  it("does nothing while a request is in flight", async () => {
    const { result } = await mountChat();
    await act(() => result.current.send("halo"));
    await act(() => result.current.retry());
    expect(backend.retry).not.toHaveBeenCalled();
  });

  it("clearError drops a banner error", async () => {
    const { result } = await mountChat();
    backend.send.mockRejectedValueOnce(new Error("IPC putus"));
    await act(() => result.current.send("halo"));
    expect(result.current.error).toBe("IPC putus");
    act(() => result.current.clearError());
    expect(result.current.error).toBeNull();
  });

  it("keeps the failed reply when the retry itself is rejected", async () => {
    const { result } = await failOnce();
    backend.retry.mockRejectedValueOnce(new Error("generation already active"));
    await act(() => result.current.retry());
    expect(result.current.messages.some((m) => m.status === "error")).toBe(true);
  });
});

describe("useChat request lifecycle races", () => {
  it("does not stay busy when the reply fails before SendMessage resolves", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    // Fast failure (e.g. local server down): started + failed arrive over the
    // event bus before the binding call returns its requestId.
    backend.send.mockImplementationOnce(async () => {
      emit({ type: "started", requestId: "req-9", messageId: "m9" }, conv);
      emit({ type: "failed", requestId: "req-9", messageId: "m9", error: { code: "network", message: "x", retryable: true } }, conv);
      return "req-9";
    });
    await act(() => result.current.send("halo"));
    await act(() => result.current.send("lagi"));
    expect(backend.send).toHaveBeenCalledTimes(2);
  });

  it("marks the stopped reply cancelled even though the late event is dropped", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    await act(() => result.current.send("halo"));
    act(() => emit({ type: "started", messageId: "m1" }, conv));
    act(() => emit({ type: "delta", delta: "seten", messageId: "m1" }, conv));
    await act(() => result.current.cancel());
    act(() => emit({ type: "cancelled", messageId: "m1" }, conv));

    expect(result.current.messages.at(-1)).toMatchObject({ id: "m1", status: "cancelled", content: "seten" });
  });
});

describe("useChat attachments", () => {
  const att = { id: "a1", name: "foto.png", mime: "image/png", size: 10, kind: "image" as const };

  it("passes attachment ids to SendMessage and shows them on the user bubble", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    await act(() => result.current.send("lihat", [att]));
    expect(backend.send).toHaveBeenCalledWith(conv, "lihat", ["a1"]);
    expect(result.current.messages.at(-1)).toMatchObject({ role: "user", content: "lihat", attachments: [att] });
  });

  it("accepts a message with attachments and no text", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    await act(() => result.current.send("   ", [att]));
    expect(backend.send).toHaveBeenCalledWith(conv, "", ["a1"]);
  });

  it("still rejects a message with neither text nor attachments", async () => {
    const { result } = await mountChat();
    await act(() => result.current.send("   ", []));
    expect(backend.send).not.toHaveBeenCalled();
  });

  it("loads attachment metadata from history", async () => {
    backend.list.mockResolvedValue([{ id: "c9", title: "t" }]);
    backend.history.mockResolvedValue([{ id: "u1", role: "user", content: "x", status: "complete", attachments: [att] }]);
    const { result } = renderHook(() => useChat());
    await waitFor(() => expect(result.current.messages).toHaveLength(1));
    expect(result.current.messages[0].attachments).toEqual([att]);
  });
});
