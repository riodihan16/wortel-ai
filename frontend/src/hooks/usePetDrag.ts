import {useCallback, useRef} from 'react';
import type {MouseEvent, PointerEvent} from 'react';

declare global {
    interface Window {
        // Injected by the Wails v2 runtime; the same IPC message its own
        // --wails-draggable handler sends to start a native window move.
        WailsInvoke?: (message: string) => void;
    }
}

// Movement below this is click jitter, not a drag. Without it a 1px wobble
// between press and release would start a window move and eat the click
// (the double-click bug from ADR-029).
const DRAG_THRESHOLD_PX = 4;

export function startWindowDrag(): void {
    window.WailsInvoke?.('drag');
}

export interface PetDragHandlers {
    onPointerDown: (e: PointerEvent<HTMLElement>) => void;
    onPointerMove: (e: PointerEvent<HTMLElement>) => void;
    onPointerUp: () => void;
    onPointerCancel: () => void;
    onClick: (e: MouseEvent<HTMLElement>) => void;
}

// One element that both opens chat on click and moves the window on drag.
// The element must opt out of --wails-draggable, otherwise the runtime starts
// the move on the first mousemove and the threshold never applies.
export function usePetDrag(onClick: () => void, startDrag: () => void = startWindowDrag): PetDragHandlers {
    const origin = useRef<{x: number; y: number} | null>(null);
    const dragged = useRef(false);

    const onPointerDown = useCallback((e: PointerEvent<HTMLElement>) => {
        dragged.current = false;
        if (e.button !== 0) {
            origin.current = null;
            return;
        }
        origin.current = {x: e.screenX, y: e.screenY};
        // Capture so a release outside W still reaches onPointerUp; otherwise
        // origin would linger and a later held move over W would drag.
        e.currentTarget.setPointerCapture(e.pointerId);
    }, []);

    const onPointerMove = useCallback((e: PointerEvent<HTMLElement>) => {
        const o = origin.current;
        if (!o || (e.buttons & 1) === 0) return;
        if (Math.hypot(e.screenX - o.x, e.screenY - o.y) < DRAG_THRESHOLD_PX) return;
        origin.current = null;
        dragged.current = true;
        startDrag();
    }, [startDrag]);

    const onPointerUp = useCallback(() => {
        origin.current = null;
    }, []);

    const handleClick = useCallback((e: MouseEvent<HTMLElement>) => {
        // detail 0 = keyboard activation, which never follows a drag.
        if (dragged.current && e.detail !== 0) {
            dragged.current = false;
            return;
        }
        onClick();
    }, [onClick]);

    return {onPointerDown, onPointerMove, onPointerUp, onPointerCancel: onPointerUp, onClick: handleClick};
}
