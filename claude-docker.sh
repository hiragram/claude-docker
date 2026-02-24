#!/bin/bash
set -euo pipefail

IMAGE_NAME="claude-code-docker"
CLAUDE_HOME="${CLAUDE_HOME:-$HOME/.claude}"
CONTAINER_CLAUDE_HOME="$HOME/.claude-docker"
CONTAINER_CLAUDE_JSON="$HOME/.claude-docker.json"

# --- Dockerが使えるかチェック ---
if ! command -v docker &> /dev/null; then
  echo "Error: docker is not installed or not in PATH." >&2
  exit 1
fi

if ! docker info &> /dev/null; then
  echo "Error: Docker daemon is not running." >&2
  exit 1
fi

# --- Dockerイメージのビルド ---
build_image() {
  echo "Building Docker image '$IMAGE_NAME'..."
  local tmpdir
  tmpdir=$(mktemp -d)
  trap "rm -rf '$tmpdir'" EXIT

  cat > "$tmpdir/entrypoint.sh" << 'ENTRYPOINT_EOF'
#!/bin/bash
set -e

# ホスト側パスへのシンボリックリンクを作成
# installed_plugins.json等がホストの絶対パスを参照しているため
if [ -n "$HOST_CLAUDE_HOME" ] && [ "$HOST_CLAUDE_HOME" != "/home/claude/.claude" ]; then
  mkdir -p "$(dirname "$HOST_CLAUDE_HOME")"
  ln -sfn /home/claude/.claude "$HOST_CLAUDE_HOME"
fi

exec "$@"
ENTRYPOINT_EOF

  cat > "$tmpdir/Dockerfile" << 'DOCKERFILE_EOF'
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends git curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash claude

# claudeユーザーとしてインストール
USER claude
SHELL ["/bin/bash", "-c"]
RUN curl -fsSL https://claude.ai/install.sh | bash

ENV PATH="/home/claude/.local/bin:${PATH}"

# entrypointの準備（rootに戻る）
USER root
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
# シンボリックリンク作成先のベースディレクトリをclaude所有にする
RUN mkdir -p /Users && chown claude:claude /Users

USER claude
WORKDIR /workspace

ENTRYPOINT ["/entrypoint.sh"]
CMD ["claude"]
DOCKERFILE_EOF

  docker build -t "$IMAGE_NAME" "$tmpdir"
  trap - EXIT
  rm -rf "$tmpdir"
  echo "Docker image '$IMAGE_NAME' built successfully."
}

# --- イメージ存在チェック ---
if ! docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
  build_image
fi

# --- コンテナ用 ~/.claude-docker/ を準備 ---
mkdir -p "$CONTAINER_CLAUDE_HOME"

# ホストの設定ファイルを同期（毎回最新を反映）
for f in settings.json CLAUDE.md; do
  if [ -f "$CLAUDE_HOME/$f" ]; then
    cp "$CLAUDE_HOME/$f" "$CONTAINER_CLAUDE_HOME/$f"
  fi
done

for d in hooks plugins commands agents; do
  if [ -d "$CLAUDE_HOME/$d" ]; then
    rm -rf "$CONTAINER_CLAUDE_HOME/$d"
    cp -a "$CLAUDE_HOME/$d" "$CONTAINER_CLAUDE_HOME/$d"
  fi
done

# --- ~/.claude.json（オンボーディング状態）がなければ作成 ---
if [ ! -s "$CONTAINER_CLAUDE_JSON" ]; then
  echo '{}' > "$CONTAINER_CLAUDE_JSON"
fi

# --- Claude起動（未認証ならclaude自身がloginを促す） ---
exec docker run -it --rm \
  -e "HOST_CLAUDE_HOME=$CLAUDE_HOME" \
  -v "$CONTAINER_CLAUDE_HOME:/home/claude/.claude" \
  -v "$CONTAINER_CLAUDE_JSON:/home/claude/.claude.json" \
  -v "$(pwd):/workspace" \
  "$IMAGE_NAME" \
  claude "$@"
