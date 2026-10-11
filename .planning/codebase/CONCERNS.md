---
last_mapped_commit: b55c8e5bb1c4199b24162ca36f0c612ef255a01c
last_mapped_at: 2026-10-11
---
<!-- refreshed: 2026-10-11 -->

# Codebase Concerns

**Analysis Date:** 2026-10-11

## Tech Debt

**Persistent error swallowing in the generation goroutine:**
- Issue: `internal/app/service.go` ignores nearly every database return with `_ =`. `EnsureConversation`, `AppendMessage`, `AppendUserTurn`, `UpdateMessage`, `PrepareRetry` all have their errors dropped (lines 175-224). A failed write silently loses history, and the only signal is that the history list is empty after a restart.
- Files: `internal/app/service.go` (lines 175-224), `internal/storage/chat.go` (`PrepareRetry` line 163 has a comment admitting the fallback)
- Impact: Conversations and messages can vanish without any user-visible error. The retry path (`reused := ok || err != nil`, line 162) deliberately treats a DB error as "row exists", which duplicates the user question when the DB is actually broken.
- Fix approach: Surface at least one terminal `chat:event` with a distinct `persist_failed` code when a write in the active turn fails, or log via `runtime.LogErrorf` (never via `log.Printf`, per `AGENTS.md`). Do not silently accept the duplicate-question trade-off.

**`ErrUnknown`/staging errors reach the UI as raw strings:**
- Issue: Staging and attachment errors cross the Wails binding as Go `error` strings; `useAttachments.messageOf` (`frontend/src/hooks/useAttachments.ts:29`) falls back to `String(e)`. There is no `PublicError`-style structured code for attachment failures (unlike chat, which has `ai.PublicError`).
- Files: `internal/attach/staging.go`, `internal/attach/validate.go`, `frontend/src/hooks/useAttachments.ts`
- Impact: The frontend cannot branch on failure reason (retryable vs not, too-many vs too-big); it only displays the backend message. Adding a new attachment rule requires matching on prose.
- Fix approach: Return a structured result (code + message) from `StagePaths`/`StageData`, mirroring `ai.PublicError`, and map it in `useAttachments` to typed UI states.

**Duplicated `errNoDB` definitions:**
- Issue: The same "storage belum tersedia" message is re-declared in three places: `errNoDB` (`internal/app/history.go:11`), `errStorageUnavailable()` (`app.go:448`), and `errNoAttachmentStorage` (`internal/app/attachments.go:16`). Each has subtly different wording.
- Files: `internal/app/history.go`, `app.go`, `internal/app/attachments.go`
- Impact: Inconsistent user-facing text and no single place to change the "restart the app" guidance.
- Fix approach: Consolidate into one exported sentinel in `internal/app` and reference it from `app.go`.

**`gofmt`/`go vet` are clean, but there is no lint gate for Go:**
- Issue: CI runs `gofmt`, `go vet`, `go test`, `go test -race` only (``.github/workflows/ci.yml``). No `staticcheck`/`golangci-lint`. Ignored-error patterns (`_ =`) therefore never get flagged automatically.
- Files: `.github/workflows/ci.yml`
- Impact: The error-swallowing above goes unnoticed; future debt of the same shape accumulates silently.
- Fix approach: Add `staticcheck` to the `backend` CI job; it flags unused assignments and many correctness issues.

**Monolithic `app.go` binding surface:**
- Issue: `app.go` is 467 lines and holds every Wails binding (tray, hotkey, appearance, provider, history, attachment, autostart, version). It mixes adapter wiring with mode/visibility state.
- Files: `app.go`
- Impact: Any binding change touches the same file as window/tray logic; hard to test in isolation.
- Fix approach: Split into per-domain binding structs (e.g. `providers_binding.go`, `history_binding.go`) that delegate to the existing `internal/app` service, keeping `app.go` for lifecycle only.

**Frontend mega-components:**
- Issue: `frontend/src/components/chat/ChatPanel.tsx` (505 lines) and `frontend/src/components/settings/ProviderForm.tsx` (374 lines) carry both layout and a lot of imperative UI state (drag depth, rename/delete arming, scroll-stick, capture).
- Files: `frontend/src/components/chat/ChatPanel.tsx`, `frontend/src/components/settings/ProviderForm.tsx`
- Impact: Every UI tweak risks the carefully documented invariants (draft/error/stick behaviour) at the top of the file.
- Fix approach: Extract `useConversationActions` (rename/delete arming) and `useStickyScroll` hooks; the invariants then live with the code that enforces them.

