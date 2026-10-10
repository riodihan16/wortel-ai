//go:build windows

package hotkey

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRegisterAndStop(t *testing.T) {
	var hits int32
	if err := Start(Hotkey{Mods: ModControl | ModShift | ModNoRepeat, VK: 0x57}, func() {
		atomic.AddInt32(&hits, 1)
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	Stop()
	time.Sleep(50 * time.Millisecond)
	// Second Stop must be safe (idempotent).
	Stop()
}
