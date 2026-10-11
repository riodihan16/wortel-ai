import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { runtimeMock } from "../../test/wails";
import type { AttachmentMeta } from "../../types/chat";
import type { AttachmentsApi } from "../../hooks/useAttachments";
import { ChatPanel } from "./ChatPanel";

function meta(id: string, kind: AttachmentMeta["kind"] = "image"): AttachmentMeta {
  return { id, name: `${id}.${kind === "image" ? "png" : "pdf"}`, mime: kind === "image" ? "image/png" : "application/pdf", size: 2048, kind };
}

function api(over: Partial<AttachmentsApi> = {}): AttachmentsApi {
  return {
    pending: [],
    error: null,
    pick: vi.fn(async () => {}),
    stagePaths: vi.fn(async () => {}),
    stageFiles: vi.fn(async () => {}),
    remove: vi.fn(async () => {}),
    reset: vi.fn(),
    clearError: vi.fn(),
    ...over,
  };
}

function renderPanel(over: { attach?: AttachmentsApi; draft?: string; busy?: boolean } = {}) {
  const onSend = vi.fn();
  const attach = over.attach ?? api();
  const utils = render(
    <ChatPanel
      messages={[]}
      status="idle"
      error={null}
      busy={over.busy ?? false}
      conversations={[]}
      conversationId="c1"
      onSend={onSend}
      onCancel={() => {}}
      onRetry={() => {}}
      onDismissError={() => {}}
      onNew={() => {}}
      onSwitch={() => {}}
      onRename={() => {}}
      onDelete={() => {}}
      draft={over.draft ?? ""}
      onDraftChange={() => {}}
      modelPicker={null}
      onOpenSettings={() => {}}
      hasProvider
      attach={attach}
    />
  );
  return { ...utils, onSend, attach };
}

beforeEach(() => runtimeMock.reset());
afterEach(() => cleanup());

describe("ChatPanel attachments", () => {
  it("has a clip button that opens the picker", () => {
    const { attach } = renderPanel();
    fireEvent.click(screen.getByRole("button", { name: "Lampirkan berkas" }));
    expect(attach.pick).toHaveBeenCalled();
  });

  it("disables the clip button while busy", () => {
    renderPanel({ busy: true });
    expect((screen.getByRole("button", { name: "Lampirkan berkas" }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("shows pending items as thumbnail or chip with a remove button", () => {
    const attach = api({ pending: [meta("a"), meta("b", "pdf")] });
    renderPanel({ attach });
    const img = screen.getByRole("img", { name: "a.png" }) as HTMLImageElement;
    expect(img.getAttribute("src")).toBe("/attachments/a");
    expect(screen.getByText("b.pdf")).toBeTruthy();
    expect(screen.getByText("2.0 KB")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Hapus lampiran a.png" }));
    expect(attach.remove).toHaveBeenCalledWith("a");
  });

  it("sends with attachments only, then resets draft and pending", () => {
    const attach = api({ pending: [meta("a")] });
    const { onSend } = renderPanel({ attach });
    const send = screen.getByRole("button", { name: "Kirim pesan" }) as HTMLButtonElement;
    expect(send.disabled).toBe(false);
    fireEvent.click(send);
    expect(onSend).toHaveBeenCalledWith("", attach.pending);
    expect(attach.reset).toHaveBeenCalled();
  });

  it("keeps send disabled with neither text nor attachments", () => {
    renderPanel();
    expect((screen.getByRole("button", { name: "Kirim pesan" }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("stages pasted images instead of inserting them", () => {
    const { attach } = renderPanel();
    const file = new File([new Uint8Array([1])], "shot.png", { type: "image/png" });
    fireEvent.paste(screen.getByRole("textbox", { name: "Tulis pesan" }), {
      clipboardData: { files: [file], getData: () => "" },
    });
    expect(attach.stageFiles).toHaveBeenCalledWith([file]);
  });

  it("leaves a plain text paste alone", () => {
    const { attach } = renderPanel();
    fireEvent.paste(screen.getByRole("textbox", { name: "Tulis pesan" }), {
      clipboardData: { files: [], getData: () => "halo" },
    });
    expect(attach.stageFiles).not.toHaveBeenCalled();
  });

  it("listens for dropped files only while mounted", () => {
    const { attach, unmount } = renderPanel();
    expect(runtimeMock.fileDrop).not.toBeNull();
    runtimeMock.fileDrop?.(0, 0, ["C:\\x\\a.png"]);
    expect(attach.stagePaths).toHaveBeenCalledWith(["C:\\x\\a.png"]);
    unmount();
    expect(runtimeMock.fileDrop).toBeNull();
  });

  it("shows the attachment error with a close button", () => {
    const attach = api({ error: "foto.exe: tipe berkas tidak didukung" });
    renderPanel({ attach });
    expect(screen.getByRole("alert").textContent).toContain("foto.exe");
    fireEvent.click(screen.getByRole("button", { name: "Tutup pesan lampiran" }));
    expect(attach.clearError).toHaveBeenCalled();
  });
});
