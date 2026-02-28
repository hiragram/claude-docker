#!/bin/bash
PLANS_DIR="./plans"
mkdir -p "$PLANS_DIR"

show_latest() {
    latest=$(ls -t "$PLANS_DIR"/*.md 2>/dev/null | head -1)
    printf '\033[2J\033[3J\033[H'
    if [[ -n "$latest" ]]; then
        echo "=== $latest ==="
        echo ""
        if command -v glow &>/dev/null; then
            glow --style ~/.config/glow/neon.json "$latest"
        else
            cat "$latest"
        fi
    else
        echo "Waiting for .md files in $PLANS_DIR..."
    fi
}

show_latest

if command -v fswatch &>/dev/null; then
    while true; do
        fswatch -1 "$PLANS_DIR" >/dev/null
        show_latest
    done
elif command -v entr &>/dev/null; then
    while true; do
        find "$PLANS_DIR" -name "*.md" 2>/dev/null | entr -d sh -c "$(declare -f show_latest); show_latest"
    done
else
    echo "Error: fswatch or entr required."
    echo "Install: brew install fswatch"
    exit 1
fi
