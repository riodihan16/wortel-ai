import { describe, expect, it, vi, afterEach } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ImagePromptBar } from "./ImagePromptBar";

function setup(over: Partial<Parameters<typeof ImagePromptBar>[0]> = {}) {
  const onGenerate = vi.fn();
  const onCancel = vi.fn();
  const onDismissError = vi.fn();
  const onOpenSettings = vi.fn();
  const props = {
    pending: false,
    error: null as string | null,
    imageModel: "img-model" as string | null,
    chatBusy: false,
    onGenerate,
    onCancel,
    onDismissError,
    onOpenSettings,
    ...over,
  };
  const utils = render(<ImagePromptBar {...props} />);
  return { ...utils, onGenerate, onCancel, onDismissError, onOpenSettings };
}

afterEach(() => cleanup());

describe("ImagePromptBar", () => {
  it("disables generate without an image model and points to Settings", () => {
    setup({ imageModel: null });
    const btn = screen.getByRole("button", { name: "Buat gambar" });
    expect((btn as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByText(/Belum ada model gambar/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "Atur di Pengaturan" })).toBeTruthy();
  });

  it("generates on Enter when a model is set", () => {
    const { onGenerate } = setup();
    const ta = screen.getByLabelText("Prompt gambar");
    fireEvent.change(ta, { target: { value: "a red panda" } });
    fireEvent.keyDown(ta, { key: "Enter" });
    expect(onGenerate).toHaveBeenCalledWith("a red panda");
  });

  it("shows a distinct drawing state with cancel while pending", () => {
    const { onCancel } = setup({ pending: true });
    expect(screen.getByText("Wortel sedang menggambar…")).toBeTruthy();
    const stop = screen.getByRole("button", { name: "Hentikan pembuatan gambar" });
    fireEvent.click(stop);
    expect(onCancel).toHaveBeenCalled();
  });

  it("cancels on Escape while pending", () => {
    const { onCancel } = setup({ pending: true });
    fireEvent.keyDown(screen.getByLabelText("Prompt gambar"), { key: "Escape" });
    expect(onCancel).toHaveBeenCalled();
  });

  it("shows the PAYG billing note when a model is configured", () => {
    setup();
    expect(screen.getByText(/saldo PAYG, bukan paket chat/)).toBeTruthy();
  });

  it("renders and dismisses an error", () => {
    const { onDismissError } = setup({ error: "Saldo PAYG tidak cukup." });
    expect(screen.getByText("Saldo PAYG tidak cukup.")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Tutup pesan gambar" }));
    expect(onDismissError).toHaveBeenCalled();
  });
});
