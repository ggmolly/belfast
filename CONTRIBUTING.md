# Contributing to Belfast

Thanks for considering contributing to Belfast. Issues are welcome, PRs even more.

This is a reverse engineering project. No copyrighted material (or material that could be copyrighted) may be submitted in issues or PRs. `.proto` files must not be shared or pushed.

## Pull requests

- Fork the repo and open PRs from your fork.
- Keep PRs small and single-purpose.
- Run `gofmt` on any Go files you touch.
- Run `go test ./...` to make sure nothing regresses.
- Refactors should be discussed first.
- Dependency updates are discouraged unless explicitly requested.
- Reviews are required before merge.
- Linking an issue is optional but recommended.

## Commit messages

Use the following format:

```
<type>(<optional-component>): one-line summary
```

Examples:

```
feat(auth): add token validation
fix: handle nil commander
chore(proto): update generator
```

## Local setup

There are no tutorials for running the server locally, and local setup support is not provided.
