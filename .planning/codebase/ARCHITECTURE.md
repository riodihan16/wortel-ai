---
last_mapped_commit: b55c8e5bb1c4199b24162ca36f0c612ef255a01c
last_mapped_at: 2026-10-11
---
<!-- refreshed: 2026-10-11 -->

# Architecture

**Analysis Date:** 2026-10-11

## System Overview

Wortel AI is a Wails v2 desktop companion for Windows: a single Go binary that embeds a React WebView frontend and calls OpenAI-compatible chat endpoints. There is no localhost server, no telemetry, and no local model. The Go backend owns networking, secrets, storage, and Windows integration; the React frontend only calls Wails bindings and consumes Wails events.

```text
┌─────────────────────────────────────────────────────────────┐
│                    React Frontend (WebView2)                 │
│  `frontend/src/App.tsx` · `frontend/src/hooks/*`             │
│  pet / chat / settings — one window, one view at a time      │
└───────────────────────┬─────────────────────────────────────┘
                        │ Wails bridge (bindings + events)
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              Wails App Facade — `app.go`                     │
│  ~35 bound methods; lifecycle; window/tray/hotkey wiring     │
└───────────────────────┬─────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┬───────────────┐
        ▼               ▼               ▼               ▼
┌──────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────────────┐
│ app.Service  │ │app.Providers│ │  tray /    │ │  autostart /     │
│ `internal/   │ │ `internal/  │ │ hotkey /   │ │  window          │
│  app/        │ │  app/       │ │ window     │ │  `internal/...`  │
│  service.go` │ │  providers  │ │ (Win32)    │ │                  │
└──────┬───────┘ └──────┬─────┘ └────────────┘ └──────────────────┘
       │                │
       ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  ai.Client   │ │  storage.DB  │ │ credentials  │
│ `internal/ai`│ │ `internal/   │ │ .OSStore     │
│ HTTP + SSE   │ │  storage`    │ │ `internal/   │
│              │ │ SQLite (WAL) │ │  credentials`│
└──────┬───────┘ └──────────────┘ └──────┬───────┘
       │ HTTPS + SSE                     │
       ▼                                 ▼
┌────────────────────┐        ┌──────────────────────────┐
│ User-chosen AI     │        │ Windows Credential       │
│ provider endpoint  │        │ Manager (API keys)       │
└────────────────────┘        └──────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Wails entry | Configures window (frameless, transparent, always-on-top), embeds `frontend/dist`, binds `App` | `main.go` |
| App facade | Bound methods for the frontend; lifecycle hooks; owns window mode/tray/hotkey state | `app.go` |
| Generation service | Orchestrates one stream per conversation; cancel map; persistence; emits `chat:event` | `internal/app/service.go` |
| Provider manager | Profile CRUD, credential storage, active-provider resolution, connection test | `internal/app/providers.go` |
| Context builder | Builds provider request from stored rows; applies attachment window/budget | `internal/app/context.go` |
| History | Frontend-safe conversation/message reads; rename/delete; busy guard | `internal/app/history.go` |
| Attachments | Stage dialog/drop/paste files; classify; asset HTTP handler | `internal/app/attachments.go`, `internal/app/attachment_handler.go` |
| Appearance/Shortcut | Persist theme, font scale, global shortcut keys in `app_settings` | `internal/app/appearance.go`, `internal/app/shortcut.go` |
| AI client | URL normalization, HTTP request, SSE parse, error mapping | `internal/ai/provider.go`, `internal/ai/sse.go`, `internal/ai/errors.go`, `internal/ai/http.go` |
| AI types | `ChatEvent`, `PublicError`, `Message`, multimodal `Part` | `internal/ai/types.go`, `internal/ai/parts.go` |
| Config validation | Provider input validation, model normalization | `internal/config/providers.go` |
| Credential store | `Store` interface; `OSStore` (go-keyring) and `mapstore` (tests) | `internal/credentials/` |
| Storage | SQLite open/migrate; conversations/messages/providers/settings repositories | `internal/storage/` |
| Attachment validation | Kind/mime sniffing, size limits, safe names | `internal/attach/validate.go`, `internal/attach/staging.go` |
| Window placement | Pure sizing math + Win32 glue (drag guard, foreground) | `internal/window/place.go`, `internal/window/place_windows.go`, `internal/window/activate_windows.go` |
| Tray | Win32 tray icon/menu; intent mapping | `internal/tray/tray.go`, `internal/tray/tray_windows.go` |
| Global hotkey | Parse combo, register `WM_HOTKEY` | `internal/hotkey/hotkey.go`, `internal/hotkey/hotkey_windows.go` |
| Autostart | Per-user `Run` registry key | `internal/autostart/autostart.go`, `internal/autostart/autostart_windows.go` |

## Pattern Overview

**Overall:** Layered facade + service, platform-tagged adapters, event-driven streaming.

**Key Characteristics:**
- The Wails facade (`app.go`) is a **thin adapter**: it validates "is storage wired?", calls `app.Service`/`app.Providers`, and forwards Wails runtime calls. Business logic lives in `internal/app`.
- `internal/ai` and `internal/storage` are **siblings, not a stack**: `internal/ai` never imports `internal/storage` and vice versa. `internal/app` is the only package that composes them.
- Every OS-touching feature is split into a **pure, platform-free core** and a `_windows.go` adapter (`hotkey.go`/`hotkey_windows.go`, `tray.go`/`tray_windows.go`, `window/place.go`/`place_windows.go`, `autostart.go`/`autostart_windows.go`). `_other.go` stubs keep non-Windows builds/tests compiling.
- Streaming is **event-based**: the backend emits one `ChatEvent` (`chat:event`) with a `type` discriminator over the Wails event bus; the frontend matches deltas by `requestId` + `conversationId` + `messageId`.
- Secrets never cross to the frontend: the API key is read from the OS store only inside `internal/app/providers.go` `Resolve` and passed directly into the HTTP client.

## Layers

**Frontend (React + WebView2):**
- Purpose: Render pet/chat/settings and manage visual state.
- Location: `frontend/src/`
- Contains: `App.tsx` (mode switch + window sizing), `components/{chat,pet,settings,ui}`, `hooks/`, `styles/`, `types/`.
- Depends on: generated Wails bindings `frontend/wailsjs/go/main/App` and runtime `frontend/wailsjs/runtime/runtime`.
- Used by: the user.

**Wails facade (`app.go`):**
- Purpose: Expose bound methods, own lifecycle and OS orchestration.
- Location: `app.go` (package `main`).
- Contains: bound methods (`SendMessage`, `ListProviders`, `SetWindowMode`, `ToggleChat`, `SetAutostart`, …), `startup`/`shutdown`, `toggleFor` decision table.
- Depends on: `internal/app`, `internal/ai`, `internal/attach`, `internal/autostart`, `internal/config`, `internal/credentials`, `internal/hotkey`, `internal/storage`, `internal/tray`, `internal/window`.
- Used by: `main.go` (binds it), frontend (calls it).

**App service layer (`internal/app`):**
- Purpose: Orchestration — generation, provider management, history, attachments, settings.
- Location: `internal/app/`
- Contains: `Service` (streaming + cancel), `Providers` (profiles + credentials + settings), `ChatMessage`/`Conversation` DTOs.
- Depends on: `internal/ai`, `internal/attach`, `internal/config`, `internal/credentials`, `internal/hotkey`, `internal/storage`.
- Used by: `app.go` only.

**AI layer (`internal/ai`):**
- Purpose: The network contract — build requests, parse SSE, map errors.
- Location: `internal/ai/`
- Contains: `Provider` interface, `OpenAICompatible`, `MockProvider`, `ScanSSE`, `SafeClient`, `MapHTTPError*`, `Part` constructors.
- Depends on: Go stdlib only (`net/http`, `bufio`, `encoding/json`) plus `context`.
- Used by: `internal/app` and `internal/config` (URL normalization).

**Storage layer (`internal/storage`):**
- Purpose: SQLite persistence and migrations.
- Location: `internal/storage/`
- Contains: `DB`, `migrate`, `migrations` map, repository methods over `conversations`, `messages`, `attachments`, `provider_profiles`, `app_settings`.
- Depends on: `modernc.org/sqlite` (pure Go) + `internal/config` (model types).
- Used by: `internal/app`.

**Platform adapters:**
- Purpose: Isolate Win32 details behind small interfaces.
- Location: `internal/window`, `internal/tray`, `internal/hotkey`, `internal/autostart`.
- Contains: pure cores + `_windows.go` + `_other.go` stubs.
- Used by: `app.go`.

## Data Flow

### Primary Request Path (send a message)

1. `App.tsx` renders `ChatPanel` with `useChat.send` (`frontend/src/hooks/useChat.ts:197`).
2. `useChat.send` optimistically appends the user bubble, then calls `begin` → `SendMessage(conv, text, ids)` binding (`frontend/src/hooks/useChat.ts:215`).
3. `App.SendMessage` forwards to `Service.SendMessage` (`app.go:304`).
4. `Service.generate` trims/validates, checks the busy map, resolves staged attachments, creates `context.WithCancel`, records `requestID → conversationID`, and emits `chat:event{type:"started"}` (`internal/app/service.go:91`).
5. Goroutine: `db.EnsureConversation` + `AppendUserTurn` persist the turn (`internal/app/service.go:175`); `resolve()` builds a live `ai.OpenAICompatible` from the active profile (`internal/app/providers.go:211`).
6. `buildContext` assembles messages from `db.RecentContext(20)` with the attachment window (`internal/app/service.go:196`, `internal/app/context.go:25`).
7. `live.StreamChat` POSTs to the normalized endpoint and parses the SSE stream (`internal/ai/provider.go:116`, `internal/ai/sse.go:29`).
8. Each chunk emits `chat:event{type:"delta", requestId, messageId, delta}` (`internal/app/service.go:206`); end emits `completed`/`cancelled`/`failed` and `db.UpdateMessage` finalizes status.
9. `useChat`'s `EventsOn("chat:event")` handler matches by `requestId` + `conversationId` and updates the assistant bubble (`frontend/src/hooks/useChat.ts:108`).

### Startup

1. `main.go` `wails.Run` binds `App`, sets frameless/transparent/always-on-top window options (`main.go:27`).
2. `App.startup` creates the runtime service with `ai.MockProvider` fallback, then opens SQLite (`storage.DefaultPath` → `%LOCALAPPDATA%\WortelAI\wortel.db`) and wires `app.NewProviders(db, credentials.OSStore{})` (`app.go:69`).
3. Tray starts, edge guard installs, and `applyShortcut` registers the global hotkey (`app.go:81`, `app.go:102`). Each is non-fatal on failure and surfaced via `TrayError`/`HotkeyError`.
4. `App.tsx` mounts: `useChat` loads conversation list, `useProviders` loads profiles, and settings (`GetAppearance`, `GetShortcut`, `AutostartEnabled`, `Version`) hydrate from bindings (`frontend/src/App.tsx:117`).

### Window mode switching

1. Frontend calls `SetWindowMode("pet"|"panel")` via `switchTo` (`frontend/src/App.tsx:36`).
2. `App.SetWindowMode` asks `window.SizeOf`, then `window.Apply` (Win32 placement inside the monitor work area) with a fallback to `runtime.WindowSetSize` (`app.go:178`).

**State Management:**
- Frontend: local React state via hooks. `App.tsx` owns `mode`, `draft`, `appearance`, `shortcut` (draft and attachments live in `App` so they survive the pet/panel unmount). No Redux/Zustand.
- Backend: `Service` holds `active` (conversationID→cancel) and `requests` (requestID→conversationID) maps under `sync.Mutex`; `App` holds `hidden`/`mode` under its own mutex.
- Persistence: SQLite for conversations, messages, attachments, provider profiles, and non-secret settings; Windows Credential Manager for API keys.

## Key Abstractions

**`ai.Provider` interface:**
- Purpose: Abstract a streaming chat backend so mock and HTTP are interchangeable.
- Examples: `internal/ai/provider.go:52` (`MockProvider`, `OpenAICompatible`).
- Pattern: Strategy — `Service.provider` is the fallback; `Providers.Resolve` swaps in the live client at request time.

**`credentials.Store` interface:**
- Purpose: Keep secrets out of SQLite/logs and mockable in tests.
- Examples: `internal/credentials/store.go:13`; `OSStore` (`internal/credentials/osstore.go`), `mapstore` (`internal/credentials/mapstore.go`).
- Pattern: Repository/port-adapter.

**`ChatEvent` single-variant event:**
- Purpose: One structured payload for all stream lifecycle messages.
- Examples: `internal/ai/types.go:5`; mirrored in `frontend/src/types/chat.ts:10`.
- Pattern: Discriminated union over the `type` field; contract mirrored by hand (not codegen) on the TS side.

**Platform core + tagged adapter:**
- Purpose: Keep logic testable off-Windows.
- Examples: `internal/hotkey/combo.go` (parse) vs `hotkey_windows.go` (register); `internal/window/place.go` (math) vs `place_windows.go` (Win32); `internal/tray/tray.go` (menu intent mapping) vs `tray_windows.go`.
- Pattern: Pure-core + build-tagged IO shell.

## Entry Points

**`main.go`:**
- Location: `main.go`
- Triggers: process start (`companion.exe`).
- Responsibilities: Wails options (frameless 180×180 pet window, max 460×680), embed `frontend/dist`, bind `App`, fatal dialog on error.

**`App.startup` / `App.shutdown`:**
- Location: `app.go:69`, `app.go:155`
- Triggers: Wails lifecycle callbacks.
- Responsibilities: Wire service/storage/credentials/tray/hotkey; on shutdown `CancelAll` (waits for goroutines), stop hotkey, stop tray.

**`frontend/src/main.tsx`:**
- Location: `frontend/src/main.tsx`
- Triggers: WebView loads `index.html`.
- Responsibilities: Mount React root, cancel default file-drop navigation so WebView2 does not navigate away.

**`chat:event` event bus:**
- Location: emitted at `internal/app/service.go`, consumed at `frontend/src/hooks/useChat.ts:169`.
- Triggers: `runtime.EventsEmit(ctx, ...)`.
- Responsibilities: Carry `started|delta|completed|cancelled|failed` to the UI.

## Architectural Constraints

- **Threading:** Go goroutines per generation, tracked by `sync.WaitGroup`; `CancelAll` waits for them at shutdown so no write races teardown (`internal/app/service.go:252`). Wails runtime calls are thread-safe; tray callbacks run on the tray thread.
- **Global state:** Package-level singletons — `internal/hotkey` and `internal/tray` hold process-wide registered state; `internal/window/windows` caches the main window handle under a package mutex (`internal/window/activate_windows.go:16`). There is intentionally one window.
- **Circular imports:** Avoided by design — `internal/ai` and `internal/storage` are leaves that never import each other or `internal/app`; composition happens only in `internal/app` and `app.go`.
- **Failure degradation:** Storage, tray, hotkey, and edge-guard failures are all non-fatal: the app runs without them and reports via `HotkeyError`/`TrayError` (`app.go:72`, `app.go:81`, `app.go:91`).
- **Secret boundary:** API key exists only in the form field until `Save`, then in `Providers.Resolve` backend memory; it is wiped best-effort after use (`internal/app/providers.go:230`). Never logged, never emitted, never in SQLite.

## Anti-Patterns

### Putting business logic in the Wails facade

**What happens:** It is tempting to add logic directly to `app.go` bound methods.
**Why it's wrong:** `app.go` is package `main` and untestable off-Windows; logic there cannot be unit-tested and mixes adapter with domain.
**Do this instead:** Keep `app.go` as a validation + forward layer; put logic in `internal/app` (see how `SendMessage` at `app.go:304` just delegates to `Service.SendMessage`). Add new features as methods on `Service`/`Providers` with tests in `internal/app/*_test.go`.

### Making `internal/ai` aware of storage

**What happens:** Reaching into `internal/storage` from `internal/ai` to build context.
**Why it's wrong:** It creates the forbidden dependency direction and a package cycle risk.
**Do this instead:** Compose in `internal/app` — see `buildContext` reading `storage.ContextRow` and producing `[]ai.Message` (`internal/app/context.go:25`).

### Committing platform code without a pure core

**What happens:** Adding Win32 calls directly into an `app.go` method or a mixed file.
**Why it's wrong:** The logic becomes unbuildable/unrunnable off Windows and untestable.
**Do this instead:** Put pure math/mapping in the untagged file and gate the syscall in `_windows.go` with `//go:build windows`, plus an `_other.go` stub — see `internal/window/place.go` + `place_windows.go` + `place_other.go`.

### Passing multiple values across the Wails binding

**What happens:** Returning `(value, bool)` from a bound method.
**Why it's wrong:** Wails treats the second non-error return as the error slot, so a tuple arrives as a bare struct and a zero value cannot be distinguished from a real one (documented at `app.go:422`).
**Do this instead:** Return a pointer (`*config.ProviderProfile`, `nil` → JSON `null`) as `ActiveProvider` does (`app.go:429`).

## Error Handling

**Strategy:** Map every provider/transport failure to a frontend-safe `ai.PublicError` (code + message + retryable); never forward raw bodies, headers, or secrets.

**Patterns:**
- HTTP status → `PublicError` via `MapHTTPErrorWithHeaders`/`MapHTTPErrorWithBody` (`internal/ai/errors.go:21`, `errors.go:56`), honoring `Retry-After` on 429/503.
- Transport → `toNetError` returns raw `context.Canceled`/`DeadlineExceeded` so the service can distinguish cancel from failure (`internal/ai/errors.go:88`, consumed at `internal/app/service.go:216`).
- Storage failures during a live stream are **best-effort**: a DB error must not abort the stream (see comment at `internal/app/service.go:172`).
- User-facing messages are Indonesian and actionable; the frontend shows a per-reply `errorText` from the `failed` event and a global `error` for non-reply failures (`frontend/src/hooks/useChat.ts:153`).

## Cross-Cutting Concerns

**Logging:** `runtime.LogWarningf` for degraded subsystems only (`app.go:92`, `app.go:149`). Request headers/bodies/secrets are never logged.
**Validation:** Backend validates all input even though the frontend pre-checks — provider name/URL/models (`internal/config/providers.go:81`), attachment kind/size (`internal/attach/validate.go:60`), window mode (`internal/window/place.go:32`), appearance ranges (`internal/app/appearance.go:54`).
**Authentication:** Per-provider bearer token read from Windows Credential Manager; `SafeClient` strips `Authorization` on cross-host redirects (`internal/ai/http.go:14`).
**Safe rendering:** AI markdown is untrusted — `SafeMarkdown` uses `react-markdown` + `rehype-sanitize` with no raw HTML and http/https-only links (`frontend/src/components/chat/SafeMarkdown.tsx:73`).

---

*Architecture analysis: 2026-10-11*
