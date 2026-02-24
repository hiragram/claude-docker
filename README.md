# claude-docker

Run [Claude Code](https://docs.anthropic.com/en/docs/claude-code) inside a Docker container with your host settings and persistent authentication.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/hiragram/claude-docker/main/install.sh | bash
```

This installs the `claude-docker` command to `~/.local/bin/`.

To update, run the same command again.

## Usage

```bash
claude-docker
```

All arguments are passed through to `claude`:

```bash
claude-docker --dangerously-skip-permissions
claude-docker -p "explain this codebase"
```

The current directory is mounted as the workspace inside the container.

## What it does

On first run, `claude-docker`:

1. Builds a lightweight Docker image (Debian slim + git + curl)
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

These are copied to `~/.claude-docker/` to avoid conflicts with the host-side Claude Code (which uses macOS Keychain for credentials).

## Data storage

| Path | Purpose |
|------|---------|
| `~/.claude-docker/` | Container-side Claude config (credentials, settings copy) |
| `~/.claude-docker.json` | Onboarding state |
| Docker volume `claude-code-local` | Claude Code installation (persists auto-updates) |

## Uninstall

```bash
rm ~/.local/bin/claude-docker
rm -rf ~/.claude-docker ~/.claude-docker.json
docker rmi claude-code-docker
docker volume rm claude-code-local
```

## Requirements

- Docker
- bash
