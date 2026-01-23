
<p align="center">
  <img src="rwl.png" alt="Ralph Wiggum Loop" width="300">
</p>

# gumloop

**The Ralph Wiggum loop**

Run AI coding agents in autonomous loops until task completion.

Based on [Geoffrey Huntley's methodology](https://ghuntley.com/ralph/) — context rotation with git persistence.

## Why?

LLM context fills up but never clears. As an agent works, stale information accumulates until reasoning degrades.

**The fix:** Kill the agent after each task. Loop handles continuation. Progress lives in git.

```
Iteration 1: Fresh context → work → commit → EXIT (context freed)
Iteration 2: Fresh context → work → commit → EXIT (context freed)
Iteration 3: Fresh context → work → commit → EXIT (context freed)
```

Each iteration starts clean. Progress persists in files and git, not in the agent's memory.

## Requirements

- Go 1.21+ (for building from source)
- Git
- A coding agent CLI (**installed and authenticated by you**)

### Supported Agents

gumloop wraps these coding agent CLIs. **You must install and authenticate them yourself** — gumloop does not do this for you.

| Agent | Install | Authenticate |
|-------|---------|--------------|
| [Claude Code](https://docs.anthropic.com/en/docs/claude-code) | `npm install -g @anthropic-ai/claude-code` | `claude` (follow prompts) |
| [Codex](https://github.com/openai/codex) | `npm install -g @openai/codex` | `export OPENAI_API_KEY=...` |
| [Gemini CLI](https://github.com/google-gemini/gemini-cli) | `npm install -g @google/gemini-cli` | `gemini` (follow prompts) |
| [Cursor Agent CLI](https://docs.cursor.com/en/cli/using) | `curl https://cursor.com/install -fsS \| bash` | `cursor-agent login` |
| [OpenCode](https://github.com/opencode-ai/opencode) | See repo | See repo |
| [Ollama](https://ollama.ai/) | `brew install ollama` | Local - no auth needed |

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/adriancodes/gumloop/main/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/adriancodes/gumloop.git
cd gumloop
go build -o ~/.local/bin/gumloop ./cmd/gumloop
```

## Quick Start

```bash
# Navigate to your project (must be a git repo)
cd my-project

# Interactive setup wizard
gumloop init

# Run once - ask a question
gumloop run -p "What does this codebase do?"

# Run once - make a change
gumloop run -p "Add input validation to the login form"

# Loop until done - let it work autonomously
gumloop run -p "Fix all failing tests" --choo-choo

# Loop with a cap
gumloop run -p "Refactor the API layer" --choo-choo 10

# Loop with verification
gumloop run -p "Fix bugs" --choo-choo --verify "npm test"
```

## Commands

### `gumloop run`

Execute an AI agent with a prompt.

```bash
gumloop run -p "Your prompt here"           # Run once with inline prompt
gumloop run --prompt-file PROMPT.md         # Run once with prompt file
gumloop run --choo-choo                     # Loop until no changes detected
gumloop run --choo-choo 20                  # Loop, max 20 iterations
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-p, --prompt <TEXT>` | Inline prompt text |
| `--prompt-file <FILE>` | Use a prompt file (default: PROMPT.md) |
| `--cli <AGENT>` | Agent: claude, codex, gemini, cursor, opencode, ollama |
| `--model <MODEL>` | Model override (e.g., sonnet, opus, gpt-4) |
| `--choo-choo [N]` | Loop mode, optionally with max iterations |
| `--no-push` | Don't push to remote after iterations |
| `--stuck-threshold <N>` | Exit after N iterations without commits (default: 3) |
| `--verify <CMD>` | Run verification command after each iteration |
| `--memory` | Enable session memory (persists context between runs) |

### `gumloop init`

Interactive setup wizard for new projects.

```bash
gumloop init                    # Interactive wizard
gumloop init --non-interactive  # Use defaults
```

Creates `.gumloop.yaml` config and optionally a `PROMPT.md` template.

### `gumloop config`

Manage configuration values.

```bash
gumloop config show             # Show effective merged config
gumloop config list             # List project config
gumloop config list --global    # List global config
gumloop config get cli          # Get a specific value
gumloop config set cli codex    # Set project config
gumloop config set cli codex --global  # Set global config
```

**Config keys:** `cli`, `model`, `prompt_file`, `auto_push`, `stuck_threshold`, `verify`, `memory`

### `gumloop memory`

Inspect or clear session memory.

```bash
gumloop memory show    # Display current session memory
gumloop memory clear   # Delete session memory file
```

### `gumloop recover`

Discard changes or reset commits.

```bash
gumloop recover       # Discard uncommitted changes
gumloop recover 3     # Reset last 3 commits
```

### `gumloop update`

Update gumloop to the latest version.

```bash
gumloop update
```

### `gumloop uninstall`

Remove gumloop from your system.

```bash
gumloop uninstall
```

## Configuration

Configuration uses a cascade system: **defaults → global → project → CLI flags**

### Project Config (`.gumloop.yaml`)

```yaml
# .gumloop.yaml
cli: claude
model: sonnet
prompt_file: PROMPT.md
auto_push: true
stuck_threshold: 3
verify: "npm test"
```

### Global Config (`~/.config/gumloop/config.yaml`)

Same format as project config. Project settings override global settings.

### Defaults

| Key | Default |
|-----|---------|
| `cli` | `claude` |
| `model` | (none) |
| `prompt_file` | `PROMPT.md` |
| `auto_push` | `true` |
| `stuck_threshold` | `3` |
| `verify` | (none) |
| `memory` | `false` |

## Examples

### One-off tasks

```bash
# Explore a codebase
gumloop run -p "Explain the architecture of this project"

# Quick fix
gumloop run -p "Fix the TypeScript error in src/utils.ts"

# Add a feature
gumloop run -p "Add a dark mode toggle to the settings page"

# Write tests
gumloop run -p "Write unit tests for the auth module"
```

### Autonomous loops

```bash
# Fix all test failures
gumloop run -p "Run tests. Fix the most critical failure. Commit." --choo-choo

# Migrate a codebase
gumloop run -p "Convert one JavaScript file to TypeScript. Commit." --choo-choo 50

# Clear tech debt
gumloop run -p "Find and fix one TODO comment. Commit." --choo-choo

# Improve test coverage
gumloop run -p "Find untested code. Write one test. Commit." --choo-choo 20
```

### Using a prompt file

Create `PROMPT.md`:

```markdown
Study the codebase. Find the most important bug to fix.
Fix it completely. Run tests. Commit with a descriptive message.

Rules:
99999. Search before implementing - don't duplicate existing code.
99999. No placeholders or TODOs. Implement completely.
99999. Run tests before committing.
```

Then run:

```bash
gumloop run --prompt-file PROMPT.md --choo-choo
```

## How It Works

1. **Fresh start** — Agent loads only the prompt (small, deterministic)
2. **Read state from disk** — Git history, modified files
3. **One task** — Implement, test, commit
4. **Check for changes** — If no git changes, work is complete
5. **Loop** — Back to step 1 with fresh context (if `--choo-choo`)

### Completion Detection

The loop stops when:
- No git changes detected (agent has nothing left to do)
- Max iterations reached (if specified)
- Stuck detected: N iterations with changes but no commits (default: 3)
- User presses Ctrl+C

### Run Metrics

When complete, gumloop shows statistics including why the loop exited:

```
┌─────────────────────────────────────┐
│           RUN COMPLETE              │
├─────────────────────────────────────┤
│  Agent:       claude                │
│  Iterations:  3                     │
│  Commits:     2                     │
│  Duration:    1m 45s                │
├─────────────────────────────────────┤
│  Exit: ✅ Complete (no changes)
└─────────────────────────────────────┘
```

Exit reasons: `Complete (no changes)`, `Max iterations`, `Stuck (N iterations without commit)`, `Interrupted`

## Safety

### Built-in protections

- Refuses to run in dangerous directories: `~`, `/`, `/etc`, `/usr`, `/var`, `/tmp`
- Requires a git repository
- Warns before `--choo-choo` mode in home subdirectories

### Git is your safety net

```bash
git diff                  # See what changed
git reset --hard          # Undo everything
git reset --hard HEAD~3   # Undo last 3 commits
gumloop recover           # Same as git reset --hard && git clean -fd
gumloop recover 3         # Same as git reset --hard HEAD~3
```

### For overnight/unattended runs

Use external sandboxing: [E2B](https://e2b.dev/), [Fly Sprites](https://fly.io/), [Modal](https://modal.com/), or a dedicated VM.

## Tuning Your Prompts

Ralph will fail. That's expected. Add guardrails when you see patterns:

```markdown
# PROMPT.md

Study the codebase. Find the most important thing to implement.
Implement it. Run tests. Commit and exit.

## Rules
- Before implementing, search to confirm it's not already done.
- No placeholders or TODOs. Implement completely.
- Always run the test suite before committing.
- Keep commits atomic - one logical change per commit.
```

Build your prompt through observation — when you see the agent make a mistake, add a rule to prevent it.

## Advanced Techniques

### Priority Numbering

LLMs pay more attention to items later in a prompt. Use high numbers for critical rules:

```markdown
1. Study the codebase.
2. Choose one task to implement.
3. Implement it completely.

99999. Search before implementing to avoid duplicating existing code.
999999. Run tests before committing.
9999999. Never commit if tests fail.
```

### Verification

Use the `--verify` flag to run automated checks after each iteration. The command runs through `sh -c`, so you have full shell power.

**Single command:**
```bash
gumloop run --choo-choo --verify "npm test"
```

**Multiple commands** — chain with `&&` (stops on first failure):
```bash
gumloop run --choo-choo --verify "npm run lint && npm test && npm run typecheck"
```

**Script file** — recommended for complex setups:
```bash
gumloop run --choo-choo --verify "./scripts/verify.sh"
```

Example `scripts/verify.sh`:
```bash
#!/bin/bash
set -e  # Exit on first failure

npm run lint
npm test
npm run typecheck
```

**Monorepo setups:**
```bash
# Turborepo - test affected packages
gumloop run --choo-choo --verify "npx turbo run test --filter='...[HEAD^]'"

# Nx - test affected projects
gumloop run --choo-choo --verify "npx nx affected --target=test"

# pnpm workspaces
gumloop run --choo-choo --verify "pnpm -r test"

# Specific packages only
gumloop run --choo-choo --verify "npm test -w @myorg/core -w @myorg/api"
```

For large monorepos, a script that only tests changed packages is recommended:
```bash
#!/bin/bash
# scripts/verify.sh
CHANGED=$(git diff --name-only HEAD^)

# Always lint (fast)
npm run lint

# Only run tests if source files changed
if echo "$CHANGED" | grep -q "^packages/"; then
  npx turbo run test --filter='...[HEAD^]'
fi
```

### Spec-Driven Development

For large projects, use a three-phase workflow: **Spec → Plan → Execute**.

1. Write a detailed `SPEC.md` describing what you want to build
2. Generate a lightweight `IMPLEMENTATION_PLAN.md` with checkboxes
3. Loop with a prompt that reads the plan and implements one task per iteration

See `references/prompt-templates.md` for detailed examples.

## Session Memory

By default, each gumloop run starts from scratch — the agent has no knowledge of previous sessions. With `--memory`, gumloop saves a lightweight summary of what happened and injects it into the next run's prompt.

This solves the "50 First Dates" problem: without memory, the agent wastes iterations rediscovering completed work every time you restart a loop.

### Enable memory

```bash
# Per-run flag
gumloop run --memory --choo-choo -p "Implement the auth module"

# Or set it in config (always enabled for this project)
gumloop config set memory true
```

### How it works

1. During a run, gumloop saves `.gumloop-memory.yaml` after each iteration:

```yaml
# gumloop session memory (auto-generated, safe to edit "remaining" field)
started: "2026-02-04T14:30:05Z"
branch: feat/add-auth
agent: Claude Code
iterations: 7
commits: 5
exit_reason: Max iterations reached
commit_log:
  - hash: a1b2c3d
    message: Add JWT middleware and token validation
  - hash: d4e5f6g
    message: Add auth routes and login handler
remaining: |
  Refresh token rotation has not been implemented yet.
```

2. On the next run with `--memory`, this context is prepended to your prompt:

```
--- PREVIOUS SESSION ---
Last session: 7 iterations, 5 commits on branch feat/add-auth
Agent: Claude Code | Exited: Max iterations reached

Commits made:
- a1b2c3d Add JWT middleware and token validation
- d4e5f6g Add auth routes and login handler

Note: Refresh token rotation has not been implemented yet.
--- END PREVIOUS SESSION ---

[your prompt here]
```

3. The agent picks up where it left off instead of studying the codebase from scratch.

### The `remaining` field

You can hand-edit the `remaining` field in `.gumloop-memory.yaml` to give the next session a specific hint:

```yaml
remaining: |
  Focus on the refresh token rotation endpoint.
  The JWT middleware is done, don't touch it.
```

### Memory file location

- Saved as `.gumloop-memory.yaml` in the project root
- Automatically added to `.gitignore` (it's local state, not code)
- Overwritten each session (only the most recent session is kept)
- Safe to delete — the next run simply starts fresh

## Architecture

gumloop is written in Go using:
- [Cobra](https://github.com/spf13/cobra) for CLI commands
- [Viper](https://github.com/spf13/viper) for configuration
- [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- [Bubbletea](https://github.com/charmbracelet/bubbletea) for the init wizard

```
cmd/gumloop/          Entry point
internal/
  cli/                Cobra commands (run, init, config, recover, update, uninstall)
  runner/             Loop execution and metrics
  agent/              Agent definitions and command building
  adapter/            Output parsing (Claude JSON stream, Codex JSON, etc.)
  config/             Configuration cascade system
  git/                Git operations and safety checks
  memory/             Session memory persistence
  ui/                 Lipgloss styles and Bubbletea wizards
```

## Uninstall

```bash
gumloop uninstall
```

Or manually:

```bash
rm ~/.local/bin/gumloop
rm -rf ~/.config/gumloop  # Optional: remove global config
```

## Philosophy

> "Ralph is a Bash loop."

> "Deterministically bad in an undeterministic world."

> "LLMs are a mirror of operator skill."

Simple loop. Context rotation. Git persistence. That's it.

## Links

- [ghuntley.com/ralph](https://ghuntley.com/ralph/) — Origin of the technique
- [ghuntley.com/loop](https://ghuntley.com/loop/) — Context engineering deep dive
- [ghuntley.com/cursed](https://ghuntley.com/cursed/) — Advanced patterns

## License

MIT