## Known Bugs

**`usePetDrag`/`usePetSignal` interaction is only proven by unit tests, not real input:**
- Symptoms: The 4 px drag threshold and click-vs-drag discrimination (ADR-041/ADR-029) are covered by `usePetDrag.test.ts` (10 tests) but the actual native drag and single-click feel has repeatedly needed manual user verification (`docs/08-roadmap-testing.md` marks several items `[~]`).
- Files: `frontend/src/hooks/usePetDrag.ts`, `frontend/src/hooks/usePetSignal.ts`
- Trigger: Any change to `--wails-draggable` inheritance or pet event handlers.
- Workaround: Manual test on a real build; see ADR-029/ADR-041 in `docs/09-decisions.md`.
- Fix approach: Add a CDP-driven probe (as ADR-042 did for placement) to a documented script under `scripts/`; manual-only verification is the current regression risk.

**`toggleFor` depends on Wails having no visibility probe:**
- Symptoms: `App.hidden`/`App.mode` are tracked in memory (`app.go:32-35`) because Wails exposes no `IsVisible`. If the user minimises via the taskbar rather than the tray, `hidden` is stale and the global shortcut may bring the window forward instead of collapsing to pet.
- Files: `app.go` (`toggleFor`, `ToggleChat`, `setHidden`), `internal/window/activate_windows.go`
- Trigger: Minimise/restore through Windows UI instead of app controls.
- Workaround: None documented; the tray callbacks keep the flag in sync only for app-initiated hide/show.
- Fix approach: Query `IsWindowVisible`/`IsIconic` via the existing `internal/window` user32 wrappers instead of trusting the in-memory flag.

## Security Considerations

**API key zeroing is best-effort and does not cover all copies:**
- Risk: `Providers.Resolve` (`internal/app/providers.go:228-234`) and `Test` (lines 292-296) wipe the `[]byte` clone after use, but the key also lives as a `string` (`key := string(secret)`) which is immutable and cannot be zeroed, and it is placed into `ai.OpenAICompatible.APIKey` where it persists for the provider's lifetime.
- Files: `internal/app/providers.go`, `internal/ai/provider.go` (`OpenAICompatible.APIKey`)
- Current mitigation: The secret never reaches the DB, frontend, or logs; `SafeClient` drops `Authorization` on cross-host redirects (`internal/ai/http.go`); `AGENTS.md` forbids logging request/header objects.
- Recommendations: Prefer `[]byte` throughout the client and set `Authorization` from bytes immediately before the request; document that the `string` copy is unavoidable given `net/http`'s header API. Consider a per-request credential fetch rather than a long-lived `APIKey` field.

**`NormalizeBaseURL` allows cleartext HTTP to arbitrary hosts:**
- Risk: `internal/ai/provider.go:88-105` accepts any `http://` host. `config.Validate` flags non-loopback HTTP as `insecure` but only warns; the request still goes out in cleartext, exposing the bearer token on the wire.
- Files: `internal/ai/provider.go`, `internal/config/providers.go` (`isLoopback`, `InsecureTransport`)
- Current mitigation: `InsecureTransport` surfaced to the UI as a warning; loopback exempted.
- Recommendations: Require explicit user confirmation before saving a non-loopback HTTP provider, or refuse entirely (local providers via `localhost`/`127.0.0.1` remain allowed).

**Attachment serving route has no MIME allowlist on read:**
- Risk: `GET /attachments/{id}` (`internal/app/attachment_handler.go`) serves stored bytes with whatever `mime` was recorded at classify time. `X-Content-Type-Options: nosniff` is set, which mitigates sniffing, but the content type is trusted from storage.
- Files: `internal/app/attachment_handler.go`, `internal/attach/validate.go` (`Classify`)
- Current mitigation: UUID regex gate rejects anything but a well-formed id; `Classify` derives mime from magic bytes (`http.DetectContentType`) plus a fixed extension map; `Content-Disposition: inline`.
- Recommendations: Consider `Content-Disposition: attachment` for non-image kinds, or verify the stored mime against the allowlist at serve time. Low severity given the WebView is the only consumer.

