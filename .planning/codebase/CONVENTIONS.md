---
last_mapped_commit: b55c8e5bb1c4199b24162ca36f0c612ef255a01c
last_mapped_at: 2026-10-11
---
# Coding Conventions

**Analysis Date:** 2026-10-11

## Overview

Two-language codebase: a Go 1.25 backend (Wails v2) at the repo root and under `internal/`, plus a React 19 + TypeScript frontend under `frontend/src/`. Conventions are enforced partly by tooling (`.editorconfig`, `gofmt`, `go vet`, `tsc --strict`) and heavily by the project's `AGENTS.md` contract. There is **no ESLint, Prettier, or Biome config** in the repo — frontend style is upheld by TypeScript strict mode and adherence to existing patterns.

## Naming Patterns

**Files (Go):**
- Lowercase, single-word, snake_case-free: `service.go`, `providers.go`, `mapstore.go`, `sse.go`, `errors.go`, `testhelper_test.go`.
- Platform variants use the `_windows` / `_other` suffix: `hotkey_windows.go`, `hotkey_other.go`, `tray_windows.go`, `autostart_windows.go`, `fatal_other.go`.
- Test files are always `<subject>_test.go` co-located in the same package (`provider_test.go`, `db_test.go`).

**Files (TypeScript/React):**
- Components: `PascalCase.tsx` — `ChatPanel.tsx`, `SafeMarkdown.tsx`, `ProviderForm.tsx`, `MessageItem.tsx`.
- Hooks: `camelCase.ts` with the `use` prefix — `useChat.ts`, `useAttachments.ts`, `useProviders.ts`, `usePetDrag.ts`.
- Library utilities: `camelCase.ts` — `formatSize.ts`, `markdownToText.ts`.
- Types: `camelCase.ts` in `frontend/src/types/` — `chat.ts`, `provider.ts`.
- Tests: co-located with the subject, suffixed `.test.ts` / `.test.tsx`. Component integration tests use a scenario suffix: `ChatPanel.attach.test.tsx`, `MessageItem.attach.test.tsx`.

**Functions/Methods (Go):**
- Exported: `PascalCase` (`SendMessage`, `MapHTTPErrorWithBody`, `NormalizeModels`, `Classify`, `Fit`, `StagePaths`).
- Unexported: `camelCase` (`isCtx`, `flush`, `mk`, `truncateKeepExt`, `firstOr`, `titleFor`).
- Constructors: `NewX` — `NewService`, `NewProviders`, `NewRuntimeService`, `NewStaging`, `NewMapStore`.
- Small unexported closures for local helpers inside a function, e.g. `flush` in `internal/ai/sse.go`, `mk` in `internal/ai/errors.go`, `reject` and `fail` in `internal/app/service.go`.

**Functions/Variables (TypeScript):**
- Functions and hooks: `camelCase` (`loadConversation`, `refreshList`, `busyNow`, `setActiveModel`).
- Component functions: `PascalCase` (`ChatPanel`, `MessageItem`, `SafeMarkdown`, `IconSliders`).
- Constants: `UPPER_SNAKE_CASE` (`MAX_ATTACHMENTS` in `frontend/src/hooks/useAttachments.ts`).
- Backend binding calls imported from `wailsjs` are PascalCase and match the Go method names 1:1 (`SendMessage`, `ListProviders`, `StageAttachmentPaths`).

**Types (both languages):**
- Go types/structs: `PascalCase` (`Service`, `Providers`, `PublicError`, `ProviderProfile`, `MessageRow`, `Meta`, `Staging`).
- Go error sentinels: `Err`-prefixed `PascalCase` vars (`ErrNotFound` in `internal/credentials/store.go`, `ErrUnsupported` in `internal/attach/validate.go`, `ErrTooMany`, `ErrTooBig`, `ErrUnknown`, `ErrDuplicate` in `internal/attach/`).
- Lowercase error vars use the `err` prefix (`errNoDB`, `errNoConversation`, `errConversationBusy` in `internal/app/history.go`; `errNoAttachmentStorage` in `internal/app/attachments.go`).
- TypeScript interfaces: `PascalCase`, prefixed `I` only never — plain names (`ChatEvent`, `ChatMessage`, `AttachmentMeta`, `ProviderProfile`, `TestResult`).
- TypeScript union string literals: lowercase (`ChatStatus = "idle" | "sending" | "streaming" | "complete" | "cancelled" | "error"` in `frontend/src/types/chat.ts`).

## Code Style

