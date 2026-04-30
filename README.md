# agent-workspace (`aw`)

A CLI tool for launching agent workspaces with configurable profiles. Supports Docker containers, git worktrees, zellij sessions, and combinations thereof.

## Install

### Shell script

```bash
curl -fsSL https://raw.githubusercontent.com/hiragram/agent-workspace/main/install.sh | bash
```

This downloads the binary to `~/.local/bin/`.

### From source

```bash
go install github.com/hiragram/agent-workspace@latest
```

## Usage

```bash
# Run the default profile
aw

# Run a specific profile
aw <profile-name>

# Self-update
aw update

# Show version
aw --version
```

## Configuration

> **[Detailed Configuration Guide](docs/configuration.md)** -- Full reference for all options, validation rules, and examples.

Create `.agent-workspace.yml` in your git repository root:

```yaml
default: worktree-zellij

profiles:
  # Run Claude Code inside a Docker container
  claude:
    environment: docker
    launch: claude

  # Create a worktree and open a shell
  worktree-shell:
    worktree:
      base: origin/main
    environment: host
    launch: shell

  # Create a worktree and run Claude on host
  worktree-claude:
    worktree: {}
    environment: host
    launch: claude

  # Create a worktree, mount in Docker, run Claude
  worktree-docker:
    worktree: {}
    environment: docker
    launch: claude

  # Create a worktree with a full zellij dev environment
  worktree-zellij:
    worktree: {}
    environment: docker
    launch: zellij
    zellij:
      layout: default
```

If no `.agent-workspace.yml` is found, `aw` uses a built-in default that creates a worktree and starts a zellij dev environment with Docker-based Claude (equivalent to `worktree-zellij` above).

### Profile options

- **`worktree`** (optional): Creates a git worktree.
  - `base` — base ref for the new worktree. Defaults to `origin/main`.
  - `dir` — directory under which worktrees are created. Defaults to `<repoRoot>/worktrees`. Supports `~` expansion; relative paths are resolved against the repo root.
  - `on-create` / `on-end` — shell hooks run after the worktree is created / after the launched process exits.
- **`environment`** (required): `"host"` or `"docker"` — where the main process runs.
- **`launch`** (required): `"shell"`, `"claude"`, or `"zellij"` — what to launch.
- **`zellij`** (optional): Zellij session config. Only valid with `launch: zellij`.

### Top-level defaults

Any profile field can also be declared at the top level of the config. Top-level values act as defaults for every profile, and each profile overrides them field-by-field (sub-structs like `worktree` and `zellij` are merged, not replaced):

```yaml
default: worktree-zellij

# Shared by every profile below
worktree:
  base: origin/main
  dir: ~/.aw/worktrees
environment: host

profiles:
  shell:
    launch: shell                # inherits worktree + environment from top level
  docker-claude:
    environment: docker          # overrides only environment
    launch: claude
    worktree:
      dir: /tmp/aw-worktrees     # overrides only worktree.dir; base is inherited
```

## What it does (Docker mode)

On first run with `environment: docker`:

1. Builds a lightweight Docker image (Debian slim + git + curl + Node.js + gh)
2. Installs Claude Code into a persistent Docker volume
3. Prompts you to log in via OAuth (browser-based)

On subsequent runs, it starts instantly with your existing authentication and settings.

## Host settings

The following files from `~/.claude/` are synced into the container on each launch:

- `settings.json` - Claude Code configuration
- `CLAUDE.md` - global instructions
- `hooks/` - custom hook scripts
- `plugins/` - installed plugins and skills
- `commands/` - custom slash commands
- `agents/` - custom agent definitions

These are copied to `~/.agent-workspace/` to avoid conflicts with the host-side Claude Code (which uses macOS Keychain for credentials).

## Data storage

| Path | Purpose |
|------|---------|
| `~/.agent-workspace/` | Container-side Claude config (credentials, settings copy) |
| `~/.agent-workspace.json` | Onboarding state |
| Docker volume `claude-code-local` | Claude Code installation (persists auto-updates) |

## Uninstall

```bash
# Remove binary
rm ~/.local/bin/aw

# Remove data
rm -rf ~/.agent-workspace ~/.agent-workspace.json
docker rmi claude-code-docker
docker volume rm claude-code-local
```

## Development

```bash
# Run tests
go test ./...

# Build
go build -o aw .

# Lint
golangci-lint run
```

## Requirements

### Host (required)

These tools must be installed on the host and available on `PATH`.

| Tool | When needed | Purpose |
|------|-------------|---------|
| `git` | `worktree` profiles | Worktree creation, repo root detection, remote fetch |
| `docker` | `environment: docker` profiles | Build image, create volume, run container |
| `zellij` | `launch: zellij` profiles | Multi-pane session |

### Host (additional — required by zellij layout panes)

When `launch: zellij` is used, the layout spawns helper panes that shell out to the following tools. Install the ones for the panes you actually use.

**`git-diff-picker` pane** — interactive diff viewer
- `fzf` — fuzzy picker (listen mode)
- `delta` — side-by-side diff renderer
- `lsof` — free-port detection for the fzf listen server
- `curl` — posts reload commands to the fzf server

**`pr-status` pane** — current branch's PR status
- `gh` (GitHub CLI) — fetches PR info and checks
- `jq` — parses PR JSON

**`plans-watcher` pane** — live Markdown preview of `plans/`
- `fswatch` **or** `entr` — file-change watcher (either works; fswatch preferred)
- `glow` *(optional)* — Markdown renderer; falls back to `cat` if missing

### Container (bundled — for reference)

These are installed automatically inside the Docker image; you do **not** need them on the host. Listed here so you know what's available inside `environment: docker` containers.

- Base: Debian bookworm-slim
- `git`, `curl`, `wget`, `ca-certificates`, `openssh-client`, `sudo`, `setpriv`
- `python3`, `python3-pip`, `python3-venv`
- `gh` (GitHub CLI)
- Node.js 22, `corepack`, `pnpm`
- Go 1.23.6

### Optional

- `gpg` / `gpg-agent` — only if you sign git commits
- SSH keys / `ssh-agent` — only if you push/pull over SSH