## Performance Bottlenecks

**Image re-billing across the context window:**
- Problem: `AttachmentTurns = 3` (`internal/app/context.go:18`) re-sends full attachment bytes for the three newest user turns on every request. Image/PDF bytes are read from SQLite on each send and base64-inlined by `ai.Parts`.
- Files: `internal/app/context.go`, `internal/storage/attachments.go`, `internal/ai/parts.go`
- Cause: Providers charge per-request for vision input; the window is a deliberate cost/latency trade-off (documented in `docs/05`).
- Improvement path: Add an explicit per-conversation "attachment context budget" setting, or hash attachments and send a placeholder when the model already saw an identical image. The current byte budget (`attach.MaxMessageBytes`) bounds request size but not provider billing.

**SQLite BLOB reads on the hot path:**
- Problem: `db.AttachmentData` is called for every attachment inside the 3-turn window on every send (`internal/storage/attachments.go`); metadata reads avoid `data` but full reads do not.
- Files: `internal/storage/attachments.go`, `internal/app/context.go`
- Cause: Bytes are stored in the `attachments` table; there is no on-disk blob store or cache.
- Improvement path: Measure with a multi-image conversation; if the read is measurable, cache decoded parts per `message_id` in memory for the session.

**Frontend bundle growth from markdown stack:**
- Problem: `react-markdown` + `remark-gfm` + `rehype-sanitize` grew the JS bundle from ~207 kB to ~372 kB (gzip 65→116 kB) per ADR-028.
- Files: `frontend/package.json`, `frontend/src/components/chat/SafeMarkdown.tsx`
- Cause: Accepted trade-off for safe, correct GFM rendering.
- Improvement path: Lazy-load `SafeMarkdown` with `React.lazy` so the initial pet-mode payload stays small; measurement should precede any change.

## Fragile Areas

**`--wails-draggable` inheritance:**
- Files: `frontend/src/styles/chat.css`, `frontend/src/styles/pet.css`, `frontend/src/styles/settings.css`, `frontend/src/App.tsx`, `frontend/src/components/chat/ChatHeader.tsx`
- Why fragile: The custom property inherits to all children, so any new control placed inside a drag region must explicitly set `--wails-draggable: none` or the first click starts a window drag. This has regressed at least twice (ADR-029, ADR-043).
- Safe modification: When adding any interactive element inside `#chat-header`, `#settings-header`, or pet surfaces, set `--wails-draggable: none` and verify with a real click.
- Test coverage: No automated guard; only documented convention.

**`ChatPanel` streaming/scroll invariants:**
- Files: `frontend/src/components/chat/ChatPanel.tsx` (`stick`, `lastCount`, `lastLen`, `newBelow`)
- Why fragile: The scroll-stick logic is imperative ref bookkeeping tied to render order; a retry that drops a bubble or a status-only update is deliberately treated as "not news" (lines 159-176). Changing the message update shape in `useChat` can silently break the "Balasan baru" affordance.
- Safe modification: Keep the "new text = new message or newest message growing" rule intact; add a component test that drives deltas and a retry.
- Test coverage: `ChatPanel.attach.test.tsx` covers attachments, not scroll.

**`useChat` request lifecycle guards:**
- Files: `frontend/src/hooks/useChat.ts` (`sendingRef`, `activeRequest`, `settled`)
- Why fragile: Three separate guards coordinate the race between the binding return value and the out-of-order `started`/`failed` events. ADR-039 exists because a missing `finally` locked the UI permanently. Any new terminal event or early-return path risks reintroducing the lock.
- Safe modification: Every path that sets `sendingRef`/`activeRequest` must clear them; keep `settled` in sync when adding event types.
- Test coverage: `useChat.test.ts` covers cancel failures and basic states; fast-failure event ordering is only partially covered.

