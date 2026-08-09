# AGENTS.md

Repository guidance for freshservice-label.

## Approach

- Stay within the requested scope and preserve unrelated local changes.
- This is a small internal webhook and label renderer, not a product platform.
- Simplify and modernize existing code before adding abstractions, compatibility layers, or speculative configuration.
- Follow the shared Woodstar tooling baseline without importing complexity the service doesn't need.

## Repository Map

- CLI, HTTP server, and preview command: `cmd/freshservice-label`
- Ticket decoding, rendering, printing, and HTTP behavior: `internal/ticketprinter`
- Local preview command and output: `cmd/preview` and `preview.png`

Keep the request-to-label path explicit. Don't split the single capability into generic transport, template, or job frameworks.

## Commands

Use Mise tasks as the repository contract.

- Dependencies: `mise run deps`
- Build: `mise run build`
- Tests: `mise run test`
- Lint: `mise run lint`; fixes: `mise run lint-fix`
- Format: `mise run format`; check: `mise run fmt-check`
- Preview: `mise run preview`
- Module and workflow checks: `mise run tidy-check`, `mise run workflow-lint`

## Engineering Rules

- Prefer concrete Go types, small consumer-owned interfaces, and wrapped errors.
- Bound request bodies and errors; don't log credentials or full ticket payloads.
- Keep label rendering deterministic and cover output changes with focused tests or preview evidence.
- Tests mustn't contact Freshservice or physical printers.

## Commits

- Use focused Conventional Commits.
- Don't push, publish images, or contact live services unless explicitly requested.
- Report checks run, skipped checks, and unresolved failures.
