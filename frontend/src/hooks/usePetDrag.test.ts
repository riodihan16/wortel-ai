import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import type { MouseEvent, PointerEvent } from "react";
import { usePetDrag } from "./usePetDrag";

// Minimal synthetic events: the hook only reads button/buttons/screen
// coordinates and click detail, so a plain object is enough.
const capture = vi.fn();

function ptr(screenX: number, screenY: number, buttons = 1, button = 0) {
  return {
    screenX, screenY, buttons, button, pointerId: 1,
    currentTarget: { setPointerCapture: capture },
  } as unknown as PointerEvent<HTMLElement>;
}

function click(detail = 1) {
  return { detail } as unknown as MouseEvent<HTMLElement>;
}

function setup() {
  const onClick = vi.fn();
  const startDrag = vi.fn();
  const { result } = renderHook(() => usePetDrag(onClick, startDrag));
  return { h: result.current, onClick, startDrag };
}

describe("usePetDrag", () => {
  it("treats press and release without movement as a click", () => {
    const { h, onClick, startDrag } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerUp();
    h.onClick(click());
    expect(startDrag).not.toHaveBeenCalled();
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("ignores jitter below the threshold so the click still opens chat", () => {
    const { h, onClick, startDrag } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerMove(ptr(102, 101));
    h.onPointerUp();
    h.onClick(click());
    expect(startDrag).not.toHaveBeenCalled();
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("starts a window drag once movement passes the threshold, only once", () => {
    const { h, startDrag } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerMove(ptr(110, 100));
    h.onPointerMove(ptr(130, 100));
    expect(startDrag).toHaveBeenCalledTimes(1);
  });

  it("suppresses the click that follows a drag", () => {
    const { h, onClick } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerMove(ptr(100, 120));
    h.onPointerUp();
    h.onClick(click());
    expect(onClick).not.toHaveBeenCalled();
  });

  it("clicks normally on the next press after a drag", () => {
    const { h, onClick } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerMove(ptr(100, 120));
    h.onPointerDown(ptr(50, 50));
    h.onPointerUp();
    h.onClick(click());
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("never drags on moves without the primary button held", () => {
    const { h, startDrag } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerMove(ptr(140, 100, 0));
    h.onPointerMove(ptr(160, 100, 0));
    expect(startDrag).not.toHaveBeenCalled();
  });

  it("ignores non-primary button presses", () => {
    const { h, startDrag } = setup();
    h.onPointerDown(ptr(100, 100, 2, 2));
    h.onPointerMove(ptr(140, 100, 2));
    expect(startDrag).not.toHaveBeenCalled();
  });

  it("captures the pointer on press so release outside W still ends the gesture", () => {
    const { h } = setup();
    capture.mockClear();
    h.onPointerDown(ptr(100, 100));
    expect(capture).toHaveBeenCalledWith(1);
  });

  it("forgets the press on pointercancel so a later move cannot drag from a stale origin", () => {
    const { h, startDrag } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerCancel();
    h.onPointerMove(ptr(160, 100));
    expect(startDrag).not.toHaveBeenCalled();
  });

  it("keeps keyboard activation working even right after a drag", () => {
    const { h, onClick } = setup();
    h.onPointerDown(ptr(100, 100));
    h.onPointerMove(ptr(100, 120));
    h.onClick(click(0));
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