**Window subclassing under the Wails wndproc:**
- Files: `internal/window/guard_windows.go`, `internal/window/place_windows.go`
- Why fragile: `KeepOnScreen` chains `CallWindowProcW` beneath Wails' own `GWLP_WNDPROC`. If a future Wails version installs its own subclass or replaces the proc, the guard is bypassed or calls into freed memory.
- Safe modification: Pin the Wails version (already `v2.16.0`); re-verify `WM_MOVING` handling after any Wails upgrade.
- Test coverage: `internal/window/place_test.go` covers pure geometry; the subclass has no automated test (`WM_MOVING` needs real input).

**Direct `unsafe`/Win32 struct layouts:**
- Files: `internal/tray/tray_windows.go` (`notifyIconData`, `wantNIDSize = 976`), `internal/hotkey/hotkey_windows.go` (`msgT`), `internal/window/guard_windows.go`
- Why fragile: Hand-rolled Win32 structs must match ABI exactly. `TestNIDSize` pins `NOTIFYICONDATAW` size, but the other layouts rely on comments and are not all size-asserted.
- Safe modification: Add size assertions for `msgT` and the `wndClassEx`/`point` structs as was done for `notifyIconData`.
- Test coverage: `internal/tray/tray_test.go` asserts the NID size; hotkey/window layouts lack equivalent assertions.

## Scaling Limits

**In-memory attachment staging:**
- Current capacity: Max 4 files / 20 MB total per message (`internal/attach/validate.go`), held entirely in `Staging` memory (`internal/attach/staging.go`).
- Limit: A user staging a 15 MB PDF plus images keeps ~20 MB resident until send; concurrent multi-window use multiplies this (though the app is single-window by design).
- Scaling path: The cap is enforced and pre-checked, so this is bounded. If larger attachments are ever allowed, stage to a temp file instead of memory.

**SQLite single-file database with no size management:**
- Current capacity: `%LOCALAPPDATA%\WortelAI\wortel.db` with WAL; attachments stored as BLOBs (`internal/storage/db.go` migration v5).
- Limit: Image/PDF BLOBs accumulate indefinitely; deleting a conversation cascades attachments (`ON DELETE CASCADE`), but there is no pruning of old conversations, no `VACUUM`, and no size display.
- Scaling path: Add a "data usage" readout in Settings and an optional prune of conversations older than N days; run `VACUUM` after bulk deletes.

**`requests` map is bounded only by cancellation:**
- Current capacity: `Service.requests` (`internal/app/service.go:21`) is deleted in the goroutine's `defer`; `CancelAll` waits on the `WaitGroup`.
- Limit: Correct today (ADR-025 fixed an earlier leak), but the map has no hard cap; a pathological client that starts thousands of concurrent conversations would grow it — prevented only by the per-conversation `active` guard.
- Scaling path: Not user-reachable in the single-window MVP; leave as is but note the invariant.

## Dependencies at Risk

**`github.com/zalando/go-keyring` v0.2.6:**
- Risk: Wraps Windows Credential Manager via `wincred`; credential behaviour differs across Windows builds and can fail under non-interactive sessions (`internal/credentials/osstore.go` maps any error to a generic `errStoreOp`).
- Impact: If the OS store is unavailable, provider save/get fails with an opaque message (`"gagal menyimpan API key ke credential store"`) and the user cannot chat.
- Migration plan: The `credentials.Store` interface (`internal/credentials/store.go`) already abstracts this; `mapstore.go` provides an in-memory implementation for tests. A file-based or DPAPI fallback could be added behind the same interface if OS store proves unreliable.

**`github.com/wailsapp/wails/v2` pinned to v2.16.0:**
- Risk: The code subclasses the Wails window proc (`internal/window/guard_windows.go`), inspects binding behaviour (ADR-031 documents a non-obvious multi-return bug), and relies on `WindowClassName`. These are version-sensitive internals.
- Impact: A major/minor Wails upgrade can break placement, tray, or bindings in ways unit tests cannot catch.
- Migration plan: Re-run the ADR-042 placement probe and a manual drag/tray smoke after any Wails bump; keep `go.mod` pinned until v3 is needed.

