# Configuration Guide

This document describes the `.agent-workspace.yml` configuration file in detail.

## Overview

`aw` reads its configuration from a file named `.agent-workspace.yml` placed at the root of your git repository. If no configuration file is found (or you're not in a git repository), a built-in default is used that runs Claude Code in Docker.

## File location

```
your-repo/
  .agent-workspace.yml   <-- place it here
  src/
  ...
```

`aw` finds the file by running `git rev-parse --show-toplevel` to locate the repository root, then looks for `.agent-workspace.yml` in that directory.

## Minimal example

The simplest valid configuration defines a single profile:

```yaml
profiles:
  my-profile:
    environment: docker
    launch: claude
```

## Full example

```yaml
default: worktree-zellij

profiles:
  claude:
    environment: docker
    launch: claude

  worktree-shell:
    worktree:
      base: origin/main
    environment: host
    launch: shell

  worktree-claude:
    worktree: {}
    environment: host
    launch: claude

  worktree-docker:
    worktree: {}
    environment: docker
    launch: claude

  worktree-zellij:
    worktree: {}
    environment: docker
    launch: zellij
    zellij:
      layout: default

  worktree-with-setup:
    worktree:
      base: origin/main
      on-create: "./scripts/setup.sh"
    environment: docker
    launch: claude
```

## Top-level fields

### `default`

| | |
|---|---|
| Type | `string` |
| Required | No |
| Default | _(none)_ |

The name of the profile to use when you run `aw` without arguments. Must match one of the keys in `profiles`.

If omitted, running `aw` without arguments prints the list of available profiles instead of launching one.

### `profiles`

| | |
|---|---|
| Type | `map[string]Profile` |
| Required | Yes (at least one profile) |

A map of named profiles. Each key is the profile name (used as `aw <name>`), and the value is a [Profile](#profile-fields) object.

## Profile fields

### `environment` (required)

| | |
|---|---|
| Type | `string` |
| Values | `"host"`, `"docker"` |

Where the main process runs.

- **`host`** -- Runs the launched command directly on your machine.
- **`docker`** -- Runs inside a Docker container. On first run, `aw` builds a lightweight Docker image (Debian slim + git + curl + Node.js + gh), installs Claude Code into a persistent volume, and prompts for OAuth login.

### `launch` (required)

| | |
|---|---|
| Type | `string` |
| Values | `"shell"`, `"claude"`, `"zellij"` |

What command to launch.

- **`shell`** -- Opens an interactive shell.
- **`claude`** -- Launches Claude Code.
- **`zellij`** -- Starts a zellij session with a multi-pane layout (plans watcher, git diff picker, PR status, and Claude Code).

### `worktree` (optional)

| | |
|---|---|
| Type | `object` or omitted |

If present, `aw` creates a git worktree before running the profile. The worktree is created in a temporary location, and the launched process's working directory is set to it.

To enable worktree creation with all defaults, use an empty object:

```yaml
worktree: {}
```

#### `worktree.base`

| | |
|---|---|
| Type | `string` |
| Default | `"origin/main"` |

The git ref to base the worktree branch on. This can be a remote branch, local branch, tag, or commit hash.

```yaml
worktree:
  base: origin/develop
```

#### `worktree.on-create`

| | |
|---|---|
| Type | `string` |
| Default | _(none)_ |

A shell command to run after the worktree is created. The command is executed via `sh -c` with the working directory set to the newly created worktree path.

The following environment variables are available to the hook script:

| Variable | Description |
|---|---|
| `AW_WORKTREE_PATH` | Absolute path to the created worktree |
| `AW_WORKTREE_BRANCH` | Branch name of the created worktree |
| `AW_REPO_ROOT` | Absolute path to the git repository root |
| `AW_PROFILE_NAME` | Name of the profile being run |
| `AW_ENVIRONMENT` | Profile environment (`host` or `docker`) |

If the hook exits with a non-zero status, the pipeline is aborted.

```yaml
worktree:
  on-create: "npm install && npm run setup"
```

### `zellij` (optional)

| | |
|---|---|
| Type | `object` or omitted |

Configuration for zellij sessions. **Only valid when `launch` is `"zellij"`**. Setting this on a profile with a different `launch` mode causes a validation error.

#### `zellij.layout`

| | |
|---|---|
| Type | `string` |
| Default | `"default"` |

The layout to use for the zellij session. Currently only `"default"` is supported, which creates a multi-pane layout with:

- Claude Code (main pane)
- Plans watcher
- Git diff picker
- PR status

## Built-in default

When no `.agent-workspace.yml` is found, `aw` behaves as if the following configuration were present:

```yaml
default: worktree-zellij

profiles:
  claude:
    environment: docker
    launch: claude

  worktree-zellij:
    worktree: {}
    environment: docker
    launch: zellij
    zellij:
      layout: default
```

## Validation rules

`aw` validates your configuration on every run. The following rules are enforced:

1. **At least one profile must be defined.** An empty `profiles` map is an error.
2. **`environment` is required** on every profile. Must be `"host"` or `"docker"`.
3. **`launch` is required** on every profile. Must be `"shell"`, `"claude"`, or `"zellij"`.
4. **`zellij` config requires `launch: zellij`.** Specifying `zellij:` on a profile with a different launch mode is an error.
5. **`default` must reference an existing profile.** If `default` is set, it must match one of the keys in `profiles`.

### Example error messages

```
Error: environment is required ("host" or "docker")
Error: unknown environment: "kubernetes" (must be "host" or "docker")
Error: launch is required ("shell", "claude", or "zellij")
Error: unknown launch mode: "tmux" (must be "shell", "claude", or "zellij")
Error: zellij config is only valid with launch: zellij
Error: default profile "nonexistent" not found in profiles
```

## Valid combinations

The following table shows all valid combinations of `worktree`, `environment`, and `launch`:

| worktree | environment | launch | Description |
|----------|-------------|--------|-------------|
| _(omitted)_ | `host` | `shell` | Open a shell in the current directory |
| _(omitted)_ | `host` | `claude` | Run Claude Code in the current directory |
| _(omitted)_ | `docker` | `shell` | Open a shell inside Docker |
| _(omitted)_ | `docker` | `claude` | Run Claude Code inside Docker |
| _(omitted)_ | `docker` | `zellij` | Start a zellij session with Docker-based Claude |
| `{}` | `host` | `shell` | Create a worktree, open a shell in it |
| `{}` | `host` | `claude` | Create a worktree, run Claude Code in it |
| `{}` | `docker` | `shell` | Create a worktree, mount in Docker, open a shell |
| `{}` | `docker` | `claude` | Create a worktree, mount in Docker, run Claude Code |
| `{}` | `docker` | `zellij` | Create a worktree, start zellij with Docker-based Claude |
| `{base: ...}` | `host` | `zellij` | Create a worktree from custom ref, start zellij on host |

All other combinations follow the same pattern. `worktree` is always optional and independent of `environment`/`launch`.

## Host settings sync (Docker mode)

When using `environment: docker`, the following files and directories from `~/.claude/` are automatically synced into the container at each launch:

**Files:**
- `settings.json` -- Claude Code configuration
- `CLAUDE.md` -- Global instructions

**Directories:**
- `hooks/` -- Custom hook scripts
- `plugins/` -- Installed plugins and skills
- `commands/` -- Custom slash commands
- `agents/` -- Custom agent definitions

These are copied to `~/.agent-workspace/` to avoid conflicts with the host-side Claude Code.

## Tips

- Use `aw profiles` to see all available profiles and which config file they were loaded from.
- Profile names can be any valid YAML string. Keep them short and descriptive (e.g., `claude`, `worktree-shell`).
- You can commit `.agent-workspace.yml` to your repository so all contributors share the same workspace profiles.
- If you need different profiles for different machines, use separate branches or a gitignored override (not currently supported, but the built-in default handles the no-config case gracefully).
