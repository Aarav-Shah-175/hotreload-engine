# hotreload

`hotreload` is a Go CLI that watches a project tree, rebuilds on change, restarts the server, and streams logs in real time.

## Planned Architecture

- `cmd/hotreload`: CLI entrypoint and wiring.
- `internal/config`: flag parsing and validation.
- `internal/engine`: orchestration loop (debounce, build cancelation, restarts).
- `internal/watcher`: recursive `fsnotify` watcher with ignore rules.
- `internal/build`: build command runner with context cancelation.
- `internal/process`: server process lifecycle and process-tree termination.
- `internal/logging`: centralized `log/slog` setup.

## Status

Scaffolded. Core implementation is added incrementally in follow-up commits.
