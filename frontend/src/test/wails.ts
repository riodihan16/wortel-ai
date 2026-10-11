import { vi } from "vitest";

// Shared test double for the Wails runtime. The real module calls
// window.runtime.*, which jsdom does not provide; this emulates the same
// subscribe/emit contract so hooks receive synthetic backend events, and
// records the file-drop listener so a test can fire a drop.
type Handler = (...args: unknown[]) => void;
type DropHandler = (x: number, y: number, paths: string[]) => void;

const listeners = new Map<string, Set<Handler>>();

export const runtimeMock = {
  fileDrop: null as DropHandler | null,
  on(event: string, cb: Handler) {
    let set = listeners.get(event);
    if (!set) {
      set = new Set();
      listeners.set(event, set);
    }
    set.add(cb);
    return () => set!.delete(cb);
  },
  emit(event: string, ...args: unknown[]) {
    listeners.get(event)?.forEach((cb) => cb(...args));
  },
  reset() {
    listeners.clear();
    runtimeMock.fileDrop = null;
  },
};

vi.mock("../../wailsjs/runtime/runtime", () => ({
  EventsOn: (event: string, cb: Handler) => runtimeMock.on(event, cb),
  EventsOff: (event: string) => {
    listeners.delete(event);
  },
  EventsEmit: () => {},
  OnFileDrop: (cb: DropHandler) => {
    runtimeMock.fileDrop = cb;
  },
  OnFileDropOff: () => {
    runtimeMock.fileDrop = null;
  },
}));
