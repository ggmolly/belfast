# sqlc scaffolding

This directory is the starting point for the Postgres-only sqlc rewrite.

- `internal/db/migrations`: schema / migrations sqlc reads as input
- `internal/db/queries`: hand-written queries sqlc reads as input
- `internal/db/gen`: generated Go code output

To (re)generate the sqlc package locally:

`make sqlc`
