---
last_mapped_commit: b55c8e5bb1c4199b24162ca36f0c612ef255a01c
last_mapped_at: 2026-10-11
---
# Technology Stack

**Analysis Date:** 2026-10-11

## Languages

**Primary:**
- Go 1.25.0 (module `companion`; toolchain and CI pin 1.26.4) - Backend: AI request orchestration, credential storage, SQLite, Win32 integrations. Entry point `main.go`, bindings in `app.go`, all packages under `internal/`.
- TypeScript 5.6.3 - Frontend React UI. Sources under `frontend/src/`, typed bindings in `frontend/wailsjs/`.

**Secondary:**
- CSS (plain, design tokens in OKLCH) - `frontend/src/styles/` (`tokens.css` plus per-area surfaces `shell.css`, `chat.css`, `pet.css`, `settings.css`, `markdown.css`). Tailwind was removed; only plain CSS with custom properties is used.
- SQL (SQLite DDL) - Schema and migrations in `internal/storage/db.go`.
- PowerShell / YAML - CI and release pipelines in `.github/workflows/ci.yml` and `.github/workflows/release.yml`.

## Runtime

**Environment:**
- Windows 10/11 x64 is the only supported target. The app relies on Win32 APIs (Credential Manager, registry, user32/shell32) and WebView2.
- Desktop shell: Wails v2.16.0, which embeds the React frontend in a single Go binary via WebView2.
- WebView2 Runtime is required at run time; the app shows a friendly error when it is absent (`main.go` -> `fatalDialog`).

**Package Manager:**
- Go modules (`go.mod`, `go.sum`) for the backend.
- npm (lockfile `frontend/package-lock.json`) for the frontend. The root `package.json` only holds the `opencode-gpt-image-plugin` dev dependency (tooling, not shipped).
- Lockfiles: present (`go.sum`, `frontend/package-lock.json`, root `package-lock.json`).

## Frameworks

**Core:**
- Wails v2.16.0 (`github.com/wailsapp/wails/v2`) - Desktop shell, native window options, asset server, event bus, and Go<->JS bindings. Config: `wails.json`, options wired in `main.go`.
- React 19.1.0 (`react`, `react-dom`) - UI rendering. Entry `frontend/src/main.tsx`, root `frontend/src/App.tsx`.
- Vite 7.3.7 - Frontend build/dev server. Config `frontend/vite.config.ts` (React plugin plus a custom plugin that 404s `/attachments/` in dev so Wails falls back to Go).

**Testing:**
- Go standard `testing` package - backend unit tests across `internal/**/*_test.go`. Run with `go test ./...` and `CGO_ENABLED=1 go test -race ./...`.
- Vitest 3.2.7 with jsdom 30.1.2 - frontend unit tests. Config `frontend/vitest.config.ts` (separate from the app build). Run with `npm test`.
- @testing-library/react 16.3.3 and @testing-library/dom 10.4.2 - component/hook testing.

**Build/Dev:**
- Wails CLI v2.16.0 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`) - `wails dev` and `wails build`.
- TypeScript compiler - `tsc` runs first in the frontend `build` script for typecheck.
- `gofmt`, `go vet` - formatting/lint gates in CI.

## Key Dependencies

**Critical:**
- `github.com/wailsapp/wails/v2` v2.16.0 - The whole desktop runtime; v2 is pinned, v3 beta is explicitly disallowed (README).
- `github.com/zalando/go-keyring` v0.2.6 - Stores API keys in the OS credential manager (Windows Credential Manager via `github.com/danieljoos/wincred`). Used in `internal/credentials/osstore.go`.
- `modernc.org/sqlite` v1.46.0 - Pure-Go SQLite driver (no CGO for the DB). Used in `internal/storage/db.go`; data at `%LOCALAPPDATA%\WortelAI\wortel.db`.
- `github.com/google/uuid` v1.6.0 - IDs for providers, conversations, messages, attachments, requests (`internal/app/service.go`, `internal/attach/staging.go`).
- `golang.org/x/sys` v0.46.0 - Win32 syscalls for tray, hotkey, window placement, and autostart (`internal/tray`, `internal/hotkey`, `internal/window`, `internal/autostart`).
- `react-markdown` ^10.1.0 + `remark-gfm` ^4.0.1 + `rehype-sanitize` ^6.0.0 - Renders untrusted AI Markdown safely, no raw HTML. Used in `frontend/src/components/chat/SafeMarkdown.tsx`.

**Infrastructure:**
- `github.com/wailsapp/go-webview2` v1.0.22, `github.com/jchv/go-winloader`, `github.com/go-ole/go-ole` - Wails native Windows backends (indirect).
- `git.sr.ht/~jackmordaunt/go-toast/v2` v2.0.3, `github.com/bep/debounce` - Wails runtime helpers (indirect).
- `golang.org/x/crypto` v0.53.0 - Transitive via Wails (indirect).

## Configuration

**Environment:**
- No `.env` usage; the app is local-first with no server-side config. All user configuration lives in SQLite (`app_settings` table) and the OS credential store.
- User data location: `%LOCALAPPDATA%\WortelAI\` (`internal/storage/db.go` `DefaultPath`), falling back to `os.UserCacheDir()`.
- Version stamp: `-ldflags "-X main.Version=..."` (see `main.go`, `app.go` `Version()`); defaults to `dev`.

**Build:**
- `wails.json` - App name/output, frontend install/build/dev commands, product metadata (`info.companyName` = "Wortel AI", version 0.1.0).
- `frontend/vite.config.ts` - Vite build and dev-server plugin.
- `frontend/vitest.config.ts` - Test-only config (jsdom, `src/**/*.test.{ts,tsx}`).
- `frontend/tsconfig.json` (+ `tsconfig.node.json`) - Strict TS, `moduleResolution: Node`, `jsx: react-jsx`.
- `build/windows/info.json` - Windows exe string-table metadata templated from `wails.json`.
- `.editorconfig`, `.gitattributes` - Cross-editor formatting/line-ending rules.

## Platform Requirements

**Development:**
- Go 1.26.4 (Wails requires Go 1.21+).
- Node.js 22 + npm.
- Wails CLI v2.16.0.
- WebView2 Runtime for running the app.
- MinGW-w64 (gcc) on PATH for `go test -race` (CGO required only for race builds; the DB driver itself is pure Go).

**Production:**
- Single portable Windows executable `build/bin/companion.exe`; frontend assets are embedded, so end users need neither Go nor Node.js and there is no installer.
- Distribution via GitHub Releases (`.github/workflows/release.yml`), optionally code-signed through SignPath Foundation; unsigned builds trigger Windows SmartScreen.

---

*Stack analysis: 2026-10-11*