**`modernc.org/sqlite` (pure-Go, no CGO):**
- Risk: Pure-Go SQLite is slower than the CGO driver and its `_pragma` URI handling is version-specific (`internal/storage/db.go:44` uses `?_pragma=...` query params plus explicit `PRAGMA` statements).
- Impact: If a future version changes URI pragma semantics, WAL/foreign-key settings may silently not apply.
- Migration plan: The explicit `PRAGMA` loop after `Open` is the real guard; keep both and add an assertion test that `foreign_keys` and `journal_mode` are active.

## Missing Critical Features

**No conversation export/backup:**
- Problem: Users cannot export or back up `wortel.db` from within the app; uninstall instructions in `README.md` tell them to delete the folder manually.
- Blocks: Safe data migration without leaving the app; no path to preserve history across machines.
- Note: This is arguably out of MVP scope but is a common expectation for a history-bearing app.

**No "Data & privacy" settings section:**
- Problem: ADR-044 notes the `Data & privacy` group from `docs/10 §6` was not built; only a one-line statement exists in the About section.
- Blocks: Users cannot see storage location, size, or clear data from the UI.
- Fix approach: Add a read-only "data location + size" block to the existing About section as a low-risk first step.

**`speaking` state is not distinct from `thinking`:**
- Problem: ADR-044 records that the backend sends a single `streaming` status for both the wait-for-token phase and the delta phase, so the pet cannot differentiate `speaking`.
- Blocks: A richer pet animation; the gap is documented rather than half-implemented (correctly).
- Fix approach: Emit a `first_delta` boundary event from `Service.generate` if the distinction is ever needed.

## Test Coverage Gaps

**Tray message loop and menu dispatch:**
- What's not tested: `wndProcFn` message handling, `popup`/`TrackPopupMenu`, explorer-restart re-add (`TaskbarCreated`), and the `WM_COMMAND` → action mapping.
- Files: `internal/tray/tray_windows.go`, `internal/tray/tray_test.go` (only asserts NID size)
- Risk: A tray regression (blank icon, dead menu) ships unnoticed; CI cannot drive a real notification area.
- Priority: Medium — needs a manual smoke checklist entry rather than unit tests.

**Hotkey registration on a real input path:**
- What's not tested: Whether `RegisterHotKey` actually fires `WM_HOTKEY` for the configured combo; ADR-020 states SendKeys does not trigger OS hotkeys, so real verification is manual.
- Files: `internal/hotkey/hotkey_windows.go`, `internal/hotkey/hotkey_windows_test.go`
- Risk: A rebind could register successfully but never fire on the user's hardware.
- Priority: Medium — document as manual.

**Attachment serving and classification end-to-end:**
- What's not tested: `AttachmentHandler` is exercised only indirectly; there is no test that `Classify` rejects a disguised file (e.g. a `.txt` whose bytes are an executable) via the HTTP route.
- Files: `internal/app/attachment_handler.go`, `internal/attach/validate.go`, `internal/attach/validate_test.go`
- Risk: A regression in magic-byte detection lets an unexpected type through staging.
- Priority: High — add table-driven classification tests for mismatched extension/magic.

**Ignored DB-error paths:**
- What's not tested: The `_ =` write failures in `service.go` have no test asserting the observable outcome (missing history, duplicate question on retry).
- Files: `internal/app/service.go`, `internal/app/service_test.go`
- Risk: The documented duplicate-question trade-off (line 158-163) is untested and can regress silently.
- Priority: High — pair with the fix in Tech Debt.

**Frontend composer keyboard behaviour:**
- What's not tested: Enter vs Shift+Enter submission and Escape-to-cancel in the composer (explicitly deferred in `docs/08-roadmap-testing.md:150`).
- Files: `frontend/src/components/chat/ChatPanel.tsx` (`keyDown`)
- Risk: A keyboard-handling regression traps or mis-sends messages.
- Priority: Medium — add a component test using the existing `frontend/src/test/wails.ts` double.

**CI has never executed on GitHub:**
- What's not tested: Both `.github/workflows/ci.yml` and `release.yml` are unrun (repo not pushed, per ADR-040); the SignPath `sign` job is inactive until secrets exist.
- Files: `.github/workflows/ci.yml`, `.github/workflows/release.yml`
- Risk: The first push may reveal workflow bugs (path assumptions, `wails build` on runner, artifact upload).
- Priority: High before public release — validate one full pipeline run.

---

*Concerns audit: 2026-10-11*
