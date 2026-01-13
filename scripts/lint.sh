#!/bin/bash

# Run linter in all modules
echo "Running linter in modules..."
for dir in pkg/component/* pkg/fx/* pkg/module/*; do
  if [ -f "$dir/go.mod" ]; then
    echo "Linting $dir..."
    (cd "$dir" && golangci-lint run --fix) || true
  fi
done