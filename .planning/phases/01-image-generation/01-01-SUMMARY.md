# Phase 01 — Image Generation: Execution Summary

**Status:** Complete — all 6 tasks executed and committed. Automated gates pass.
**Requirement:** FR-013 · **ADR:** ADR-055
**Plan:** `01-01-PLAN.md`

## What was built

A vertical MVP slice for image generation in Wortel AI: configure an image-capable model
per provider, type a prompt in the chat panel, the Go backend calls Kenari's
`POST /v1/images/generations`, decodes whichever payload shape the provider returns
(`b64_json` or `url`/data-URL), persists the bytes locally, and renders the result inside the
conversation — with a distinct loading state, cancel, safe error messages, and PAYG billing
honesty.

**Architecture (as planned):** generated images are stored as a special assistant message row
(`messages.image_mime` / `image_data`, migration v6) and served to the WebView through the
existing `/attachments/{id}` route — no new panel or navigation. `RecentContext` excludes
empty-content assistant rows, so a generated image is never re-sent to a chat model (no double
billing). Deleting a conversation cascade-deletes its images via FK.

## Tasks

1. **Tracer — b64 happy path** (`c29413a`): `internal/ai/images.go` (GenerateImage, normalizeImagesURL,
   MIME allowlist), `internal/app/images.go` (GenerateImage/CancelImage, image:event, separate cancel
   map, 5-minute timeout), storage migration v6 + SetMessageImage/MessageImage/AppendImageMessage,
   asset-route fallback, frontend `useImageGen`/`ImagePromptBar`/MessageItem image render.
2. **URL/data-URL + timeout** (`a228405`): fetchImageURL (data URLs in place; http(s) via a fresh
   `SafeClient` with no provider key, 25 MB cap, MIME allowlist); long-timeout assertion.
3. **Error mapping** (`52ca2cf`): 402→insufficient_balance, 400 naming model→model_not_found,
   429/413/503 shared, unsupported MIME→unsupported_image; no body/key leakage; no auto-retry.
4. **Image model config + UX** (`8c1408e`): `provider_profiles.active_image_model` (migration v7),
   `ResolveImage`, `SetActiveImageModel`, settings "Gambar" radio, prompt-bar UX; no hardcoded ids.
5. **Lifecycle/isolation** (`f282d14`): CancelAll stops image goroutines, cascade delete, no-provider
   clean failure, single-flight per conversation, hook stale-event drop + unmount cleanup.
6. **Docs** (`330f6fe`): FR-013, ADR-055, 05/07/10 docs, README; regenerated wails bindings.

## Verification run (all passed)

- `gofmt -l app.go main.go internal/` → clean
- `go vet ./...` → clean
- `go test ./internal/... .` → all packages ok
- `CGO_ENABLED=1 go test -race ./internal/app/ ./internal/storage/` → ok
- `cd frontend && npx tsc --noEmit` → clean
- `cd frontend && npm test` → 103 tests pass
- `cd frontend && npm run build` → built
- `wails build -ldflags "-X main.Version=0.1.0"` → `build/bin/companion.exe`

## Manual UAT (not run in this session — requires a real Kenari image model)

1. Set a live image model in Settings → prompt → image renders via `/attachments/{id}`.
2. A URL-returning model → backend fetches/stores locally; no WebView request to the provider host.
3. Cancel (button and Escape) mid-flight → idle, no phantom image, composer unlocked.
4. Empty PAYG balance or bad model → specific message (insufficient_balance / model_not_found), no leak.
5. Restart → image still renders from the DB.
6. Delete conversation → image row gone, route 404s.

## Deferred / out of scope

- Image editing (`/v1/images/edits`), cross-conversation gallery, `?modality=` model picker,
  image downscale. See ADR-055 risks.

## Note on project state

`.planning/` has no ROADMAP.md/STATE.md/PROJECT.md — the repo was not bootstrapped via
`/gsd-new-project`, so this phase ran standalone (no roadmap tracking). Consider bootstrapping
GSD to track future phases.
