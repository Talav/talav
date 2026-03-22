# Definition of Done

> **Quality criteria for code, documentation, and features**

---

## Overview

A feature/component/fix is "done" when it meets all applicable criteria below. This ensures consistent quality and prevents incomplete work from being merged.

---

## Code Quality

### Must Have (Blocking)

- [ ] **Compiles without errors**
  - `go build ./...` succeeds
  - No syntax errors or type mismatches

- [ ] **Passes all tests**
  - `go test ./...` passes
  - No flaky or skipped tests
  - Tests are deterministic

- [ ] **No linter errors**
  - `golangci-lint run` passes
  - Fix all errors (not just warnings)
  - Run: `./scripts/lint.sh`

- [ ] **Architecture constraints satisfied**
  - `./scripts/validate-architecture.sh` passes
  - No layering violations
  - Imports respect layer boundaries

### Should Have (Strong Recommendation)

- [ ] **Test coverage meets target**
  - Components: ≥80% coverage
  - FX modules: ≥60% coverage
  - Check: `go test -cover ./...`

- [ ] **No TODOs or FIXMEs**
  - Implementation is complete
  - Known issues documented in bug log
  - Future work tracked as issues, not comments

- [ ] **Error handling is complete**
  - All errors checked and handled
  - Panics only for invariant violations
  - Error messages are actionable

---

## Testing

### Must Have

- [ ] **Unit tests exist**
  - Test files: `*_test.go` in same package
  - Tests are independent (no shared state)
  - Tests follow naming: `Test[Component]_[Method]_[Scenario]`

- [ ] **Tests use assertions**
  - Use `testify/assert` or `testify/require`
  - Clear assertion messages
  - Example: `assert.Equal(t, expected, actual, "user ID should match")`

- [ ] **Happy path tested**
  - Main use case works
  - Expected inputs produce expected outputs

- [ ] **Error paths tested**
  - Invalid inputs handled
  - Edge cases covered (nil, empty, boundary values)
  - Errors are returned/handled correctly

### Should Have

- [ ] **Table-driven tests**
  - Multiple scenarios in one test function
  - Pattern: `tests := []struct{name, input, expected}{...}`

- [ ] **Mock external dependencies**
  - Database, HTTP, filesystem mocked
  - No test dependencies on external services

### Integration Tests (Optional)

- [ ] **Build tag used**
  - `//go:build integration` at top
  - Run: `go test -tags=integration ./...`

- [ ] **FX DI container used**
  - Test wiring, not just units
  - Ensure components work together

---

## Documentation

### Must Have

- [ ] **README exists**
  - Every component/module has `README.md`
  - Explains what it does (1-2 sentences)

- [ ] **Quick Start example**
  - Copy-paste code that works
  - Shows most common use case

- [ ] **API reference**
  - Key types and functions documented
  - Include parameters, returns, errors

### Should Have

- [ ] **Examples for common patterns**
  - Configuration
  - Error handling
  - Testing with mocks

- [ ] **Related components listed**
  - "See Also" section links to related docs

### Code Comments

- [ ] **Exported types documented**
  - Package doc comment (package-level)
  - Struct/function doc comments for exports
  - Format: `// TypeName does X. It is used for Y.`

- [ ] **Complex logic explained**
  - Why, not what (code shows what)
  - Example: `// Check cache first to avoid DB roundtrip`

---

## Architecture & Design

### Must Have

- [ ] **Follows layering rules**
  - Components don't import FX/framework
  - FX modules don't import framework
  - Validated by `./scripts/validate-architecture.sh`

- [ ] **Constructor injection used**
  - Dependencies passed to `New*()` functions
  - No global variables or singletons

- [ ] **Interfaces defined where needed**
  - Dependencies are interfaces, not concrete types
  - Enables mocking for tests

### Should Have

- [ ] **Function ordering follows convention**
  - Constructors first
  - Public methods next
  - Private helpers last
  - See: `.cursor/rules/02-function-ordering.mdc`

- [ ] **Naming conventions followed**
  - Packages: lowercase, singular (`user`, not `users`)
  - Types: PascalCase (`UserService`)
  - Functions: camelCase (`getUserByID`)
  - Private: lowercase (`validateInput`)

---

## API Design (Zorya Endpoints)

### Must Have

- [ ] **Input/output structs defined**
  - Struct tags for schema: `schema`, `validate`, `openapi`
  - Example: `Name string 'json:"name" validate:"required"`

- [ ] **Validation rules specified**
  - Use `validate` tags
  - Common: `required`, `email`, `min`, `max`, `oneof`

- [ ] **Error responses are RFC 9457 compliant**
  - Use `zorya.Error404NotFound()`, etc.
  - Errors include machine-readable codes

- [ ] **OpenAPI metadata added**
  - Operation summary and description
  - Field descriptions via `openapi` tags
  - Example values for documentation

### Should Have

- [ ] **Security requirements declared**
  - Use `zorya.Secure()` for protected routes
  - Specify roles/permissions/resources

