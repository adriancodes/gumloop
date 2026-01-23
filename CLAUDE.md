# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

gumloop is a Go CLI that runs AI coding agents (Claude, Codex, Gemini, Cursor, OpenCode, Ollama) in autonomous loops. Based on Geoffrey Huntley's "Ralph Wiggum" methodology: fresh context per iteration, progress persists in git.

Core insight: LLM context has malloc but no free. Loop termination IS the free(). Each iteration: fresh context → one task → commit → exit → repeat.

## Commands

```bash
# Build
go build -o ./gumloop ./cmd/gumloop

# Run tests
go test ./...

# Run once (default)
./gumloop run -p "Fix the tests"
./gumloop run --prompt-file PROMPT.md

# Loop mode (until no changes or max iterations)
./gumloop run --choo-choo -p "Fix all bugs"
./gumloop run --choo-choo 20 -p "Migrate to TypeScript"

# Select agent
./gumloop run --cli claude -p "..."        # Claude Code (default)
./gumloop run --cli codex -p "..."         # OpenAI Codex
./gumloop run --cli gemini -p "..."        # Google Gemini
./gumloop run --cli cursor -p "..."        # Cursor Agent CLI
./gumloop run --cli opencode -p "..."      # OpenCode
./gumloop run --cli ollama -p "..."        # Ollama (local)

# Select model
./gumloop run --model sonnet -p "..."      # Use Sonnet model
./gumloop run --model opus -p "..."        # Use Opus model

# Other run options
./gumloop run --no-push -p "..."           # Skip git push
./gumloop run --stuck-threshold 5 -p "..." # Exit after 5 iterations without commits
./gumloop run --verify "npm test" -p "..." # Run tests after each iteration
./gumloop run --memory -p "..."            # Enable session memory between runs

# Other commands
./gumloop init                             # Interactive setup wizard
./gumloop init --non-interactive           # Use defaults
./gumloop config show                      # Show effective config
./gumloop config set cli codex             # Set project config
./gumloop memory show                      # Display session memory
./gumloop memory clear                     # Delete session memory file
./gumloop recover                          # Discard uncommitted changes
./gumloop recover 3                        # Reset last 3 commits
./gumloop update                           # Update to latest version
./gumloop uninstall                        # Remove gumloop
./gumloop version                          # Show version
```

## Architecture

Go CLI built with Cobra, Viper, Lipgloss, and Bubbletea:

```
cmd/gumloop/main.go       Entry point
internal/
  cli/                    Cobra commands (root, run, init, config, memory, recover, update, uninstall, version)
  runner/                 Loop execution, iteration logic, metrics
  agent/                  Agent registry and command building
  adapter/                Output parsing (Claude JSON stream, Codex JSON, passthrough)
  config/                 Configuration cascade (defaults → global → project → flags)
  git/                    Git operations and safety checks
  memory/                 Session memory persistence (YAML-based, between runs)
  ui/                     Lipgloss styles (Simpsons theme), Bubbletea wizards
  update/                 Self-update from GitHub releases
```

Key patterns:
- Agents register themselves in `init()` functions
- Config uses Viper with YAML files (`.gumloop.yaml`, `~/.config/gumloop/config.yaml`)
- Adapters parse agent output streams for progress display
- UI uses Simpsons-themed colors (Simpson Yellow, Marge Blue, Bartman Purple, etc.)

## CLI Flag Mapping

When choo-choo mode is enabled, agents get autonomous flags:
- claude: `-p --dangerously-skip-permissions --verbose --output-format stream-json`
- codex: `--full-auto --json`
- gemini: `-p --yolo --output-format text`
- cursor: `-p --force --output-format text`
- opencode: `-p -q -f text`
- ollama: (no flags - local execution)

## Safety Model

Built-in protections:
- Refuses: `~`, `/`, `/etc`, `/usr`, `/var`, `/tmp`
- Requires git repository
- Warns + confirms for --choo-choo in home subdirectories

Git is the safety net: `git reset --hard` or `gumloop recover` recovers from any iteration.

## Testing

```bash
# Unit tests
go test ./...
go test -v ./internal/agent/...   # Specific package
go test -v ./internal/ui/...      # UI tests

# Integration tests (note: test.sh needs updating for new CLI syntax)
./test.sh
```

## The Technique (for modifying this repo)

When working on gumloop itself, understand:
- Prompts should be lean (loaded every iteration, consumes context budget)
- One task per iteration (multiple tasks = context accumulation)
- Progress stored in files/git, not agent memory
- Guardrails use high numbers (99999) for priority
- Plans are disposable, prompts are refined through observation
