# Development Workflow

> **Daily development loop for humans + LLMs building with Talav**

---

## Overview

This workflow optimizes for:
- **Fast feedback loops** (build → test → iterate in <1 minute)
- **AI-assisted development** (leverage Cursor/Copilot effectively)
- **Quality gates** (linting, testing, coverage before commit)
- **Architecture enforcement** (prevent layering violations)

---

## Daily Development Loop

### 1. Start Work on Feature/Fix

```bash
# Pull latest changes
git pull origin main

# Sync workspace modules
./scripts/tidy.sh

# Run linter to ensure clean state
./scripts/lint.sh

# Create feature branch
git checkout -b feature/my-feature
```

---

### 2. Development Cycle (Repeat)

#### For New Component

1. **Create component structure**:
   ```bash
   mkdir -p pkg/component/mycomp
   cd pkg/component/mycomp
   go mod init github.com/talav/talav/pkg/component/mycomp
   ```

2. **Write component code** (AI-assisted):
   - Open `README.md` first, describe what component does
   - Ask AI: "Create a Go component that [description] following layering rules"
   - AI generates code respecting `.cursor/rules/`

3. **Write tests** (TDD style):
   ```go
   // mycomp_test.go
   func TestMyComp_DoThing(t *testing.T) {
       comp := NewMyComp()
       result := comp.DoThing("input")
       assert.Equal(t, "expected", result)
   }
   ```

4. **Run tests**:
   ```bash
   go test -v ./...
   go test -cover ./...  # Check coverage
   ```

5. **Create FX module** (if component needs DI):
   ```bash
   mkdir -p pkg/fx/fxmycomp
   cd pkg/fx/fxmycomp
   go mod init github.com/talav/talav/pkg/fx/fxmycomp
   ```

   ```go
   // module.go
   var FxMyCompModule = fx.Module(
       "mycomp",
       fx.Provide(mycomp.NewMyComp),
       // Register command if needed
       fxcore.AsRootCommand(cmd.NewMyCompCmd),
   )
   ```

6. **Add to workspace**:
   ```bash
   # Edit go.work
   use (
       // ... existing
       ./pkg/component/mycomp
       ./pkg/fx/fxmycomp
   )
   
   # Sync workspace
   ./scripts/tidy.sh
   ```

#### For Existing Component Changes

1. **Read component README** (understand current API)
2. **Ask AI for changes**: "Modify [component] to [feature], respecting existing patterns"
3. **Write/update tests first** (TDD)
4. **Implement changes**
5. **Run tests**: `go test -v ./pkg/component/mycomp/...`

---

### 3. Quality Checks (Before Commit)

```bash
# 1. Lint all modified modules
./scripts/lint.sh

# 2. Run tests with coverage
go test -cover ./...

# 3. Check architecture (no layering violations)
# (linter catches these, but manual check if needed)
go list -deps ./pkg/component/mycomp | grep "pkg/fx"
# Should return empty (components can't import FX)

# 4. Verify imports
go mod tidy -C pkg/component/mycomp
go mod verify
```

**Quality Gates (Must Pass)**:
- ✅ All tests pass
- ✅ No linter errors
- ✅ Coverage ≥80% for components (≥60% for FX modules)
- ✅ No layering violations

---

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with conventional commit message
git commit -m "feat(mycomp): add thing that does X

- Implements Y feature
- Adds tests for Z scenario
- Closes #123"

# Push branch
git push origin feature/my-feature
```

**Commit Message Format**:
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

---

### 5. Create Pull Request

**PR Description Template**:
```markdown
## What Changed
- Added MyComp component that does X
- Created FX module for DI integration
- Added tests with 85% coverage

