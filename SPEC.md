# Backend Engineering Assignment: Hot Reload Engine

# Submission Requirement

Please provide the GitHub repository link for your solution. Ensure access is granted to
**recruitments@trademarkia.com.**

## Assignment Submission Link - https://forms.gle/4RpHZpAi8rbG9QCE

# The Problem

Every time an engineer on the team changes a line of Go code, they have to manually stop the
server, rebuild it, and start it again. This is slow and inefficient.
Your task is to build a CLI tool called **hotreload** that watches a project folder for code
changes. Whenever something changes, it should automatically rebuild and restart the server.
An engineer should be able to run something like
hotreload --root ./myproject --build "go build -o ./bin/server ./cmd/server" --exec "./bin/server"

## Command Parameters

hotreload --root <project-folder> --build "<build-command>" --exec "<run-command>"
● **--root <project-folder>**
Directory to watch for file changes (including all subfolders).
● **--build "<build-command>"**
Command used to **build the project** when a change is detected.
● **--exec "<run-command>"**
Command used to **run the built server** after a successful build.
After that, they should be able to simply edit code and save the server should restart
automatically with the new changes within seconds.
However, this problem is trickier than it looks:


● Editors don’t always save files in a straightforward way.
● Servers don’t always terminate when asked.
● File trees are nested and dynamic.
● Builds and servers can fail.
Your job is to build a tool that handles these real-world scenarios gracefully.

# What We Expect

At a minimum, your tool should:
● Work for a simple project with a single folder containing a few .go files.
● Watch for file changes.
● Rebuild the project automatically.
● Restart the server automatically.
● Trigger the first build immediately when the tool starts (without waiting for a file change).

## Performance Expectations

The tool should feel responsive:
● If a developer saves a file, the server should restart in **under ~2 seconds**.
● If an editor triggers multiple file events quickly, the tool should **avoid rebuilding multiple
times unnecessarily**.

## Change Handling

If a rebuild is already in progress and a new file change occurs:
● The previous build should be discarded.
● Only the **latest state** should be built.

## Process Management

When restarting the server:
● Ensure the previous server process is **fully terminated**.
● Kill **all child processes** , not just the parent.

## Logging

```
● The server logs should stream in real time.
● Logs should not be buffered and dumped only after the process exits.
```

# Bonus Points

Each of the following improvements will increase your score:

## Project Structure Handling

```
● Support projects with multiple folders , not just a flat structure.
● Detect and watch new folders created while the tool is running.
● Handle deleted folders gracefully.
```
## Stability

```
● If the server crashes immediately after starting , the tool should avoid creating a rapid
restart loop.
```
## Process Handling

```
● Some processes don’t shut down when asked nicely your tool should handle stubborn
processes.
● Ensure that killing a process properly frees the resources it was holding.
```
## Scalability

```
● Your tool may run for hours at a time.
● Operating systems limit how many files can be watched simultaneously your
implementation should account for this.
```
## File Filtering

A typical project contains many files your tool shouldn’t care about. Your tool should ignore
things like:
● .git/
● node_modules/
● build artifacts
● temporary editor files

# Ground Rules

```
● Do not use existing hot-reload frameworks such as:
○ air
○ realize
○ reflex
```

```
● Using fsnotify as an event source is allowed, but the rest of the implementation must
be your own.
● Use the log/slog package from the Go standard library for logging.
● Your commit history matters. We want to see how the project evolved, not just the final
code.
```
# Submission

Create a **private GitHub repository** and share access with the team.
The repository should include:

1. The **hotreload** source code
2. A sample **testserver/** directory containing a simple HTTP server for demonstration
3. A way to build and run the demo, such as:
    ○ a **Makefile** , or
    ○ a **script**
4. A short **Loom video** (5–10 minutes) explaining:
● the architecture of your solution
● key design decisions
● how the tool works
● a brief demo of the hot reload functionality

# Evaluation Criteria

**Criteria Weight**
Core tool works correctly 45%
Code quality and idiomatic Go 25%
Bonus features implemented 20%
Tests for tricky components 10%
**Goal:** Build something that you would actually want to use in your daily development workflow.

## Assignment Submission Link - https://forms.gle/4RpHZpAi8rbG9QCE


