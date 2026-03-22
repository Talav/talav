# LLM Prompts

> **Canonical prompts for AI-assisted development with Talav Framework**

---

## Overview

This document provides effective prompts for working with AI assistants (Cursor, GitHub Copilot, Claude) on Talav framework development. Each prompt is optimized for clarity and context.

**Key Principles**:
1. **Reference architecture rules** - Point AI to `.cursor/rules/` for boundaries
2. **Show examples** - Reference existing components for patterns
3. **Be specific** - State requirements, constraints, and expected structure
4. **Test-first** - Ask AI to write tests or implement from test cases

---

## Component Creation Prompts

### Create New Component (General)

```
Create a new component in pkg/component/[NAME] that implements [FUNCTIONALITY].

Requirements:
1. Component must follow layered architecture rules from .cursor/rules/
2. NO imports from pkg/fx/* or pkg/module/* (component layer only)
3. Use constructor injection for dependencies
4. Follow function ordering from 02-function-ordering.mdc
5. Include comprehensive tests with testify assertions
6. Add README.md explaining what it does and how to use it

Component should have:
- [Service/Repository/Factory] struct with [methods]
- Tests with 80%+ coverage
- Clear interface definitions

Example component to follow: pkg/component/[SIMILAR_COMPONENT]
```

---

### Create Component with FX Module

```
Create a complete component + FX module for [FUNCTIONALITY]:

**Component** (pkg/component/[NAME]):
1. [Service name] struct with these methods:
   - [Method 1]: [description]
   - [Method 2]: [description]
2. Configuration struct: [Name]Config with fields:
   - [field 1]: [type] - [description]
   - [field 2]: [type] - [description]
3. Tests covering all methods
4. NO imports from pkg/fx/* (component autonomy)

**FX Module** (pkg/fx/fx[NAME]):
1. FxModule definition that:
   - Provides [Service name] via constructor
   - Loads config from fxconfig.AsConfig("[name]", defaultConfig)
   - Registers Cobra command if needed
2. CAN import pkg/component/[NAME]
3. CANNOT import pkg/module/*

Reference these examples:
- pkg/component/email + pkg/fx/fxemail (for structure)
- pkg/component/logger + pkg/fx/fxlogger (for config pattern)
```

---

## API Development Prompts

### Add HTTP Endpoint (Zorya)

```
Add a new HTTP endpoint to the Zorya API:

**Endpoint**: [METHOD] [PATH]
**Functionality**: [what it does]

**Input Struct**:
- Path parameters: [list]
- Query parameters: [list]
- Headers: [list]
- Request body: [fields with types]
- Validation rules: [required, email, min, max, etc.]

**Output Struct**:
- Status code: [default code]
- Response body: [fields with types]
- Headers: [if any]

**Handler Logic**:
[describe business logic]

**Security** (if protected):
- Authentication: [required/not required]
- Roles: [list of allowed roles]
- Permissions: [list of required permissions]

Example endpoint to follow: [similar endpoint in codebase]

Use:
- zorya.Post/Get/Put/Delete for registration
- validate tags for validation
- openapi tags for documentation
- Secure() wrapper if protected
```

---

### Add Route Security

```
Protect existing route [PATH] with security requirements:

**Requirements**:
- Authentication: [yes/no]
- Roles: [admin, user, editor, etc.]
- Permissions: [resource:action format]
- Resource template: [e.g., "organizations/{orgId}/projects"]

Update route registration to use zorya.Secure() with appropriate options:
- zorya.Auth() for authentication
- zorya.Roles(...) for role-based access
- zorya.Permissions(...) for permission-based access
- zorya.ResourceFromParams(...) for dynamic resources

Example: [reference similar protected route]
```

---

## Refactoring Prompts

### Fix Layering Violation

```
Fix layering violation in [FILE]:

**Current Problem**:
- [Component/FX Module] is importing [pkg/fx/*/pkg/module/*]
- This violates architecture rules from .cursor/rules/

**Constraints**:
- Components (pkg/component/*) can ONLY import other components
- FX modules (pkg/fx/*) can ONLY import components
- Application (pkg/module/*) can import everything

**How to Fix**:
1. If component needs FX/framework service → receive via constructor parameter
2. If FX module needs framework service → receive via DI (fx.In struct)
3. Move framework-specific code OUT of component INTO FX module

Reference .cursor/rules/ for detailed layering rules.
```

---

### Refactor to Use Dependency Injection

```
Refactor [COMPONENT] to use dependency injection instead of [CURRENT_PATTERN]:

**Current Pattern** (problematic):
[show current code - e.g., global variables, singletons, direct instantiation]

**Target Pattern**:
1. Accept dependencies via constructor parameters
2. Define clear interfaces for dependencies
3. Use constructor injection, not method injection

**Example to Follow**:
Look at pkg/component/httpserver/server.go:
- Dependencies (API, logger) passed to constructor
- No global state
- Testable with mocks

Convert [COMPONENT] to follow this pattern.
```

---

## Testing Prompts

### Write Tests for Existing Code

