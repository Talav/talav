#!/bin/bash
set -e

# 1. Sync workspace
echo "Syncing workspace..."
go work sync

# 2. Tidy all modules
echo "Tidying modules..."
for dir in pkg/component/* pkg/fx/* pkg/module/*; do
  if [ -f "$dir/go.mod" ]; then
    echo "Tidying $dir..."
    go mod tidy -C "$dir"
  fi
done

# 3. Vendor workspace dependencies
echo "Vendoring workspace dependencies..."
go work vendor