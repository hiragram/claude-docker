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

Create `.agent-workspace.yml` in your git repository root:

```yaml
default: docker-claude

profiles:
  # Run Claude Code inside a Docker container
  docker-claude:
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

If no `.agent-workspace.yml` is found, `aw` uses a built-in default that runs Claude Code in Docker (equivalent to `docker-claude` above).

### Profile options

- **`worktree`** (optional): Creates a git worktree. `base` defaults to `origin/main`.
- **`environment`** (required): `"host"` or `"docker"` — where the main process runs.
- **`launch`** (required): `"shell"`, `"claude"`, or `"zellij"` — what to launch.
- **`zellij`** (optional): Zellij session config. Only valid with `launch: zellij`.

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

- Docker (for `environment: docker` profiles)
- git (for `worktree` profiles)
- zellij (for `launch: zellij` profiles)
