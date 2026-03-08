# hotreload

`hotreload` is a Go CLI that watches a project recursively, rebuilds it on source changes, and restarts the server process with real-time logs.

## Overview of hotreload

The tool is designed for local backend development loops:

1. Start `hotreload` once.
2. Edit source files.
3. On change, `hotreload` rebuilds and restarts your server automatically.

Core behavior implemented in this repository:

- recursive directory watching
- startup build trigger (no initial file save required)
- batched and debounced change handling
- cancellation of stale rebuild cycles
- process-tree termination across platforms
- structured lifecycle logging with `log/slog`
- crash-loop protection for unstable servers

## Architecture Diagram (ASCII)

```text
+---------------------+
|   cmd/hotreload     |
| flags + slog setup  |
+----------+----------+
           |
           v
+---------------------+        fsnotify events         +----------------------+
|  internal/watcher   | -----------------------------> |    internal/engine   |
| recursive watch +   |                                | debounce + orchestration
| path filtering      | <----------------------------- | cycle cancellation   |
+---------------------+    watch new dirs dynamically  +----------+-----------+
                                                                  |
                                       +--------------------------+--------------------------+
                                       |                                                     |
                                       v                                                     v
                           +---------------------+                               +----------------------+
                           |   internal/build    |                               |  internal/process    |
                           | run build command   |                               | start/stop server    |
                           | stream build logs   |                               | kill process trees   |
                           +---------------------+                               | stream server logs   |
                                                                                 +----------------------+
```

## Component Explanations

### watcher (`internal/watcher`)

Responsibilities:

- initialize recursive watches under `--root`
- add newly created subdirectories while running
- ignore non-relevant paths/files
- emit normalized file-change events to the engine

Default ignore rules include:

- directories: `.git`, `node_modules`, `bin`, `build`, `dist`, `tmp`
- files: `*.log`
- temporary editor files: `*.swp`, `*.swo`, `*.swx`, `*.tmp`, `*.temp`, `*~`, `.#*`, `#*#`

This keeps rebuild triggers focused on meaningful source changes.

### engine (`internal/engine`)

Responsibilities:

- central orchestration loop
- trigger first build immediately on startup
- debounce noisy file events
- cancel in-flight reload cycle when a newer change arrives
- stop previous server, run build, start new server
- apply crash-loop protection (suppress repeated unstable restarts)

Crash-loop protection:

- a server exit within `1s` of start is treated as a crash
- if crashes exceed `5` within a rolling `10s` window, restart is suppressed and a warning is logged

### build runner (`internal/build`)

Responsibilities:

- execute `--build` command in root directory
- stream stdout/stderr in real time via structured logging
- support context cancellation so stale builds can be aborted

### process manager (`internal/process`)

Responsibilities:

- execute `--exec` command as the actual server process
- stream stdout/stderr in real time
- terminate previous process tree before restart
- expose lifecycle state to the engine

It is implemented with platform-specific behavior using Go build tags.

## How Debouncing Works

Editors often produce multiple file events per save. The engine uses an event batcher with a debounce window (`--debounce`, default `300ms`) to coalesce event bursts:

- first event starts/reset timer
- additional events within window reset timer
- when window elapses, only one reload cycle starts

This avoids unnecessary duplicate rebuilds while staying responsive.

## How Process Termination Works

`internal/process` uses OS-specific implementations:

- Windows (`terminate_windows.go`): `taskkill /PID <pid> /T` and `/F` fallback
- Unix (`terminate_unix.go`): process-group signaling with `SIGTERM` then `SIGKILL`

The manager first attempts graceful tree termination, waits briefly, and escalates if needed. This ensures parent and child processes are cleaned up before the next start.

## Example CLI Usage

```bash
go run ./cmd/hotreload \
  --root ./testserver \
  --build "go build -o ./bin/server ./cmd/server" \
  --exec "./bin/server"
```

### Flags

- `--root`: directory to watch recursively
- `--build`: command used to build the project
- `--exec`: command used to run the built server
- `--debounce`: debounce duration for file-event bursts (default `300ms`)

## Demo Project

A sample HTTP server is included under `testserver/` for quick validation of hot reload behavior.

## Demo Instructions (Windows and Linux)

### Linux/macOS

Build `hotreload`:

```bash
go build -o ./bin/hotreload ./cmd/hotreload
```

Run demo server with hot reload:

```bash
./bin/hotreload \
  --root ./testserver \
  --build "go build -o ./bin/server ./cmd/server" \
  --exec "./bin/server"
```

Trigger a reload:

```bash
touch ./testserver/cmd/server/main.go
```

### Windows (PowerShell)

Build `hotreload`:

```powershell
go build -o .\bin\hotreload.exe .\cmd\hotreload
```

Run demo server with hot reload:

```powershell
.\bin\hotreload.exe --root .\testserver --build "go build -o .\\bin\\server.exe ./cmd/server" --exec ".\\bin\\server.exe"
```

Trigger a reload:

```powershell
(Get-Item .\testserver\cmd\server\main.go).LastWriteTime = Get-Date
```
