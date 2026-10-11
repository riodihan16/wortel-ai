---
last_mapped_commit: b55c8e5bb1c4199b24162ca36f0c612ef255a01c
last_mapped_at: 2026-10-11
---
# External Integrations

**Analysis Date:** 2026-10-11

## APIs & External Services

**AI Chat Providers (OpenAI-compatible Chat Completions):**
- The app is a generic client for any OpenAI-compatible `/chat/completions` endpoint; no single provider is hardcoded. The live client is `ai.OpenAICompatible` in `internal/ai/provider.go`; the request path is normalized by `ai.NormalizeBaseURL` (forces `http`/`https` scheme, rejects embedded credentials, appends `/chat/completions` if missing).
  - SDK/Client: Standard Go `net/http` (no vendor SDK). HTTP client built by `ai.SafeClient()` (`internal/ai/http.go`) with a redirect policy that strips `Authorization` when a redirect crosses host.
  - Auth: `Authorization: Bearer <API_KEY>` per request (`internal/ai/provider.go`). The key is passed per call and never persisted by the AI client.
  - Streaming: Server-Sent Events parsed incrementally by `ai.ScanSSE` (`internal/ai/sse.go`); `[DONE]` or EOF ends the stream; malformed frames are skipped, not fatal.
  - Cancellation: `context.Context` per generation, stored per conversation in `internal/app/service.go`.
  - Presets offered in the settings UI (`frontend/src/components/settings/presets.ts`):
    - OpenAI - `https://api.openai.com/v1` (key `sk-…`)
    - OpenRouter - `https://openrouter.ai/api/v1` (key `sk-or-…`)
    - Groq - `https://api.groq.com/openai/v1` (key `gsk_…`)
    - LM Studio (local) - `http://localhost:1234/v1`
    - Ollama (local) - `http://localhost:11434/v1`
  - Presets only prefill an editable form; nothing is stored until the user clicks "Tes koneksi" or "Simpan".

**Provider Connectivity Probe:**
- `Providers.Test` (`internal/app/providers.go`) sends a minimal non-streaming request (`"stream": false`, `"max_tokens": 1`, message `"ping"`) with a 15s timeout to validate URL/model/credential before saving.

**Error Contract from Providers:**
- HTTP status codes are mapped to safe, actionable `PublicError` values in `internal/ai/errors.go` (400, 401, 403, 404, 429, 408, 5xx, 413, plus 400s naming image/file types -> `unsupported_input`). `Retry-After` is honored on 429/503 (capped at 600s). Response bodies are read up to 4 KB only to classify errors and are never logged or forwarded.

**Mock Provider (dev/test only):**
- `ai.MockProvider` (`internal/ai/provider.go`) emits canned tokens with delay, no network and no secrets. It is the fallback when storage/credentials are not wired (`app.go` `startup`).

**No other outbound services:** no telemetry, analytics, account system, cloud sync, or update check. A repository-wide scan of `internal/` found HTTP usage only in `internal/ai` (provider transport), `internal/app` (attachment asset handler + provider probe), and `internal/attach` (`http.DetectContentType` for MIME sniffing) — no external SaaS calls.

## Data Storage

**Databases:**
- SQLite (local file), single database at `%LOCALAPPDATA%\WortelAI\wortel.db` (fallback `os.UserCacheDir()`), resolved by `storage.DefaultPath()` in `internal/storage/db.go`.
  - Connection: opened by `storage.Open()` with `_pragma=foreign_keys(1)`, `busy_timeout(5000)`, `journal_mode(WAL)`.
  - Client: `modernc.org/sqlite` (pure Go, no CGO) via `database/sql`.
  - Schema version: `currentVersion = 5`; sequential migrations applied in transactions (`schema_version` table). Tables: `provider_profiles`, `app_settings`, `conversations`, `messages`, `attachments`, indexes `idx_messages_conv` and `idx_attachments_msg`.
  - No secrets are stored here; `provider_profiles` has no secret column, only `credential_ref`.

**File Storage:**
- Attachments (images, PDFs, text) are stored as BLOBs in the `attachments` table (`internal/storage/db.go` v5, `internal/storage/attachments.go`), not on the filesystem. Metadata reads never select the `data` column.
- Frontend fonts are bundled static assets (`frontend/src/assets/fonts/`, Geist / Geist Mono, OFL licensed).

**Caching:**
- None. WAL mode is used for SQLite concurrency; there is no application-level cache. The app explicitly avoids polling (idle-light principle, README).

## Authentication & Identity

