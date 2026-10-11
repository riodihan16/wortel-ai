---
last_mapped_commit: b55c8e5bb1c4199b24162ca36f0c612ef255a01c
last_mapped_at: 2026-10-11
---
# Testing Patterns

**Analysis Date:** 2026-10-11

## Test Framework

**Backend (Go):**
- Standard library `testing` only. No third-party assertion library, no testify.
- HTTP tests use `net/http/httptest` (`httptest.NewServer`, `httptest.NewRecorder`, `httptest.NewRequest`).
- Run commands:

```bash
go test -count=1 ./...          # all backend tests (CI uses this)
go test -race -count=1 ./...    # race detector (CI, CGO_ENABLED=1)
go test ./internal/ai/...       # one package
```

**Frontend (TypeScript/React):**
- Runner: Vitest 3.x
- Assertions: Vitest's built-in `expect`
- Component/hook testing: `@testing-library/react` 16 (`render`, `renderHook`, `screen`, `fireEvent`, `act`, `waitFor`) with `@testing-library/dom`
- Environment: `jsdom`
- Config: `frontend/vitest.config.ts`

```ts
export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.{ts,tsx}"],
    restoreMocks: true,
  },
});
```

- Run commands (from `frontend/`):

```bash
npm test              # vitest run (single pass)
npm run test:watch    # vitest (watch)
npx tsc --noEmit      # typecheck gate (CI runs this before tests)
```

## Test File Organization

**Location:**
- Go: co-located in the same package/directory as the subject, always `<subject>_test.go`.
- TypeScript: co-located next to the subject (`useChat.ts` + `useChat.test.ts`). Shared test infrastructure lives in `frontend/src/test/`.

**Naming:**
- Go: `TestSubjectBehavior` in PascalCase, sometimes compound (`TestPrepareRetryReusesUserAndDropsFailedReply`, `TestError400NamingImagesIsUnsupportedInput`).
- TypeScript: `describe("<subject>", ...)` + `it("<expected behavior>", ...)` with behavior-phrased names, e.g. `it("releases the busy guard when CancelGeneration rejects", ...)`.
- Component integration tests add a scenario suffix: `ChatPanel.attach.test.tsx`, `MessageItem.attach.test.tsx`.

**Structure:**

```
internal/
├── ai/                provider_test.go, sse_test.go, http_test.go, parts_test.go
├── app/               service_test.go, providers_test.go, attachments_test.go,
│                      attachment_handler_test.go, context_test.go, shortcut_test.go,
│                      appearance_test.go
├── attach/            validate_test.go, staging_test.go
├── storage/           db_test.go, chat_test.go, providers_test.go,
│                      attachments_test.go, testhelper_test.go
├── hotkey/            combo_test.go, hotkey_windows_test.go, rebind_windows_test.go
├── credentials/       mapstore_test.go
├── window/            place_test.go
├── tray/              tray_test.go
└── autostart/         autostart_windows_test.go, export_test_windows.go
frontend/src/
├── hooks/             useChat.test.ts, useAttachments.test.ts, useProviders.test.ts,
│                      useCopy.test.ts, usePetDrag.test.ts, usePetSignal.test.ts
├── lib/               formatSize.test.ts, markdownToText.test.ts
├── components/chat/   SafeMarkdown.test.tsx, ChatPanel.attach.test.tsx,
│                      MessageItem.attach.test.tsx
└── test/              wails.ts (shared Wails runtime mock)
```

## Platform-Build-Tag Tests

Windows-specific logic is tested behind build tags. Files start with `//go:build windows` and exercise real OS behavior:
- `fatal_windows_test.go` (root) tests `isWebView2Missing` message matching.
- `internal/autostart/autostart_windows_test.go` touches the real `HKCU\Run` key and restores prior state in a `defer`, so it is safe on a dev machine:

```go
func TestRoundTrip(t *testing.T) {
	priorEnabled := Enabled()
	priorValue := readRaw()
	defer func() {
		if priorEnabled {
			exe, _ := os.Executable()
			_ = Enable(exe)
		} else {
			_ = Disable()
		}
		_ = priorValue
	}()
	// ... disable, enable, assert, disable
}
```

- `internal/autostart/export_test_windows.go` exports internals (`readRaw`) to the test package via a build-tagged file rather than changing the production API.

## Go Test Structure

**Table-driven tests** are the dominant pattern. Use an anonymous struct slice with a `name` field and a `for _, tc := range` loop:

```go
func TestErrorMapping429500408400(t *testing.T) {
	for _, tc := range []struct {
		status int
		code   string
		retry  bool
	}{
		{429, "rate_limited", true},
		{500, "provider_server", true},
		{408, "timeout", true},
		{400, "bad_request", false},
	} {
		srv := sseServer(t, tc.status, `{"error":"x"}`, nil)
		pub, ok := streamErr(t, &OpenAICompatible{BaseURL: srv.URL}).(*PublicError)
		srv.Close()
		if !ok || pub.Code != tc.code || pub.Retryable != tc.retry || pub.StatusCode != tc.status {
			t.Fatalf("status %d: got %#v", tc.status, pub)
		}
	}
}
```

