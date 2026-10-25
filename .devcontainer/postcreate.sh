#!/usr/bin/env bash
set -euo pipefail

uid="$(id -u)"
gid="$(id -g)"

# Every path that gets a named volume in devcontainer.json.
dirs=(
  /datahome
  /go/pkg/mod
  "$HOME/.cache/go-build"
  "$HOME/.cache/npm"
  "$HOME/.cache/golangci-lint"
  "$HOME/.npm-global"
  "/workspace/tmp"
)

for d in "${dirs[@]}"; do
  # Create the directory if the volume or image didn't.
  [ -d "$d" ] || sudo mkdir -p "$d"

  # Only chown when the top level is wrong. A recursive chown over a large
  # module cache is slow, and a volume is wrong at the root or not at all.
  if [ "$(stat -c '%u:%g' "$d")" != "$uid:$gid" ]; then
    echo "Fixing ownership of $d"
    sudo chown -R "$uid:$gid" "$d"
  fi
done

# Make sure the parent dirs themselves are yours. Docker creates missing
# parents of mount targets as root.
for p in "$HOME/.cache" "$HOME/.config"; do
  [ -d "$p" ] || mkdir -p "$p"
  [ "$(stat -c '%u' "$p")" = "$uid" ] || sudo chown "$uid:$gid" "$p"
done

# Shell history file for bash (fish creates its own under XDG_DATA_HOME).
mkdir -p /datahome/bash /datahome/fish
touch /datahome/bash/.bash_history

exit 0
