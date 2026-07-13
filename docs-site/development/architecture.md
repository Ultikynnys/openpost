# Architecture

## Frontend

- SvelteKit
- TailwindCSS
- Paraglide
- Vitest
- Bun

## Backend

- Go
- Echo
- Huma
- SQLite by default, Postgres for cloud deployments
- Bun ORM

HTTP routes are defined with Huma whenever they are part of the typed product API. Echo remains the transport adapter and owns the small number of routes that are not JSON API operations, such as multipart uploads, public media/avatar serving, OAuth/MCP protocol endpoints, and the embedded SPA.

Handlers authenticate and validate request boundaries, services own product rules and provider orchestration, and Bun-backed database packages own persistence. Provider API behavior stays in `internal/platform`; provider selection and public-media behavior come from adapter maps and the central capability catalog.

## Background jobs

Publishing and other durable work flows through a database-backed jobs table.

## Media

Media uses the `BlobStorage` abstraction with local filesystem storage by default and S3-compatible storage for cloud deployments.

## Deployment

The built frontend is embedded into the Go binary for single-binary deployment.

## Client surfaces

The web app, CLI, MCP server, and direct HTTP clients share the same backend authorization, validation, quotas, and audit records. They intentionally differ in interaction design. See [Product Surface Parity](/reference/surface-parity) for the supported workflow matrix.
