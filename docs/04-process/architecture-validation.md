# Architecture Validation

> **Automated enforcement of layering rules and import restrictions**

---

## Overview

The Talav framework uses automated tooling to enforce architecture constraints defined in `.cursor/architecture.yaml`. This prevents layering violations and maintains clean separation of concerns.

---

## Architecture Constraints File

**Location**: `.cursor/architecture.yaml`

This machine-readable file defines:
- Layer definitions (component, fx, application)
- Import restrictions per layer
- Naming conventions
- Quality gates (test coverage, linting)
- Common patterns and anti-patterns
- Validation rules for automated checking

**Key Rules**:
```yaml
layers:
  - name: component
    paths: ["pkg/component/*"]
    can_import: ["pkg/component/*", "stdlib", "third-party"]
    cannot_import: ["pkg/fx/*", "pkg/module/*"]
  
  - name: fx
    paths: ["pkg/fx/*"]
    can_import: ["pkg/component/*", "stdlib", "third-party"]
    cannot_import: ["pkg/module/*"]
  
  - name: application
    paths: ["pkg/module/*", "main.go"]
    can_import: ["*"]  # Can import everything
```

---

## Validation Script

**Location**: `scripts/validate-architecture.sh`

### What It Checks

1. **Import Restrictions**
   - Components don't import `pkg/fx/*` or `pkg/module/*`
   - FX modules don't import `pkg/module/*`

2. **Circular Dependencies**
   - No circular imports between packages
   - Uses `go mod graph` to detect cycles

### Usage

```bash
# Validate all modules
./scripts/validate-architecture.sh

# Output on success:
# ✓ Architecture validation passed
#   No layering violations detected

# Output on failure:
# ✗ VIOLATION: pkg/component/mycomp
#   Layer: component
#   Forbidden import: github.com/talav/talav/pkg/fx/fxcore
```

### Exit Codes

- `0` - No violations found
- `1` - Violations detected

---

## Pre-Commit Hook

**Location**: `.git/hooks/pre-commit`

Automatically runs before each commit to catch violations early.

### Checks Performed

1. **Architecture Validation** (blocking)
   - Runs `validate-architecture.sh`
   - Blocks commit if violations found

2. **Linting** (blocking)
   - Runs `golangci-lint` on changed modules
   - Blocks commit if linter errors found

3. **Test Coverage** (warning only)
   - Checks coverage for changed modules
   - Warns if below target (80% for components, 60% for FX)
   - Does NOT block commit (allows WIP commits)

### Example Output

```bash
$ git commit -m "Add new component"

🚀 Running pre-commit checks...

[1/3] Validating architecture...
✓ Architecture validation passed

[2/3] Running linter...
  Linting pkg/component/mycomp...
✓ Linter passed

[3/3] Checking test coverage...
  ✓ pkg/component/mycomp (85%)
✓ Coverage passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ All pre-commit checks passed
```

### Bypassing the Hook

If you need to commit despite failures (not recommended):

```bash
git commit --no-verify -m "WIP: temporary commit"
```

**Use sparingly** - fix violations instead of bypassing checks.

---

## CI Integration

The same validation runs in CI to catch violations that bypass local checks.

### GitHub Actions Example

```yaml
name: Architecture Validation

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Install golangci-lint
        run: |
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
      
      - name: Validate architecture
        run: ./scripts/validate-architecture.sh
      
      - name: Run linter
        run: ./scripts/lint.sh
      
      - name: Check coverage
        run: |
          for dir in pkg/component/* pkg/fx/*; do
            if [ -f "$dir/go.mod" ]; then
              (cd "$dir" && go test -cover ./...)
            fi
          done
```

---

## Common Violations & Fixes

### Violation: Component Imports FX

**Error**:
```
✗ VIOLATION: pkg/component/mycomp
  Layer: component
  Forbidden import: github.com/talav/talav/pkg/fx/fxcore
```

**Fix**: Remove FX import, receive dependencies via constructor

**Before** (wrong):
```go
package mycomp

import "github.com/talav/talav/pkg/fx/fxcore"  // ❌ VIOLATION

func NewService() *Service {
    // Component knows about FX
}
```

