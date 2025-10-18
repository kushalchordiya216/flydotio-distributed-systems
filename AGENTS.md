# Repository Guidelines

## Project Structure & Module Organization
The workspace is organized around Maelstrom exercises. Distributed node logic lives in `maelstrom-broadcast`, while other scenarios live beside it (e.g., `maelstrom-echo`, `maelstrom-unique-ids`). Keep new Go packages co-located with their workload directory and prefer small files such as `main.go` plus any helper modules. Helpers will be restricted to their own packages unless explicitly specified otherwise.


## Build, Test, and Development Commands
cd inside a module and use `go install .` to build and install binaries. Then run `./maelstrom/maelstrom test -w <workload> --bin ~/go/bin/<module_name>` from the root directory to test . Run these commands before opening a pull request to catch regressions early.

## Coding Style & Naming Conventions
- Follow standard Go conventions: tabs for indentation, camelCase for locals, and exported identifiers in PascalCase with a short doc comment when they leave the package.
- Always run `gofmt` (or `goimports`) on touched files; the CI mirrors this formatting.
- Keep functions small and add brief comments only when logic is non-obvious.
- If there are known libraries available for a problem, prefer using them over rolling your own.
- Prefer composition over inheritance with small, purpose-specific interfaces
- Write short, focused functions with single responsibility
- Handle errors explicitly using wrapped errors for traceability
- Use goroutines safely with proper synchronization mechanisms
- Guard shared state with channels or sync primitives
- abstract types into a separate file and also consolidate helper function related to a specific domain together into separate files for better readability