**Formatting:**
- Enforced for Go by `gofmt` in CI (`.github/workflows/ci.yml` runs `gofmt -l .` and fails on any output). Tabs for Go indentation, spaces for everything else per `.editorconfig`.
- `.editorconfig` sets: UTF-8, LF line endings, final newline, trailing-whitespace trimming, 2-space indent for non-Go, tabs for `*.go` and `Makefile`. Markdown keeps trailing whitespace; PowerShell uses CRLF.
- `.gitattributes` normalizes all text to LF (`* text=auto eol=lf`) except `*.ps1`, `*.bat`, `*.cmd` which stay CRLF.
- Frontend has no formatter configured. Follow existing 2-space, double-quoted, semicolon-terminated style seen throughout `frontend/src/`.

**Linting:**
- Go: `go vet ./...` is mandatory in CI.
- TypeScript: no linter. `tsc --noEmit` (strict) is the only static gate. `tsconfig.json` enables `strict: true`, `forceConsistentCasingInFileNames`, `isolatedModules`. Do not introduce `any` without justification (`AGENTS.md` mandate); tests cast with `as unknown as` when synthesizing events (see `usePetDrag.test.ts`).

## Import Organization

**Go:**
1. Standard library.
2. Third-party packages (`github.com/...`).
3. Local module imports under `companion/internal/...`.

Groups are blank-line separated. Example (`internal/app/service.go`):

```go
import (
	"context"
	"errors"
	"strings"
	"sync"

	"companion/internal/ai"
	"companion/internal/attach"
	"companion/internal/storage"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)
```

**TypeScript:**
1. React / library imports.
2. Wails bindings from `../../wailsjs/go/main/App` and `../../wailsjs/runtime/runtime`.
3. Local relative imports (`../types/...`, `./SafeMarkdown`).

Example (`frontend/src/hooks/useChat.ts`):

```ts
import { useCallback, useEffect, useRef, useState } from "react";
import { EventsOff, EventsOn } from "../../wailsjs/runtime/runtime";
import { CancelGeneration, History, /* ... */ } from "../../wailsjs/go/main/App";
import type { AttachmentMeta, ChatEvent, ChatMessage, ChatStatus } from "../types/chat";
```

**Path Aliases:** None. All frontend imports are relative. There is no `@/` alias in `frontend/tsconfig.json`.

## Go Structural Conventions

**Business logic vs Wails adapter:** Backend logic lives in `internal/` packages (`ai`, `app`, `attach`, `config`, `credentials`, `storage`, `hotkey`, `tray`, `window`, `autostart`). The root `app.go` is a thin binding layer. `AGENTS.md` requires keeping business logic out of the Wails adapter. The `app.Service` type in `internal/app/service.go` orchestrates and accepts an `emit func(string, ai.ChatEvent)` — production wires `runtime.EventsEmit` via `NewRuntimeService`, tests inject a sink.

**Frontend-safe types:** Backend structs that cross the binding are documented as secret-free and carry JSON tags. Examples: `config.ProviderProfile` (`internal/config/providers.go`) states "no secret, only configured flag"; `attach.Meta` (`internal/attach/validate.go`) states "no bytes, no path"; `app.TestResult` (`internal/app/providers.go`) "Never carries secrets". Follow this when adding a type that reaches React.

**Context propagation:** `context.Context` is the first parameter for cancellable/network operations (`SendMessage(ctx, ...)`, `StreamChat(ctx, ...)`, `ScanSSE(ctx, ...)`, credential `Store` methods). Cancellation cleanup must be explicit — see `CancelAll` waiting on `s.wg` (`internal/app/service.go`).

**Error wrapping:** Use `fmt.Errorf("...: %w", err)` to preserve causes (pervasive in `internal/storage/db.go`), and `errors.Is` / `errors.As` to inspect them (`internal/app/service.go` uses `errors.Is(err, context.Canceled)`; `internal/ai/errors.go` uses `errors.As` for `Timeout()` detection).

## Error Handling

**Go:**
- Return `error`, never panic, on recoverable failures. Test names assert this explicitly: `TestUseAfterCloseDoesNotPanic`, `TestOpenRejectsEmptyPath` (`internal/storage/db_test.go`).
- Safe, actionable, Indonesian-language messages for end users: `PublicError` (`internal/ai/errors.go`) maps HTTP status to a user-readable `Message` (e.g. "API key ditolak. Cek key dan endpoint."). Never include secrets in error text — `AGENTS.md` mandate, asserted by `TestSaveAndResolve` which rejects a profile JSON containing the API key.
- Sentinel errors for control flow: `errors.Is(err, ErrNotFound)`, `errors.Is(err, ErrActiveDelete)`.
- A `PublicError` struct carries `Code`, `Message`, `Retryable`, `StatusCode`, `RetryAfterSec` for UI consumption.

