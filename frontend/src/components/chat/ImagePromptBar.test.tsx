import { describe, expect, it, vi, afterEach } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ComposerPlusMenu } from "./ImagePromptBar";

function setup(over: Partial<Parameters<typeof ComposerPlusMenu>[0]> = {}) {
  const onUpload = vi.fn();
  const onToggleImage = vi.fn();
  const props = {
    attachDisabled: false,
    imageActive: false,
    hasImageModel: true,
    onUpload,
    onToggleImage,
    ...over,
  };
  const utils = render(<ComposerPlusMenu {...props} />);
  return { ...utils, onUpload, onToggleImage };
}

afterEach(() => cleanup());

describe("ComposerPlusMenu", () => {
  it("opens a menu with Upload file and Buat gambar", () => {
    setup();
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    expect(screen.getByRole("menuitem", { name: /Upload file/ })).toBeTruthy();
    expect(screen.getByRole("menuitem", { name: /Buat gambar/ })).toBeTruthy();
  });

  it("calls onUpload from the Upload file item", () => {
    const { onUpload } = setup();
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    fireEvent.click(screen.getByRole("menuitem", { name: /Upload file/ }));
    expect(onUpload).toHaveBeenCalled();
  });

  it("calls onToggleImage from the Buat gambar item", () => {
    const { onToggleImage } = setup();
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    fireEvent.click(screen.getByRole("menuitem", { name: /Buat gambar/ }));
    expect(onToggleImage).toHaveBeenCalled();
  });

  it("marks image mode active with a check", () => {
    setup({ imageActive: true });
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    const item = screen.getByRole("menuitem", { name: /Buat gambar/ });
    expect(item.className).toContain("menu-option--on");
  });

  it("disables Upload file when attachment picking is unavailable", () => {
    setup({ attachDisabled: true });
    fireEvent.click(screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" }));
    expect((screen.getByRole("menuitem", { name: /Upload file/ }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("closes on Escape and returns focus to the trigger", () => {
    setup();
    const trigger = screen.getByRole("button", { name: "Tambah lampiran atau buat gambar" });
    fireEvent.click(trigger);
    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("menuitem", { name: /Upload file/ })).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });
});
