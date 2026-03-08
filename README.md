# hotreload

`hotreload` is a Go CLI that watches project files recursively, rebuilds on change, and restarts the server while streaming logs in real time.

## CLI

```bash
go run ./cmd/hotreload \
  --root ./testserver \
  --build "go build -o ./bin/server ./cmd/server" \
  --exec "./bin/server"
```

Flags:

- `--root`: directory to watch recursively.
- `--build`: build command to execute.
- `--exec`: command to run after a successful build.
- `--debounce`: debounce window for bursty file events (default `300ms`).

## Architecture

### 1. `cmd/hotreload`

Thin entrypoint: parse config, initialize `slog`, and run engine.

### 2. `internal/engine`

The orchestrator:

- starts recursive watcher
- enqueues first build on startup
- debounces change events
- cancels an in-flight build/reload when newer changes arrive
- stops previous server before launching the new one

### 3. `internal/watcher`

`fsnotify`-based recursive watcher:

- adds watches for existing subdirectories
- adds new directories dynamically when created
- filters ignored paths and temp files
- emits normalized change events

Ignored by default:

- `.git`
- `node_modules`
- `bin`, `build`, `dist`, `tmp`
- common temporary editor files (`*.swp`, `*.swx`, `*.tmp`, `*~`, `.#*`)

### 4. `internal/build`

Build command runner using `exec.CommandContext`:

- shell command execution (`cmd /C` on Windows, `sh -c` on Unix)
- cancellation via context when superseded by new changes
- line-by-line log streaming through `log/slog`

### 5. `internal/process`

Server process lifecycle:

- start command and stream stdout/stderr live
- terminate previous process tree on restart
- on Unix, command runs in its own process group and is terminated via group signal
- on Windows, uses `taskkill /T` (and `/F` fallback) to kill parent + children

## Demo

A sample project is included at `testserver/`.

Run demo with Make:

```bash
make run-demo
```

This starts `hotreload` watching `testserver`, rebuilding `testserver/bin/server`, and running it on port `8080`.

## Make targets

- `make run-demo` - run hotreload against demo server
- `make build-hotreload` - build the hotreload binary into `./bin/hotreload`
- `make clean` - remove build artifacts