**After** (correct):
```go
package mycomp

// ✅ No FX imports, receives dependencies via constructor
func NewService(logger *slog.Logger, repo Repository) *Service {
    return &Service{logger: logger, repo: repo}
}
```

---

### Violation: FX Module Imports Framework

**Error**:
```
✗ VIOLATION: pkg/fx/fxmycomp
  Layer: fx
  Forbidden import: github.com/talav/talav/pkg/module/framework
```

**Fix**: Receive framework services via FX DI

**Before** (wrong):
```go
package fxmycomp

import "github.com/talav/talav/pkg/module/framework"  // ❌ VIOLATION

func someFunction(app *framework.Application) {
    // FX module depends on framework
}
```

**After** (correct):
```go
package fxmycomp

// ✅ If framework services needed, receive via FX DI
type Params struct {
    fx.In
    Logger      *slog.Logger              // Provided by framework via FX
    Coordinator *ServiceCoordinator       // Provided by framework via FX
}
```

---

### Violation: Circular Dependency

**Error**:
```
✗ CIRCULAR DEPENDENCY: pkg/component/mycomp
```

**Fix**: Break the cycle by introducing interfaces or restructuring

**Pattern**: If A imports B and B imports A, introduce interface in A, implement in B

**Example**:
```go
// pkg/component/a/service.go
type BService interface {  // ✅ Interface in A
    DoThing() error
}

type AService struct {
    b BService  // Depends on interface
}

// pkg/component/b/service.go
type BService struct {
    // Can import A without cycle
}

func (b *BService) DoThing() error {
    // Implements A's interface
}
```

---

## LLM Guidance

When using AI assistants (Cursor, Copilot), include this context:

```
Before generating code, read .cursor/architecture.yaml.

Key constraints:
1. Components (pkg/component/*) CANNOT import pkg/fx/* or pkg/module/*
2. FX modules (pkg/fx/*) CANNOT import pkg/module/*
3. Application (pkg/module/*) can import everything

Validate imports against these rules before suggesting code.
```

---

## Manual Validation

### Check Specific Module

```bash
# Check what a module imports
go list -f '{{range .Imports}}{{.}}{{"\n"}}{{end}}' ./pkg/component/mycomp | grep talav

# Should only see pkg/component/* imports
# NOT pkg/fx/* or pkg/module/*
```

### Check All Dependencies

```bash
# See full dependency tree
go list -deps ./pkg/component/mycomp

# Check for forbidden imports
go list -deps ./pkg/component/mycomp | grep "talav/pkg/fx"
# Should return empty if no violations
```

---

## Troubleshooting

### Hook Not Running

```bash
# Check if hook is executable
ls -la .git/hooks/pre-commit

# Make executable if needed
chmod +x .git/hooks/pre-commit

# Test hook manually
.git/hooks/pre-commit
```

### False Positives

If the validator incorrectly flags a violation, check:

1. **Module boundaries**: Is the module path correct in `go.mod`?
2. **Import paths**: Are you importing the full path?
3. **Workspace sync**: Run `./scripts/tidy.sh` to sync modules

---

## Maintenance

### Updating Constraints

When architecture evolves:

1. Update `.cursor/architecture.yaml` with new rules
2. Run validation to catch existing violations
3. Fix violations before merging
4. Document rationale in `docs/03-logs/decision-log.md`

### Adding New Layers

If adding a new layer (e.g., `pkg/adapter/*`):

1. Add layer definition to `.cursor/architecture.yaml`
2. Define import restrictions
3. Update validation script if needed
4. Update documentation

---

## Benefits

### For Developers
- Catch violations immediately (pre-commit)
- Clear error messages with actionable fixes
- No manual code review needed for architecture

### For Teams
- Consistent architecture across codebase
- Onboard new developers faster (rules enforced automatically)
- Reduce technical debt (violations prevented, not cleaned up later)

### For LLMs
- Machine-readable constraints (no ambiguity)
- Validation feedback improves code generation
- Consistent patterns across generated code

---

## Related Documents

- [Architecture Rules](../../.cursor/rules/) - Human-readable layering rules
- [Development Workflow](dev-workflow.md) - Daily development process
- [Decision Log](../03-logs/decision-log.md) - Why we enforce layering
- [System Map](../00-context/system-map.md) - Current architecture