**TypeScript/React:**
- `try/catch` with `e instanceof Error ? e.message : "<fallback>"` for user-facing `error` state. See `frontend/src/hooks/useProviders.ts`, `frontend/src/hooks/useChat.ts`.
- Guards must be released in `finally` so a rejected binding call never locks the UI permanently — see `cancel()` in `frontend/src/hooks/useChat.ts`:

```ts
try {
  await CancelGeneration(convRef.current);
} catch (e) {
  setError(e instanceof Error ? e.message : "Gagal membatalkan.");
} finally {
  sendingRef.current = false;
  activeRequest.current = null;
  setStatus("idle");
}
```

- Error state is split: per-message reply errors live on the message (`errorText`), while global errors (rename/delete/send rejection) live in the hook's `error` state (`frontend/src/types/chat.ts`, `useChat.ts`).

## Logging

- Go: no `log.Printf` on request/header objects that may carry secrets (`AGENTS.md`). No structured logging framework is configured; error information travels through return values and `PublicError`, not log statements.
- Frontend: no logging framework. User-facing feedback is rendered in the UI (error banners, inline `errorText`), not written to console.

## Comments

**When to Comment:**
- Comments explain **why**, especially non-obvious invariants, races, and rationale. They are dense and intentional, e.g. `SafeClient` in `internal/ai/http.go` explains why the Authorization header is dropped on cross-host redirects.
- Security and lifecycle reasoning is documented inline: `internal/app/service.go` explains best-effort persistence and retry reconciliation; `useChat.ts` explains the `settled` set that prevents a fast-failure requestId from locking `send()`.
- Migration statements in `internal/storage/db.go` each carry a comment describing the version's purpose and idempotency.

**JSDoc/TSDoc:**
- Go uses standard `// ExportedName ...` doc comments on every exported identifier.
- TypeScript uses `/** ... */` JSDoc on interface fields where semantics are non-obvious, e.g. `errorText?: string; /** Why this reply failed... */` in `frontend/src/types/chat.ts`.
- The `/* Hallmark · component: ... */` header block convention appears on UI/design components such as `frontend/src/components/ui/Icon.tsx`.

## Function Design

**Size:** Small, single-purpose helpers are preferred (`flush`, `truncateKeepExt`, `firstOr`, `hasRow`, `titleFor`). Orchestration functions like `Service.generate` (`internal/app/service.go`) are longer and heavily commented because they encode concurrency and lifecycle rules.

**Parameters:** Explicit, typed parameters. `context.Context` first when applicable. Long argument lists are acceptable for binding calls (e.g. `SendMessage(ctx, conversationID, text, attachmentIDs)`) — validation happens inside.

**Return Values:** Go uses the `(value, error)` convention; multi-value returns are used for lookups (`(mime, data, ok)` in `Service.Attachment`, `(provider, model, error)` in `Providers.Resolve`). TypeScript hooks return a typed object of state + callbacks rather than arrays (see the return of `useChat`, `useProviders`).

## Module Design

**Exports (Go):** Small exported surface; unexported helpers stay package-private. Interfaces are declared by the consumer-side concern and kept minimal — `credentials.Store` (`internal/credentials/store.go`) has exactly four methods (`Set`, `Get`, `Exists`, `Delete`). Test doubles implement these (`MapStore` in `internal/credentials/mapstore.go`).

**Exports (TypeScript):** Named exports only (`export function useChat`, `export const IconSliders`). No default exports observed in source files.

**Barrel Files:** None. Imports reference concrete files directly.

## Security Conventions (project-critical)

These are treated as hard conventions, not preferences (`AGENTS.md`, verified by tests):
- Never store or log API keys in source, SQLite, `localStorage`, or logs. Credentials live in the OS store via `internal/credentials/`.
- Never render raw HTML from AI output. `frontend/src/components/chat/SafeMarkdown.tsx` sanitizes with `rehype-sanitize`; raw HTML stays inert and only `http`/`https` links become anchors (guarded by `SafeMarkdown.test.tsx`).
- Persisted/returned types must be secret-free (`ProviderProfile.HasCredential` is a boolean flag, never the key).
- Markdown from AI is untrusted input.

---

*Convention analysis: 2026-10-11*