- [ ] **Content negotiation supported**
  - Handles JSON (default)
  - CBOR if needed

---

## Configuration

### Must Have

- [ ] **Config struct defined**
  - Use `mapstructure` tags
  - Default values via `default` tag
  - Example: `Port int 'mapstructure:"port" default:"8080"`

- [ ] **Config validated**
  - Check required fields in constructor
  - Panic if invalid (fail-fast)

### Should Have

- [ ] **Environment variables supported**
  - Via Viper's automatic env binding
  - Follow naming: `COMPONENT_FIELD`

- [ ] **Example config in docs**
  - Show YAML structure
  - Include all fields with descriptions

---

## Database/Persistence (If Applicable)

### Must Have

- [ ] **Migrations provided**
  - SQL migration files for schema changes
  - Up and down migrations

- [ ] **Transactions used correctly**
  - Begin → operations → commit/rollback
  - No partial updates on error

- [ ] **Indexes defined**
  - Performance-critical queries indexed
  - Foreign keys have indexes

### Should Have

- [ ] **Seed data for development**
  - Use seeder component
  - Realistic test data (faker)

---

## Security (If Applicable)

### Must Have

- [ ] **No secrets in code**
  - Use environment variables or secret store
  - No hardcoded passwords, API keys, tokens

- [ ] **SQL injection prevented**
  - Use parameterized queries (GORM handles this)
  - Never concatenate user input into SQL

- [ ] **Input validation**
  - Validate all user input
  - Sanitize before use

### Should Have

- [ ] **Authentication enforced**
  - Protected routes use `zorya.Secure()`
  - Authorization checks in place

- [ ] **Rate limiting considered**
  - For public endpoints

---

## Git & Version Control

### Must Have

- [ ] **Commit message follows format**
  - Format: `type(scope): subject`
  - Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
  - Example: `feat(email): add HTML template support`

- [ ] **Atomic commits**
  - One logical change per commit
  - Commit compiles and tests pass

- [ ] **Branch from main**
  - Use feature branches: `feature/my-feature`
  - Keep branches short-lived (<1 week)

### Should Have

- [ ] **Pre-commit checks passed**
  - Architecture validation
  - Linter
  - Tests
  - Hook runs automatically (`.git/hooks/pre-commit`)

---

## Pull Request

### Must Have

- [ ] **PR description filled out**
  - What changed
  - Why (link to issue)
  - How to test

- [ ] **Checklist completed**
  - Tests pass
  - Linter clean
  - Coverage target met
  - README updated

- [ ] **CI passes**
  - All CI checks green
  - No failing tests or linter errors

### Should Have

- [ ] **Review from maintainer**
  - Code reviewed by someone familiar with codebase
  - Architecture review if significant changes

- [ ] **Related docs updated**
  - Decision log if architecture decision
  - Implementation log for significant changes
  - Bug log if fixing a bug

---

## Deployment (When Applicable)

### Must Have

- [ ] **Configuration deployed**
  - Environment-specific config updated
  - Secrets in place

- [ ] **Migrations run**
  - Database schema updated
  - Data migrations completed

- [ ] **Health check passes**
  - `/health` endpoint returns 200
  - All services responding

### Should Have

- [ ] **Monitoring in place**
  - Logs visible in logging system
  - Metrics being collected
  - Alerts configured

- [ ] **Rollback plan documented**
  - How to rollback if issues
  - What data needs cleanup

---

## Checklist Template

Copy this for your PRs:

```markdown
## Definition of Done

### Code Quality
- [ ] Compiles without errors
- [ ] All tests pass
- [ ] No linter errors
- [ ] Architecture constraints satisfied
- [ ] Test coverage ≥80% (components) or ≥60% (FX)

### Testing
- [ ] Unit tests exist
- [ ] Happy path tested
- [ ] Error paths tested
- [ ] External dependencies mocked

### Documentation
- [ ] README updated
- [ ] Quick Start example included
- [ ] API reference complete
- [ ] Code comments for exported types

### Architecture
- [ ] Follows layering rules
- [ ] Constructor injection used
- [ ] Function ordering follows convention

### Git
- [ ] Commit message follows format
- [ ] Pre-commit checks passed
- [ ] PR description complete

### Additional (if applicable)
- [ ] Config struct defined with defaults
- [ ] Migrations provided
- [ ] Security requirements met
- [ ] OpenAPI metadata added
```

---

## Exceptions

Sometimes "done" criteria can be relaxed:

### WIP Commits
- Can skip test coverage
- Can have TODOs
- Must note in commit message: `WIP: description`
- Don't merge to main

### Experimental Code
- Can skip documentation
- Must be in separate branch
- Clearly marked as experimental

### Hotfixes
- Can skip pre-commit checks (use `--no-verify`)
- Must fix critical production issue
- Follow up with proper fix + tests ASAP

---

## Related Documents

- [Development Workflow](dev-workflow.md) - Daily dev process
- [Architecture Validation](architecture-validation.md) - Automated checks
- [Testing Best Practices](../../.cursor/rules/03-testing.mdc) - Testing patterns