```
Write comprehensive tests for [COMPONENT/FUNCTION]:

**Requirements**:
1. Use testify/assert and testify/require
2. Table-driven tests for multiple scenarios
3. Test both happy path and error cases
4. Coverage target: 80%+

**Test Cases to Include**:
- [Scenario 1]: [expected behavior]
- [Scenario 2]: [edge case]
- [Error case 1]: [how it should fail]

**Pattern to Follow**:
Look at [SIMILAR_COMPONENT]_test.go for structure:
- Helper functions at top
- Test functions in logical groups
- Table-driven with subtests

Mock external dependencies using interfaces.
```

---

### Add Integration Test

```
Create integration test for [FEATURE]:

**Setup**:
1. Use FX to wire dependencies (DI container approach)
2. Use testcontainers if database needed
3. Use httptest if HTTP server needed

**Test Flow**:
1. [Setup step 1]
2. [Action to test]
3. [Assertion]
4. [Cleanup]

**Example**:
Reference existing integration tests in [COMPONENT]_integration_test.go
Use build tag: //go:build integration
```

---

## Documentation Prompts

### Write Component README

```
Write comprehensive README for pkg/component/[NAME]:

**Structure**:
1. One-sentence description
2. Features (bullet list)
3. Installation (go get command)
4. Quick Start (copy-paste example)
5. API Reference (key types and functions with descriptions)
6. Configuration (if applicable)
7. Examples (common use cases)
8. See Also (related components)

**Tone**: Technical, concise, code-heavy (more examples, less prose)

**Reference**: Look at pkg/component/zorya/README.md for excellent example
```

---

### Add OpenAPI Documentation

```
Enhance OpenAPI documentation for [ENDPOINT]:

Add these struct tags:
1. `openapi` tags for field metadata:
   - title: Short field description
   - description: Longer explanation
   - example: Example value
   - format: date, email, uri, etc.
   - deprecated: Mark deprecated fields

2. `validate` tags for constraints (auto-documented):
   - required, email, min, max, pattern, oneof

3. Operation-level docs:
   - Summary: One-line description
   - Description: Detailed explanation
   - Tags: Grouping (users, posts, etc.)

Example endpoint with good docs: [reference]
```

---

## Debugging Prompts

### Diagnose Test Failure

```
Test [TEST_NAME] in [FILE] is failing with error:
[PASTE ERROR]

**Context**:
- [What the test is trying to do]
- [What's expected vs actual]

**Analyze**:
1. Read the test code
2. Read the implementation
3. Identify the mismatch
4. Suggest fix with explanation

Use verbose test output: go test -v -run [TEST_NAME] ./...
```

---

### Fix Module Resolution Error

```
Getting module resolution error:
[PASTE ERROR]

Context:
- Using Go workspaces (go.work)
- Modules in pkg/component/* and pkg/fx/*

**Diagnose**:
1. Check if go.work includes the module
2. Check if require directive in go.mod is correct
3. Run go work sync to update go.work.sum
4. Run go mod tidy in the module directory

**Expected versions**:
- Workspace modules should use pseudo-versions like v0.0.0-DATE-COMMIT
- Run: go list -m all | grep talav to see what's imported

Provide commands to fix the issue.
```

---

## Prompt Templates by Task

### "I want to build a REST API"

```
Help me build a REST API with Talav:

**API Purpose**: [what the API does]
**Entities**: [User, Post, Comment, etc.]
**Endpoints**: [list main endpoints]
**Auth**: [authentication method]
**Database**: [PostgreSQL/MySQL/SQLite]

**Generate**:
1. Project structure (main.go with framework bootstrap)
2. Domain models for each entity
3. Repository interfaces and implementations (GORM)
4. Service layer with business logic
5. Zorya HTTP handlers with validation
6. FX modules for DI wiring
7. Configuration (config.yaml)
8. README with setup instructions

Follow Talav patterns from docs/04-process/dev-workflow.md
```

---

### "I want to add authentication"

```
Add JWT-based authentication to existing Talav API:

**Current State**:
- API has endpoints in pkg/component/[NAME]
- No authentication yet

**Requirements**:
1. JWT token generation on login
2. Middleware to validate JWT on protected routes
3. Extract user from token, store in context
4. Protect routes using zorya.Secure(zorya.Auth())

**Implementation Steps**:
1. Create pkg/component/auth with:
   - JWT service (generate, validate tokens)
   - Middleware that checks Authorization header
   - Context helpers (GetAuthUser, SetAuthUser)
2. Create pkg/fx/fxauth to wire it up
3. Register middleware in API bootstrap
4. Update protected routes with zorya.Secure()

Reference: pkg/component/security for patterns
```

---

## Related Documents

- [Development Workflow](dev-workflow.md) - Daily dev process
- [Architecture Rules](/.cursor/rules/) - Layering constraints
- [Function Ordering](/.cursor/rules/02-function-ordering.mdc) - Code organization
- [Testing Best Practices](/.cursor/rules/03-testing.mdc) - Testing patterns
