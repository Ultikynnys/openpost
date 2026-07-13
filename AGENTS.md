# OpenPost AI Agents & Development Guidelines

This document serves as a guideline for autonomous AI agents (like Copilot, Cursor, Codeium, or CLI agents) and human developers contributing to **OpenPost**. It outlines the core tech stack, architectural rules, and specific instructions for AI behavior.

Treat this as a living file: add concise repo learnings when they would save future agents time, and trim entries that become stale or too obvious.

## 1. Core Architecture & Tech Stack

**Frontend:**
- **Framework:** SvelteKit (using `@sveltejs/adapter-static` for simple SPA deployment).
- **Styling:** TailwindCSS.
- **i18n:** Paraglide.
- **Testing:** Vitest.
- **Package Manager:** pnpm.

**Backend:**
- **Language:** Go (1.25+).
- **Framework:** Echo (`github.com/labstack/echo/v4`). Huma for OpenAPI spec generation.
- **Database:** SQLite by default; Postgres is supported and required for hosted cloud mode.
- **ORM:** Bun (`github.com/uptrace/bun`).
- **Background Jobs:** Custom database-backed polling worker using the `jobs` table (no external Redis dependency).
- **Media Storage:** Local filesystem via `BlobStorage` interface (configurable via `OPENPOST_MEDIA_PATH`).

**Deployment:**
- **Strategy:** Single Go binary. SvelteKit's static output is embedded directly into the Go executable using `go:embed`.
- **Hosted app:** `app.openpost.social` runs from the Docker image published by the GitHub `Build and Release` workflow and is pinned in `~/.config/home/hosts/rgo-vps/default.nix`.

## 2. Platform Adapter Architecture

All social platform integrations follow a unified `PlatformAdapter` interface defined in `internal/platform/adapter.go`. Each platform implements this interface in its own file within `internal/platform/`:

| File | Platform | Auth Method |
|------|----------|-------------|
| `x.go` | Twitter/X | OAuth 1.0a |
| `mastodon.go` | Mastodon | OAuth 2.0 (per-instance) |
| `bluesky.go` | Bluesky | App Passwords |
| `linkedin.go` | LinkedIn | OAuth 2.0 |
| `threads.go` | Threads | Meta OAuth 2.0 |

Adapters are registered in `main.go` via a `map[string]platform.PlatformAdapter` and passed to the token manager, publisher, and OAuth handler. **No switch statements** — everything uses map lookups.

Shared HTTP helpers are in `internal/platform/http.go`:
- `DoRequest` — generic HTTP request with error handling
- `DoJSON` — JSON marshaled request
- `DoMultipart` — multipart file upload
- `DoFormURLEncoded` — form-encoded request

## 3. Agent Guidelines & Coding Mandates

When an AI agent is invoked to assist with this repository, it MUST adhere to the following rules:

### A. Commit & Branch Conventions
- **Always use Conventional Commits** (e.g., `feat:`, `fix:`, `chore:`, `refactor:`, `docs:`). Follow https://www.conventionalcommits.org/
- **Always use Conventional Branches** (e.g., `feature/add-login`, `fix/header-alignment`, `hotfix/emergency-patch`)
- **Always update the Changelog** for any major features, bug fixes, or breaking changes. Use the `## [Unreleased]` section to document changes since the last release.
- **Pre-push lint gate is mandatory.** The repo installs a `pre-push` git hook (via `devenv` on shell entry, see `devenv.nix:enterShell` and `scripts/pre-push-lint.sh`) that runs a fast local lint subset: backend format check, backend lint, and frontend lint. Pre-commit is intentionally cheap and mostly formatting/light lint. **Run `devenv shell -- lint` before release tags or high-risk changes** because full frontend type checks, tests, and production builds are intentionally not run on every push. Bypass with `OPENPOST_SKIP_PRE_PUSH_LINT=1 git push ...` only when CI or a manual full lint run has already gated the commit.
- **Production release flow:** for normal “commit, push, release, deploy prod” requests, run `pnpm release:prod "<conventional commit message>"` and let `scripts/release-prod.sh` stage all OpenPost changes, push `main`, create the next patch tag, wait for GitHub `Build and Release`, pull/restart the latest image on `rgo-vps`, and verify readiness. If a tag fails, do not retag it; fix forward with the next semver patch tag. Only fall back to the manual flow when the script itself is broken or the deployment target has drifted.
- **Go Backend:** Use Echo for HTTP handlers and Huma for OpenAPI endpoints. Follow the dependency injection pattern in `main.go`. Maintain separation of concerns: Handlers -> Services -> Database.
- **Platform Adapters:** Implement `PlatformAdapter` interface. Never put platform logic outside the `internal/platform/` package. Use shared HTTP helpers from `http.go`.
- **SvelteKit Frontend:** Always use standard Svelte 5 runes (`$state`, `$derived`, `$effect`, `$props`, `$bindable`). Use `+page.svelte`/`+page.ts` structures. Use the openapi-fetch typed client against `/api/v1` routes.
- **ORM Patterns:** Always use `github.com/uptrace/bun` for database operations. Do not write raw SQL strings unless doing complex SQLite pragmas or advanced queue polling.