## Why
Needed to support Y use case for Z feature (see #123)

## Testing
- `go test ./pkg/component/mycomp` - all pass
- Manual test: `go run examples/mycomp/main.go`

## Checklist
- [x] Tests pass
- [x] Linter clean
- [x] Coverage ≥80%
- [x] README updated
- [x] No layering violations
```

---

## AI-Assisted Development Patterns

### Pattern: Use Rules for Context

**Workflow**:
1. Open `.cursor/rules/` to see architecture constraints
2. Ask AI: "Following the layering rules, create a component for X"
3. AI respects boundaries (components don't import FX)

**Example Prompt**:
```
Create a new component `pkg/component/notification` that:
1. Sends notifications via email/SMS
2. Has a NotificationService with Send(msg) method
3. Includes tests with testify assertions
4. Follows function ordering from 02-function-ordering.mdc
5. Does NOT import pkg/fx/* (component layer)

Also create the FX module in pkg/fx/fxnotification that:
1. Provides NotificationService
2. Registers a "send-notification" command
3. Loads config from NotificationConfig
```

---

### Pattern: Show Example, Ask for Similar

**Workflow**:
1. Point AI to existing component
2. Ask for similar implementation

**Example Prompt**:
```
Look at pkg/component/email and pkg/fx/fxemail.
Create a similar structure for SMS notifications:
- pkg/component/sms with SMSService
- pkg/fx/fxsms with FX module
- Follow the same patterns (config, factory, command)
```

---

### Pattern: Test-First with AI

**Workflow**:
1. Write test cases
2. Ask AI to implement code that passes tests

**Example Prompt**:
```go
I have this test:

func TestNotificationService_Send(t *testing.T) {
    service := NewNotificationService(mockSender)
    err := service.Send(Notification{To: "user@example.com", Message: "Hi"})
    assert.NoError(t, err)
}

Implement NotificationService that makes this test pass.
Use constructor injection for dependencies.
```

---

## Common Tasks

### Add New HTTP Endpoint (Zorya)

```go
// 1. Define input/output structs
type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
}

type CreateUserOutput struct {
    Status int `status:"201"`
    Body   User
}

// 2. Implement handler
func createUser(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
    user := service.CreateUser(input.Body.Name, input.Body.Email)
    return &CreateUserOutput{Status: http.StatusCreated, Body: user}, nil
}

// 3. Register route
zorya.Post(api, "/users", createUser)
```

**Test**:
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}'
```

---

### Add Configuration Field

```go
// 1. Update config struct
type MyConfig struct {
    Existing string `mapstructure:"existing"`
    NewField int    `mapstructure:"new_field" default:"42"`  // Add this
}

// 2. Update config.yaml
myservice:
  existing: "value"
  new_field: 100

// 3. Access in component
func NewMyService(cfg MyConfig) *MyService {
    fmt.Printf("New field: %d\n", cfg.NewField)
    // ...
}
```

---

### Add Command to Component

```go
// 1. Create cmd/mycommand.go in component
package cmd

func NewMyCommand(service *MyService) *cobra.Command {
    return &cobra.Command{
        Use: "my-command",
        RunE: func(cmd *cobra.Command, args []string) error {
            return service.DoThing(cmd.Context())
        },
    }
}

// 2. Register in FX module
var FxMyModule = fx.Module(
    "my",
    fx.Provide(NewMyService),
    fxcore.AsRootCommand(cmd.NewMyCommand),  // Add this
)

// 3. Run command
go run main.go my-command
```

---

## Debugging

### Debug Test Failure

```bash
# Run single test with verbose output
go test -v -run TestMyFunc ./pkg/component/mycomp

# Run with coverage to see what's not executed
go test -cover -v -run TestMyFunc ./pkg/component/mycomp

# Debug with delve
dlv test ./pkg/component/mycomp -- -test.run TestMyFunc
```

---

### Debug Module Resolution Issues

```bash
# Check workspace sync
go work sync
go work vendor  # Optional: vendor for offline

# Check module versions
go list -m all | grep talav

# Force re-download if needed
go clean -modcache
go mod tidy
```

---

### Debug Layering Violations

```bash
# Check what a component imports
go list -deps ./pkg/component/mycomp | grep "talav"

# Should only see:
#   pkg/component/* (other components)
#   NOT pkg/fx/* or pkg/module/*

# Find who imports what
go mod why github.com/talav/talav/pkg/fx/fxcore
```

---

## Tools

### VS Code / Cursor Setup

**Extensions**:
- Go (official)
- gopls (language server)
- Cursor AI (built-in)

**Settings** (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.testFlags": ["-v", "-cover"],
  "go.coverOnSave": true
}
```

---

### Pre-Commit Hook (Optional)

```bash
# .git/hooks/pre-commit
#!/bin/bash
set -e

echo "Running linter..."
./scripts/lint.sh

echo "Running tests..."
go test ./...

echo "Checking coverage..."
go test -cover ./... | grep "coverage: " | awk '{if ($2 < 80) exit 1}'
```

---

## Related Documents

- [Definition of Done](definition-of-done.md) - Quality criteria
- [LLM Prompts](llm-prompts.md) - AI prompt templates
- [Architecture Rules](.cursor/rules/) - LLM context
