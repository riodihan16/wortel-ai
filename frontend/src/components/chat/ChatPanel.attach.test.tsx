import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { runtimeMock } from "../../test/wails";
import type { AttachmentMeta } from "../../types/chat";
import type { AttachmentsApi } from "../../hooks/useAttachments";
import type { ImageGenApi } from "../../hooks/useImageGen";
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

function imgGen(over: Partial<ImageGenApi> = {}): ImageGenApi {
  return {
    pending: false,
    status: "idle",
    error: null,
    generate: vi.fn(async () => {}),
    cancel: vi.fn(),
    clearError: vi.fn(),
    ...over,
  };
}

function renderPanel(
  over: {
    attach?: AttachmentsApi;
    draft?: string;
    busy?: boolean;
    imageModel?: string | null;
    imageGen?: ImageGenApi;
    onDraftChange?: (v: string) => void;
  } = {}
) {
  const onSend = vi.fn();
  const attach = over.attach ?? api();
  const imageGen = over.imageGen ?? imgGen();
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
      onDraftChange={over.onDraftChange ?? (() => {})}
      modelPicker={null}
      onOpenSettings={() => {}}
      hasProvider
      attach={attach}
      imageGen={imageGen}
      imageModel={over.imageModel ?? "img-model"}
    />
  );
  return { ...utils, onSend, attach, imageGen };
}

beforeEach(() => runtimeMock.reset());
afterEach(() => cleanup());

describe("ChatPanel attachments", () => {
  it("opens the picker from the + menu Upload file item", () => {
    const { attach } = renderPanel();
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    fireEvent.click(screen.getByRole("menuitem", { name: /Upload file/ }));
    expect(attach.pick).toHaveBeenCalled();
  });

  it("disables Upload file while busy", () => {
    renderPanel({ busy: true });
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    expect((screen.getByRole("menuitem", { name: /Upload file/ }) as HTMLButtonElement).disabled).toBe(true);
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

  it("stages a pasted screenshot (image, no text)", () => {
    const { attach } = renderPanel();
    const file = new File([new Uint8Array([1])], "shot.png", { type: "image/png" });
    const notPrevented = fireEvent.paste(screen.getByRole("textbox", { name: "Tulis pesan" }), {
      clipboardData: { files: [file], getData: () => "" },
    });
    expect(attach.stageFiles).toHaveBeenCalledWith([file]);
    expect(notPrevented).toBe(false);
  });

  it("keeps a paste that carries text next to an image as text", () => {
    // Excel and OneNote put a bitmap of the selection beside the text.
    const { attach } = renderPanel();
    const file = new File([new Uint8Array([1])], "image.png", { type: "image/png" });
    const notPrevented = fireEvent.paste(screen.getByRole("textbox", { name: "Tulis pesan" }), {
      clipboardData: { files: [file], getData: (t: string) => (t === "text/plain" ? "a\tb" : "") },
    });
    expect(attach.stageFiles).not.toHaveBeenCalled();
    expect(notPrevented).toBe(true);
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

  it("shows the drop overlay for a file drag and hides it on a window drop", () => {
    const { container } = renderPanel();
    fireEvent.dragEnter(container.querySelector("#chat-main")!, { dataTransfer: { types: ["Files"] } });
    expect(container.querySelector("#chat-drop")).not.toBeNull();
    fireEvent.drop(window, { dataTransfer: { types: ["Files"] } });
    expect(container.querySelector("#chat-drop")).toBeNull();
  });

  it("ignores a text drag", () => {
    const { container } = renderPanel();
    fireEvent.dragEnter(container.querySelector("#chat-main")!, { dataTransfer: { types: ["text/plain"] } });
    expect(container.querySelector("#chat-drop")).toBeNull();
  });
});

describe("ChatPanel create-image mode (single composer)", () => {
  const enterImageMode = () => {
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    fireEvent.click(screen.getByRole("menuitem", { name: /Buat gambar/ }));
  };

  it("sends the draft to image generation on Enter after enabling the mode", () => {
    const imageGen = imgGen();
    const { onSend } = renderPanel({ draft: "a red panda", imageGen });
    enterImageMode();
    fireEvent.keyDown(screen.getByLabelText("Prompt gambar"), { key: "Enter" });
    expect(imageGen.generate).toHaveBeenCalledWith("c1", "a red panda", "img-model");
    expect(onSend).not.toHaveBeenCalled();
  });

  it("clears the draft after an image prompt is sent", () => {
    const onDraftChange = vi.fn();
    const imageGen = imgGen();
    renderPanel({ draft: "a cat", imageGen, onDraftChange });
    enterImageMode();
    fireEvent.keyDown(screen.getByLabelText("Prompt gambar"), { key: "Enter" });
    expect(onDraftChange).toHaveBeenCalledWith("");
  });

  it("cancels a pending image generation on Escape and via the stop button", () => {
    const imageGen = imgGen({ pending: true });
    renderPanel({ imageGen });
    enterImageMode();
    fireEvent.keyDown(screen.getByLabelText("Prompt gambar"), { key: "Escape" });
    expect(imageGen.cancel).toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Hentikan pembuatan gambar" }));
    expect(imageGen.cancel).toHaveBeenCalledTimes(2);
  });

  it("disables Upload file in image mode", () => {
    renderPanel();
    enterImageMode();
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    expect((screen.getByRole("menuitem", { name: /Upload file/ }) as HTMLButtonElement).disabled).toBe(true);
  });
});