### B. State Management & Single Binary Constraints
- **Filesystem Constraints:** OpenPost is meant to be highly portable. Local file storage (e.g., SQLite DB file, local media uploads) should be configurable via environment variables (e.g., `OPENPOST_DATABASE_PATH`, `OPENPOST_MEDIA_PATH`).
- **Asset Embedding:** Do not modify the SvelteKit build pipeline in a way that breaks `adapter-static`. The backend relies on a static `build/` directory to embed into the binary.

### C. Security & Credentials
- Tokens for social accounts (Access Tokens, Refresh Tokens) MUST ALWAYS be encrypted at rest using the `TokenEncryptor` service (AES-256-GCM).
- Do NOT hardcode cryptographic secrets in the codebase. Always load from environment variables (e.g., `OPENPOST_ENCRYPTION_KEY`, `OPENPOST_JWT_SECRET`).

### D. Workflow for Feature Implementation
1. **Model First:** If a feature requires data, update the `models.go` and `database.go` schema creation first.
2. **Backend Logic:** Implement the Service and the Echo API Handler.
3. **Frontend Implementation:** Write Svelte components and SvelteKit routes to interact with the new endpoint.
4. **Queue (if applicable):** If the action is async (e.g., publishing a post), insert a payload into the `jobs` table instead of blocking the HTTP request.

## 4. Prompts & Agent Commands (For Quick Context)

*If you are an agent, read these context hints before performing actions:*

- **"Add a new social platform"**: Create a new file in `internal/platform/` implementing `PlatformAdapter`. Register it in `main.go` under the provider map. Add platform icon to frontend's `compose-post.svelte`. Update the accounts page (`/accounts`) with connect UI.
- **"Modify database schema"**: Update `internal/models/models.go` struct fields with appropriate bun tags. Since we rely on `.IfNotExists()` in `database.go` currently, provide migration steps or table alter scripts if the table already exists. For new tables, follow the pattern in `internal/database/migrations/NNN_<name>.sql` and add a corresponding `internal/database/migrations/<name>_test.go` regression test.
- **"Handle thread drafts"**: Thread drafts (the in-progress state of a multi-post thread) live in the `thread_drafts` table — one row per parent post, keyed by `post_id` with `ON DELETE CASCADE`. The encoded draft JSON is the same `__openpost_thread__:` blob the frontend has always used; the backend now stores it in its own column instead of smearing it into `posts.content`. The composer sends/reads it through the typed `thread_draft` field on the post create/update/get endpoints. The legacy `posts.content` blob is still accepted on input and migrated on write, and is still readable as a fallback, so existing drafts survive the migration.
- **"Create a background job"**: Do not use `goroutine` blindly for tasks that must survive server restarts. Insert a row into the `models.Job` table so the `BackgroundWorker` can pick it up.
- **"Handle media uploads"**: Use the `BlobStorage` interface for file storage. The publisher fetches media from disk via `os.ReadFile()` and passes to `adapter.UploadMedia()`. For Threads, media must be served at a publicly accessible URL.
- **"Implement threading"**: Use `Post.ParentPostID` and `Post.ThreadSequence`. The publisher detects thread chains and publishes sequentially. Each adapter's `Publish` method handles `ReplyToID` platform-specifically.
- **"Provider app admin UI"**: The in-app Instance Admin provider-app page was removed. Hosted/operator credentials should be configured through environment/Nix secrets or `OPENPOST_PROVIDER_APPS`; backend admin provider-app APIs still exist for operator tooling.
- **"Modify MCP tools"**: Keep the model-facing catalog compact. Add operation descriptors to `mcpOperationCatalog()` so `search` can reveal their schemas on demand, route execution through `callMCPOperation()`, and only add a tool to `mcpAdvertisedTools()` when clients must see its descriptor up front (for example, Apps UI metadata). Preserve authorization, workspace scoping, validation, quota, and tool-call auditing in the delegated operation path.

## 5. Media & Threading Per Platform

| Platform | Media Upload | Threading |
|----------|-------------|-----------|
| X/Twitter | `POST /2/media/upload` (chunked for video) | `reply.in_reply_to_tweet_id` |
| Mastodon | `POST /api/v2/media` (async poll for large files) | `in_reply_to_id` |
| Bluesky | `com.atproto.repo.uploadBlob` (raw binary) | `reply: {root, parent}` with uri+cid JSON |
| LinkedIn | Vector Assets API (register→PUT→URN) | Comments API (`/socialActions/{urn}/comments`) |
| Threads | Public URL in `image_url`/`video_url` | `reply_to_id` |

## 6. Provider Key Convention

Provider keys in the `providers` map follow specific formats:

| Platform | Provider Key Format | Example |
|----------|---------------------|---------|
| X/Twitter | `"x"` | `"x"` |
| Mastodon | `"mastodon:" + server.InstanceURL` | `"mastodon:https://masto.pt"` |
| Bluesky | `"bluesky"` | `"bluesky"` |
| LinkedIn | `"linkedin"` | `"linkedin"` |
| Threads | `"threads"` | `"threads"` |

**Important:** For Mastodon, the `instanceURL` stored in `SocialAccount.InstanceURL` must match exactly with the key used to register the adapter. The canonical adapter key is `"mastodon:" + server.InstanceURL` (the full URL from config, e.g., `https://masto.pt`). When looking up the provider, use `"mastodon:" + account.InstanceURL` without modification. Human-friendly server names may still be used as OAuth selection labels, but not as the persisted provider key.
