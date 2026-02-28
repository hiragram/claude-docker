#!/bin/bash
cd "$(git rev-parse --show-toplevel 2>/dev/null)" || exit 1
REPO_ROOT=$(pwd)

find_free_port() {
  local min=38000
  local max=41000
  local range=$((max - min + 1))
  local port
  local attempts=0
  local max_attempts=100

  while [ $attempts -lt $max_attempts ]; do
    port=$((min + RANDOM % range))
    if ! lsof -i :$port >/dev/null 2>&1; then
      echo $port
      return 0
    fi
    attempts=$((attempts + 1))
  done

  for port in $(seq $min $max); do
    if ! lsof -i :$port >/dev/null 2>&1; then
      echo $port
      return 0
    fi
  done

  return 1
}
FZF_PORT=$(find_free_port)
if [[ -z "$FZF_PORT" ]]; then
    echo "Error: No free port found in range 38000-41000"
    exit 1
fi

format_files() {
  while IFS= read -r filepath; do
    filename=$(basename "$filepath")
    dirpath=$(dirname "$filepath")
    printf '%s\n\033[90m%s\033[0m\0' "$filename" "$dirpath"
  done
}

make_reload_cmd() {
  cat << RELOAD_EOF
reload(git -C '$REPO_ROOT' diff --name-only origin/main | while IFS= read -r f; do printf '%s\n\033[90m%s\033[0m\0' "\$(basename "\$f")" "\$(dirname "\$f")"; done)
RELOAD_EOF
}

while true; do
  (
    sleep 1
    prev=""
    while true; do
      current=$(git -C "$REPO_ROOT" diff --name-only origin/main 2>/dev/null | sort)
      if [ "$current" != "$prev" ]; then
        curl -s -X POST "localhost:$FZF_PORT" -d "$(make_reload_cmd)" >/dev/null 2>&1 || break
        prev="$current"
      fi
      sleep 1
    done
  ) &
  reload_pid=$!

  selected=$(git -C "$REPO_ROOT" diff --name-only origin/main 2>/dev/null | format_files | fzf \
    --listen $FZF_PORT \
    --reverse \
    --ansi \
    --read0 \
    --prompt="File search> ")

  kill $reload_pid 2>/dev/null
  wait $reload_pid 2>/dev/null

  if [ -z "$selected" ]; then
    break
  fi

  filename=$(echo "$selected" | head -n1)
  dirpath=$(echo "$selected" | tail -n1 | sed 's/\x1b\[[0-9;]*m//g')

  if [ "$dirpath" = "." ]; then
    file="$filename"
  else
    file="$dirpath/$filename"
  fi

  zellij run --floating --width=80% --height=80% --name="diff: $file" --close-on-exit -- bash -c "git -C '$REPO_ROOT' diff -U5 origin/main -- '$file' | delta --side-by-side --line-numbers --minus-style=\"syntax #1a0a0a\" --minus-emph-style=\"syntax #5a2a2a\" --plus-style=\"syntax #0a1a0a\" --plus-emph-style=\"syntax #2a5a2a\" --line-numbers-minus-style=\"#ff6666\" --line-numbers-plus-style=\"#66ff66\" --paging=always --pager=\"less -Rc\""
done