Map-based tables are also used when order is irrelevant (e.g. `fatal_windows_test.go`, `internal/attach/validate_test.go`).

**Assertions:** Use `t.Fatalf` for hard failures, `t.Errorf` when continuing is useful (see `toggle_test.go`). No assertion helper library — conditionals with formatted messages. Common message forms: `t.Fatalf("got %q", got)`, `t.Fatalf("%s: got %v", name, err)`.

**Helpers:** Test helpers take `*testing.T` as the first parameter and call `t.Helper()`:

```go
func sseServer(t *testing.T, status int, body string, headers map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { /* ... */ }))
}
```

Helpers live in the same test file or a shared `_test.go` file (`internal/storage/testhelper_test.go`).

**Cleanup:** Use `t.TempDir()` for filesystem isolation and `t.Cleanup(...)` / `defer` for teardown. `t.Cleanup` is preferred in helpers:

```go
func testProviders(t *testing.T) *Providers {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "p.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewProviders(db, credentials.NewMapStore())
}
```

## Mocking

**Go — hand-written test doubles, no mocking library.**
- `ai.MockProvider{Delay: 20 * time.Millisecond}` is the built-in fake provider used across `internal/app/*_test.go`.
- Interfaces get small in-memory implementations: `credentials.MapStore` doubles the real `credentials.Store`.
- HTTP is faked with `httptest.NewServer` handlers that emit SSE frames: `sseOK(w)` writes a minimal `data:` stream plus `[DONE]` (`internal/app/attachments_test.go`).
- Closure-based injection: build a `Service` with a custom emit function to capture events, e.g. `NewService(&ai.MockProvider{}, sink.emit)`.

**Go — concurrency-safe sinks.** When the service emits from a goroutine, test doubles guard state with a mutex so `-race` stays clean:

```go
type eventSink struct {
	mu     sync.Mutex
	events []ai.ChatEvent
}
func (s *eventSink) emit(_ string, e ai.ChatEvent) {
	s.mu.Lock()
	s.events = append(s.events, e)
	s.mu.Unlock()
}
func (s *eventSink) snapshot() []ai.ChatEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ai.ChatEvent(nil), s.events...)
}
```

**Frontend — `vi.mock` for Wails bindings.**
- Bindings from `../../wailsjs/go/main/App` are mocked with a plain object of `vi.fn()`s declared *before* the import of the system under test, because `vi.mock` is hoisted. Response values live in that object so each test can steer the backend without re-mocking:

```ts
const backend = {
  history: vi.fn(async () => [] as unknown[]),
  send: vi.fn(async () => "req-1"),
  cancel: vi.fn(async () => {}),
};
vi.mock("../../wailsjs/go/main/App", () => ({
  History: (...a: unknown[]) => backend.history(...(a as [])),
  SendMessage: (...a: unknown[]) => backend.send(...(a as [])),
  CancelGeneration: (...a: unknown[]) => backend.cancel(...(a as [])),
  // ...
}));
import { useChat } from "../hooks/useChat"; // imported AFTER the mock
```

**Frontend — shared Wails runtime mock.** `frontend/src/test/wails.ts` emulates the `window.runtime.*` subscribe/emit contract that jsdom lacks, and records the file-drop handler so tests can fire a drop:

```ts
vi.mock("../../wailsjs/runtime/runtime", () => ({
  EventsOn: (event: string, cb: Handler) => runtimeMock.on(event, cb),
  EventsOff: (event: string) => { listeners.delete(event); },
  EventsEmit: () => {},
  OnFileDrop: (cb: DropHandler) => { runtimeMock.fileDrop = cb; },
  OnFileDropOff: () => { runtimeMock.fileDrop = null; },
}));
```

Tests call `runtimeMock.reset()` in `beforeEach` and `runtimeMock.emit("chat:event", payload)` to drive event handlers.

**What to Mock:**
- All Wails bindings (`wailsjs/go/main/App`) — no real Go backend runs in jsdom.
- The Wails runtime event bus (`wailsjs/runtime/runtime`) via the shared `runtimeMock`.
- `navigator.clipboard` — set/cleared with `Object.defineProperty` (`useCopy.test.ts`).

**What NOT to Mock:**
- Go business logic: services are exercised against real SQLite in a `t.TempDir()` and real `httptest` servers. Only the network *endpoint* is faked, never the code under test.
- Frontend utilities (`formatSize`, `markdownToText`) are tested directly with real inputs.

## Fixtures and Factories

**Go — inline builders and sample data.**
- Attachment bytes: `pngBytes` var and `pngB64()` helper (`internal/app/attachments_test.go`).
- Magic-byte headers as package vars: `pngHead`, `jpegHead`, `webpHead`, `pdfHead` (`internal/attach/validate_test.go`).
- Deterministic IDs: `Staging.newID` is overridable so tests pin IDs:

