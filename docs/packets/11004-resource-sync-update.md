# Packet 11004: resource sync update (SC_11004)

This document captures the current reverse-engineered understanding of packet ID `11004` for the EN client and how Belfast relates to it.

## Overview

- Direction: server -> client push.
- Purpose: refresh the client's local resource state from a list of `RESOURCE { type, num }` entries.
- Client expectation: packet handler reads `resource_list`, updates `Player` resources, and refreshes the last resource update timestamp.

## Directionality notes

The shipped EN client scripts register `11004` as an inbound handler. The EN client's send path relies on resolving a `cs_<id>` protobuf descriptor; if `cs_11004` is missing from the descriptor list, the client cannot construct and send `11004`.

In other words, a client -> server `CS_11004` request is unlikely for the EN scripts we reference; treat `11004` as a push (`SC_11004`) unless verified otherwise via captured traffic or a different client build.

## Protobuf payload

Belfast already has generated protobuf types for `SC_11004`.

- Type: `internal/protobuf.SC_11004`
- Fields:
  - `resource_list` (repeated): `[]*internal/protobuf.RESOURCE`
  - `RESOURCE` entries contain:
    - `type` (int32)
    - `num` (int32)

Server code reference: `internal/protobuf/SC_11004.pb.go`.

## Belfast current behavior

- There is no packet registry entry for `11004` (no handler is registered under `internal/entrypoint/packet_registry.go`).
- If the server receives an unregistered inbound packet id, dispatch logs a missing handler and still records the packet via debug capture.
  - Code: `internal/packets/handler.go`.

## Implementation-ready plan (future work)

If/when we need live resource UI updates without reconnects:

1) Treat `11004` as server -> client push only.
2) Add a small helper to build and send `protobuf.SC_11004` from the commander resource state, mirroring the existing resource list serialization used during player info/login.
   - Anchor: `internal/answer/player_info.go`.
3) Emit that helper once after any handler that mutates resources (e.g. `internal/answer/give_resources.go`).

Only if verified client evidence shows a real inbound `CS_11004` exists:

- Register a `11004` handler in `internal/entrypoint/packet_registry.go` that responds with `SC_11004` and has no side effects.

## Risks

- Over-sending `SC_11004` increases network chatter; prefer one emission per request (or only when resources change).
- Incorrect or incomplete resource lists may cause client UI drift; reuse the login/resource serialization logic to keep lists consistent.