**Auth Provider:**
- Custom / none. There is no account system, login, or identity provider. The only credential is the user's own AI provider API key.
- Secrets are stored in the **Windows Credential Manager** via `credentials.OSStore` (`internal/credentials/osstore.go`) using `github.com/zalando/go-keyring`. Service name `WortelAI`; account names are `wortel-provider-<uuid>`.
- The frontend never receives secret bytes: `config.ProviderProfile` exposes only `HasCredential bool` (`internal/config/providers.go`, `internal/storage/providers.go` `ToPublic`).
- `credentials.MapStore` (`internal/credentials/mapstore.go`) is an in-memory store for tests and non-Windows dev only; production on Windows must use `OSStore`.
- Secret hygiene: `Providers.Resolve` wipes the credential byte slice after use (`internal/app/providers.go`); API keys are never written to SQLite, logs, or events.

## Monitoring & Observability

**Error Tracking:**
- None. No Sentry/Datadog/telemetry integration by design (AGENTS.md prohibits telemetry).

**Logs:**
- Wails runtime logging only (`runtime.LogWarningf`) for non-fatal conditions such as tray/edge-guard/activation failures (`app.go`). Errors that could reach the UI are mapped to safe `PublicError` values that never include secrets or raw provider bodies.

## CI/CD & Deployment

**Hosting:**
- GitHub Releases. `companion.exe` (portable, Windows 10/11 x64) plus `SHA256SUMS.txt` are published as draft releases (`.github/workflows/release.yml`).

**CI Pipeline:**
- `.github/workflows/ci.yml` (on push/PR to `master`): backend job on `windows-latest` (`gofmt`, `go vet`, `go test`, `go test -race` with `CGO_ENABLED=1`); frontend job on `ubuntu-latest` (`npm ci`, `tsc --noEmit`, `npm test`, `npm run build`); release-build job producing `build/bin/companion.exe`.
- `.github/workflows/release.yml` (on `v*` tags): builds the unsigned exe, optionally submits it to **SignPath Foundation** for code signing (gated by an `environment: release` approval and `vars.SIGNPATH_ENABLED == 'true'`), then publishes signed or unsigned artifact with a SHA-256.
- Environment pins across both workflows: `GO_VERSION 1.26.4`, `NODE_VERSION 22`, `WAILS_VERSION v2.16.0`.

**Code Signing (external service):**
- SignPath (external) via `signpath/github-action-submit-signing-request@v3`. Requires repo secrets `SIGNPATH_API_TOKEN`, `SIGNPATH_ORGANIZATION_ID`, `SIGNPATH_PROJECT_SLUG`, `SIGNPATH_POLICY_SLUG`, `SIGNPATH_ARTIFACT_CONFIGURATION_SLUG`, the `release` environment, and the SignPath GitHub App. Disabled until configured; unsigned builds ship without it.

## Environment Configuration

**Required env vars:**
- None required at run time. The app reads `LOCALAPPDATA` (with `os.UserCacheDir()` fallback) for the data path. Provider API keys are entered in the UI, not via environment variables.

**Secrets location:**
- Windows Credential Manager (OS-level), not in the repo, SQLite, or logs. Repository secrets used only by CI/CD: `SIGNPATH_*` (see above). No `.env` files exist.

## Windows-Specific OS Integrations

Although not network integrations, these are the OS surfaces the app binds to (all via `golang.org/x/sys/windows`):
- **Registry (autostart):** per-user `HKCU` Run key; enable/disable via `internal/autostart/autostart_windows.go`. No admin rights, no secret.
- **Global hotkey:** `RegisterHotKey`/`GetMessageW` on a locked OS thread (`internal/hotkey/hotkey_windows.go`); default combo from `internal/hotkey/combo.go`.
- **System tray:** `Shell_NotifyIconW` with a hidden window + popup menu (`internal/tray/tray_windows.go`); re-adds the icon when Explorer restarts.
- **Window placement/DPI:** `GetWindowRect`/`MonitorFromWindow`/`GetMonitorInfo`/`GetDpiForWindow`/`SetWindowPos` (`internal/window/place_windows.go`).
- **Local HTTP asset endpoint:** the Go backend serves `GET /attachments/{uuid}` to the WebView through the Wails asset server (`internal/app/attachment_handler.go`, wired in `main.go`). Only well-formed UUIDs are served; responses set `X-Content-Type-Options: nosniff`, `Cache-Control: private, max-age=3600`, and `Content-Disposition: inline`.
- **WebView2:** the embedded browser runtime Wails uses to render React.

## Webhooks & Callbacks

**Incoming:**
- None. There is no public HTTP server, no webhook receiver, and no inbound network listener. The only local HTTP route is the WebView attachment asset handler noted above.

**Outgoing:**
- None. No outbound webhooks or callbacks; the only outbound network traffic is the user-initiated AI provider request.

---

*Integration audit: 2026-10-11*