```go
func stagingWithIDs(ids ...string) *Staging {
	s := NewStaging()
	i := 0
	s.newID = func() string { id := ids[i]; i++; return id }
	return s
}
```

**TypeScript — small factory functions.** Tests define a local `meta(...)` / `profile(...)` / `user(...)` builder and default it via `Partial<T>` overrides:

```ts
function user(over: Partial<ChatMessage> = {}): ChatMessage {
  return { id: "u1", role: "user", content: "lihat", status: "complete", ...over };
}
```

Component props are provided by a `renderPanel(over = {})` helper that fills every required prop with a no-op (`ChatPanel.attach.test.tsx`).

**Location:** Fixtures are inline in the test file; no shared fixture directory. The one shared frontend artifact is `frontend/src/test/wails.ts`.

## Coverage

**Requirements:** No coverage threshold is enforced. CI gates on: `gofmt` clean, `go vet`, `go test`, `go test -race` (backend) and `tsc --noEmit`, `npm test`, `npm run build` (frontend). See `.github/workflows/ci.yml`.

**View Coverage:**

```bash
go test -cover ./...            # backend, per package
cd frontend && npx vitest run --coverage   # frontend (no provider configured)
```

## Test Types

**Unit Tests:**
- Pure functions and parsers dominate: `NormalizeBaseURL`, `MapHTTPError`, `ScanSSE` (`internal/ai/sse_test.go`), `Classify`/`CleanName` (`internal/attach/validate_test.go`), hotkey `Parse`/`String` round-trips (`internal/hotkey/combo_test.go`), `Fit`/`Constrain`/`Placer` window math (`internal/window/place_test.go`), `toggleFor` (`toggle_test.go`).
- Frontend pure helpers: `formatSize`, `markdownToText`.

**Integration Tests:**
- Backend: `internal/app/*_test.go` wire a real `Service` + temp SQLite + `httptest` SSE server and assert end-to-end streaming, persistence, attachments, retry, and cancel. `internal/storage/*_test.go` exercise real SQLite including migrations.
- Frontend: hooks and components run in jsdom against mocked bindings but real React state/effects (`useChat.test.ts`, `ChatPanel.attach.test.tsx`).

**E2E Tests:** Not used. No Playwright/Cypress. Wails runtime behavior is covered by the hand-written `runtimeMock` in `frontend/src/test/wails.ts`.

## Common Patterns

**Async Go testing — poll for terminal events.** Because generation runs in a goroutine, tests wait on a deadline rather than sleeping blindly:

```go
func waitDone(t *testing.T, svc *Service, sink *eventSink, n int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c := 0
		for _, e := range sink.snapshot() {
			if e.Type == "completed" || e.Type == "cancelled" || e.Type == "failed" {
				c++
			}
		}
		svc.mu.Lock()
		active := len(svc.active)
		svc.mu.Unlock()
		if c >= n && active == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no terminal event or generation still active: %+v", sink.snapshot())
}
```

**Cancellation testing:** `context.WithTimeout` / `context.WithCancel` with `defer cancel()`, asserting `errors.Is(err, context.DeadlineExceeded)` or `context.Canceled` (`internal/ai/provider_test.go`, `internal/ai/sse_test.go`).

**Async TypeScript testing — wrap state updates in `act`.** Hook tests await `act(async () => ...)` for sends and `waitFor` for initial loads:

```ts
const r = renderHook(() => useChat());
await waitFor(() => expect(r.result.current.conversationId).not.toBe(""));
await act(async () => { await result.current.send("halo"); });
```

**React component testing — query by accessible role/name.** Never query by class or test id. Use `screen.getByRole("button", { name: "Kirim pesan" })`, `getByRole("alert")`, `getByRole("img", { name: "foto.png" })`. Interaction via `fireEvent.click` / `fireEvent.paste` / `fireEvent.dragEnter`.

**Fake timers:** `vi.useFakeTimers()` + `vi.advanceTimersByTime(ms)` with `afterEach(() => vi.useRealTimers())` (`useCopy.test.ts`).

**Synthetic event objects:** When a hook only reads a few fields, cast a plain object: `function ptr(...) { return { screenX, screenY, buttons, button, pointerId: 1, currentTarget: { setPointerCapture } } as unknown as PointerEvent<HTMLElement>; }` (`usePetDrag.test.ts`).

**Error-path testing:** Every module tests its failure branch explicitly — `TestOpenRejectsEmptyPath`, `TestMapStoreRejectsEmpty`, `TestClassifyRejectsOversize`, `TestSaveRequiresKeyForNew`, `TestParseRejects`, and frontend `it("releases the busy guard when CancelGeneration rejects", ...)`. Include a rejection case for every guard.

**Test doubles restore state:** Mocks are reset in `beforeEach` (`runtimeMock.reset()`, `vi.clearAllMocks()`) and `restoreMocks: true` is set globally in `vitest.config.ts`.

---

*Testing analysis: 2026-10-11*
