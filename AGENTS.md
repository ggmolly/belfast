# Belfast Agent Guide

This file is for agentic coding assistants working in this repository.
Follow repo conventions, keep diffs small, and avoid unrelated refactors.

## Repository Basics
- Language: Go (module `github.com/ggmolly/belfast`)
- Go version: 1.25.5 (see `go.mod` and CI)
- Web/API: Iris v12, Gorm, protobuf for packets
- Reverse-engineering project: never add or share `.proto` files
- Dependency updates are discouraged unless explicitly requested

## Repository Layout
- `cmd/belfast`: main server entrypoint
- `cmd/pcap_decode`: packet decode tool for pcap files
- `internal/answer`: packet handlers and responses
- `internal/api`: Iris REST API server
- `internal/connection`: TCP server/client plumbing
- `internal/orm`: database models and access
- `internal/packets`: packet registry/dispatch
- `internal/protobuf`: generated protobuf types
- `internal/misc`: shared helpers and update logic

## Build, Lint, and Test Commands
### Build / Run
- Build server: `go build ./cmd/belfast`
- Run server: `go run ./cmd/belfast --config server.toml`
- Dev hot reload (Air): `air` (uses `.air.toml`)

### Tests
- All tests: `go test ./...`
- Single package: `go test ./internal/packets`
- Single test by name: `go test ./internal/packets -run TestName`
- Single subtest: `go test ./internal/packets -run TestName/SubtestName`

### Lint / Format
- Format (required): `gofmt -w <file>.go`
- Imports: gofmt handles grouping; do not hand-align
- There is no dedicated linter configured; do not add one unless asked

### Code Generation
- Proto generation: `make proto`
- Lua-to-proto step only: `make lua-proto`
- Swagger docs: `make swag`

### CI
- GitHub Actions runs `go test ./...` on Go changes
- CI skips commits containing `chore` in the message

## Code Style Guidelines
### General Go Style
- Apply `gofmt` to every touched Go file
- Prefer standard library solutions before adding dependencies
- Keep changes focused; avoid large refactors or formatting churn
- Avoid defensive checks unless a boundary requires it
- Follow KISS/DRY; keep code straightforward and avoid duplication
- Comments only for non-obvious intent, invariants, or business rules

### Imports
- Use three groups separated by blank lines:
  1) stdlib
  2) third-party
  3) local `github.com/ggmolly/belfast/...`
- Use explicit imports; avoid dot or blank imports unless required

### Naming
- Use idiomatic Go names: `camelCase` for variables, `PascalCase` for exported
- Avoid abbreviations unless common (`cfg`, `ctx`, `id`)
- Packet handlers follow `Forge_SC12345` or descriptive names in `internal/answer`

### Types
- Prefer concrete types; avoid `interface{}` unless required by APIs
- Use pointers for protobuf messages and large structs to avoid copying
- Use slices/maps with clear element types; avoid `any` unless unavoidable

### Error Handling
- Return errors to the caller unless the layer owns the decision
- Log with `internal/logger` (`logger.LogEvent`, `logger.WithFields`) in server paths
- Keep error messages short and actionable; avoid wrapping with noisy context

### Logging
- Use structured fields where possible (`logger.FieldValue`, `logger.CommanderFields`)
- Keep log scopes stable so tooling can filter (e.g., `API/Start`)
- Do not log secrets or user tokens

### Formatting and Layout
- Keep functions reasonably sized; split when logic diverges
- Align struct literals using gofmt defaults
- Avoid blank lines in tight loops unless improving readability

### Protobuf / Packets
- Never add or commit `.proto` files
- Use `internal/protobuf` types for packet serialization
- Packet handlers are registered in `cmd/belfast/main.go:init`

### Database / ORM
- Use Gorm models as defined in `internal/orm`
- Keep DB access inside `internal/orm` and service layers
- Avoid raw SQL unless there is no Gorm equivalent
- Every DB object must be manageable through the REST API; implement SCRUD endpoints

### API / HTTP
- Iris handlers live under `internal/api`
- Keep API errors consistent with existing responses
- REST endpoints must include Swagger annotations

## Project Conventions
- Commit messages follow: `<type>(<component>): one-line summary`
- PRs should be small and single-purpose
- Refactors should be discussed before large changes
- Do not introduce new dependencies without explicit request

## Tests and Quality Bar
- Any behavior change should come with a test when feasible
- Prefer regression tests before implementing fixes
- MUST add tests for: new endpoints, packet handlers, ORM models/queries, and bug fixes that change behavior
- MUST cover failure paths for API/ORM changes (validation errors, not-found, permission checks)
- MUST update or replace broken tests instead of skipping or weakening assertions
- Keep tree green; run `go test ./...` for touched areas

## Tooling and Workflow Notes
- Config defaults live in `server.toml` / `server.example.toml`
- Environment template: `.env.example` (do not commit real `.env` files)
- Scripts: `tools/import.sh`, `tools/get_ship_icons.py`
- Packet decoding tool: `go run ./cmd/pcap_decode --help`

## Scripts and Utilities
- Packet capture helper: `./capture.sh <iface> <pcap>` (uses `wg/hosts`)
- Pcap decoder expects protobuf registry from `internal/protobuf`
- Swagger generation uses `cmd/belfast/main.go` as entrypoint

## Configuration and Runtime
- Server flags live in `cmd/belfast/main.go` (`--config`, `--no-api`, `--reseed`)
- Region defaults are resolved via `internal/region` and config values
- Database setup happens through `internal/orm` initialization

## Contribution Notes for Agents
- Keep commits focused on a single logical change
- Avoid touch-ups to unrelated files or formatting
- If unsure about behavior, read related handlers in `internal/answer`

## Behavior Guidance
- Preserve existing packet IDs and handler wiring unless the change requires it
- Prefer updating data through existing helpers in `internal/misc`
- Use config defaults rather than hardcoding regions or ports

## Safety and Legal
- Reverse engineering project: do not add copyrighted game assets
- Never check in secrets, API keys, or `.env` files
