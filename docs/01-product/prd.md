# Product Requirements Document (PRD)

> **Single source of truth for WHAT Talav Framework must do**

---

## Overview

**Product Name**: Talav Go Framework

**Version**: 0.1.0-alpha

**Last Updated**: 2026-01-14

**Status**: Draft / In Development

### Executive Summary

Talav is a modular Go framework for building production-ready HTTP APIs with enforced layered architecture and type-safe request/response handling. It provides the "missing middle ground" between raw HTTP libraries and full-stack frameworks, optimizing for developer productivity (especially with LLM assistance) while maintaining clean architecture principles.

Core value: **Ship type-safe REST APIs 3-5x faster with OpenAPI generation, automatic validation, and enforced component boundaries.**

## Problem Statement

### User Pain Points

1. **Boilerplate Hell**
   - Who experiences it: Go developers building REST APIs
   - Current workaround: Copy-paste handlers, write manual marshaling/unmarshaling, duplicate validation logic
   - Impact: 60-70% of code is repetitive plumbing, slowing development and introducing bugs

2. **Framework Lock-In vs No Structure**
   - Who experiences it: Teams choosing between raw HTTP (too minimal) and full-stack frameworks (too prescriptive)
   - Current workaround: Either accept lock-in or build custom abstractions
   - Impact: Frameworks couple business logic to framework code, making testing/reusability hard

3. **Manual OpenAPI Maintenance**
   - Who experiences it: API developers maintaining separate OpenAPI specs
   - Current workaround: Write specs by hand, sync with code manually, or use complex code generation
   - Impact: Specs diverge from implementation, wasting 2-4 hours/week keeping them in sync

4. **Layering Violations and Circular Dependencies**
   - Who experiences it: Teams with growing codebases and LLM-generated code
   - Current workaround: Manual code review, refactoring sprints
   - Impact: Technical debt accumulates, onboarding slows, AI generates inconsistent code

### Success Criteria

| Metric | Current (Baseline) | Target | Measure |
|--------|-------------------|--------|---------|
| **Lines of Code for CRUD API** | 500-600 lines (raw Chi) | 150-200 lines (Talav) | Compare equivalent APIs |
| **Time to First API** | 4-6 hours | 30-60 minutes | From `go mod init` to working `/users` endpoint |
| **OpenAPI Spec Sync Time** | 2-4 hours/week | 0 hours (auto-generated) | Manual maintenance eliminated |
| **Layering Violations** | 10-20/1000 LOC (typical) | 0 (enforced by linter) | Static analysis |
| **Test Coverage** | 40-60% (manual tests) | 80%+ (test utilities) | `go test -cover` |

## Users & Use Cases

### Target Users

**Primary Persona: Mid-Level Backend Developer**
- **Who they are**: 2-5 years Go experience, building REST APIs for web/mobile apps
- **Goals**: Ship features fast without sacrificing code quality
- **Frustrations**:
  - Too much boilerplate in handlers
  - OpenAPI specs go stale
  - Unclear where to put business logic (handlers? services? repositories?)
  - Testing requires complex setup
- **Technical level**: Intermediate - knows Go idioms, understands HTTP, may not know advanced DI patterns

**Secondary Persona: Team Using LLM Assistants (Cursor, Copilot)**
- **Who they are**: Variable experience (junior to senior), leveraging AI for faster development
- **Goals**: Let AI generate code that actually works and follows best practices
- **Frustrations**:
  - AI generates code that breaks layering
  - Inconsistent patterns across files
  - AI doesn't understand project-specific conventions
- **Technical level**: Variable - AI fills knowledge gaps, but need guardrails

### Core Use Cases

#### Use Case 1: Create Type-Safe CRUD API

**Actor**: Backend developer

**Preconditions**:
- Go 1.25+ installed
- Basic understanding of REST APIs and struct tags

**Main Flow**:
1. Developer installs framework: `go get github.com/talav/talav`
2. Developer creates `main.go` with `framework.RunDefault(fxhttpserver.FxHTTPServerModule)`
3. Developer defines input/output structs with tags:
   ```go
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
   ```
4. Developer registers handler: `zorya.Post(api, "/users", createUserHandler)`
5. System auto-generates OpenAPI spec at `/openapi.json`
6. Developer runs: `go run main.go serve-http`
7. API is live with validation, error handling, and docs

**Postconditions**:
- Working CRUD API with type-safe handlers
- OpenAPI spec auto-generated
- Validation enforced on requests
- RFC 9457 error responses

**Alternative Flows**:
- **Alt 1**: Developer uses FX modules for DI instead of manual wiring
- **Alt 2**: Developer adds custom validation rules via validator registry

#### Use Case 2: Enforce Layered Architecture in Growing Codebase

**Actor**: Tech lead on team of 5+ developers

**Preconditions**:
- Existing codebase with 10K+ LOC
- Team uses LLMs for code generation
- Codebase has layering violations (components import FX, circular deps)

**Main Flow**:
1. Tech lead adds `.cursor/rules/` with layering rules
2. Tech lead runs `scripts/lint.sh` to check all modules
3. Linter reports violations (e.g., "component imports pkg/fx/")
4. Team fixes violations by refactoring
5. Tech lead adds linter to CI/CD pipeline
6. Future PRs with violations fail CI
7. LLM-generated code respects boundaries (reads rules)

**Postconditions**:
- Zero layering violations
- CI enforces layering on every commit
- AI assistants follow architecture rules

#### Use Case 3: Add Authentication and Role-Based Access Control

**Actor**: Backend developer

**Preconditions**:
- Existing API with public endpoints
- Need to protect `/admin/*` routes with role checks

**Main Flow**:
1. Developer adds security middleware:
   ```go
   api.UseMiddleware(security.NewSecurityMiddleware(enforcer))
   ```
2. Developer wraps protected routes:
   ```go
   zorya.Get(api, "/admin/users", handler,
       zorya.Secure(zorya.Roles("admin")),
   )
   ```
3. System injects security metadata into context
4. Security middleware enforces requirements
5. Unauthorized users receive `401 Unauthorized`
6. Users without required role receive `403 Forbidden`
7. OpenAPI spec shows security requirements

**Postconditions**:
- Auth/authz enforced declaratively
- Security metadata in OpenAPI
- No changes to handler logic

## Requirements

### Functional Requirements

#### Must Have (P0)

- [ ] **FR-001: Type-Safe HTTP Handlers**
  - **Rationale**: Core framework value - eliminate manual marshaling/unmarshaling
  - **Acceptance Criteria**:
    - Generic `zorya.Register[I, O any](api, route, handler)` API works for any struct types
    - Input struct decoded from path, query, header, cookie, body based on `schema` tags
    - Output struct encoded to response format (JSON/CBOR) based on `Accept` header
    - Compile-time type safety (no `interface{}` in user code)

- [ ] **FR-002: Automatic Request Validation**
  - **Rationale**: Prevent invalid data from reaching business logic
  - **Acceptance Criteria**:
    - go-playground/validator tags (`required`, `email`, `min`, `max`, etc.) enforced automatically
    - Validation errors return `422 Unprocessable Entity` with structured error details
    - Custom validators can be registered
    - Nested struct validation works

- [ ] **FR-003: OpenAPI 3.1 Generation**
  - **Rationale**: Auto-generated docs eliminate manual maintenance
  - **Acceptance Criteria**:
    - `/openapi.json` and `/openapi.yaml` endpoints serve spec
    - Spec includes all routes, request/response schemas, validation constraints
    - Spec respects `openapi` struct tags (title, description, example, deprecated)
    - `/docs` endpoint serves interactive API documentation

- [ ] **FR-004: Enforced Layered Architecture**
  - **Rationale**: Prevent technical debt and circular dependencies
  - **Acceptance Criteria**:
    - Components (`pkg/component/*`) cannot import `pkg/fx/*` or `pkg/module/*`
    - FX modules (`pkg/fx/*`) cannot import `pkg/module/*`
    - Linter (`golangci-lint`) detects violations
    - `.cursor/rules/` documents architecture for LLMs

- [ ] **FR-005: CLI Application Bootstrap**
  - **Rationale**: Provide entry point for HTTP servers, migrations, seeders, etc.
  - **Acceptance Criteria**:
    - `framework.NewApplication()` creates app with Cobra root command
    - FX modules register commands via `fxcore.AsRootCommand`
    - Signal handling (`SIGINT`, `SIGTERM`) propagated to all commands
    - FX container auto-cleans up on shutdown

#### Should Have (P1)

- [ ] **FR-101: Declarative Route Security**
  - **Rationale**: Authentication/authorization is common requirement
  - **Acceptance Criteria**:
    - `zorya.Secure(zorya.Auth(), zorya.Roles("admin"))` declares requirements
    - Security middleware enforces via pluggable enforcer interface
    - Built-in enforcers: `SimpleEnforcer`, `CasbinEnforcer`
    - OpenAPI spec shows security requirements

- [ ] **FR-102: Configuration Management**
  - **Rationale**: Apps need environment-specific config (dev, staging, prod)
  - **Acceptance Criteria**:
    - Load from YAML files or environment variables
    - Type-safe config structs via `mapstructure`
    - Default values via `default` struct tag
    - `APP_ENV` switches between environments

- [ ] **FR-103: Structured Logging**
  - **Rationale**: Production apps need machine-readable logs
  - **Acceptance Criteria**:
    - slog-based logger with JSON/text output
    - HTTP request logging with request ID, latency, status
    - Log levels (debug, info, warn, error) configurable
    - Log schemas (standard, ECS, OTEL, GCP) supported

- [ ] **FR-104: Database Integration (ORM)**
  - **Rationale**: Most APIs need database persistence
  - **Acceptance Criteria**:
    - GORM integration via `pkg/component/orm`
    - Connection pooling, migration runner
    - PostgreSQL, MySQL, SQLite drivers
    - Transaction support

#### Nice to Have (P2)

- [ ] **FR-201: Streaming Responses (SSE, Chunked)**
  - **Rationale**: Real-time use cases (notifications, progress updates)
  - **Acceptance Criteria**:
    - `Body func(ctx zorya.Context)` field in output struct enables streaming
    - SSE example in docs

- [ ] **FR-202: Conditional Requests (ETag, If-Match)**
  - **Rationale**: Caching and optimistic concurrency control
  - **Acceptance Criteria**:
    - `conditional.Params` embeddable in input struct
    - Returns `304 Not Modified` or `412 Precondition Failed` when appropriate

- [ ] **FR-203: Content Negotiation for Custom Formats (XML, YAML)**
  - **Rationale**: Some APIs need non-JSON formats
  - **Acceptance Criteria**:
    - `zorya.WithFormat("application/xml", xmlFormat)` registers custom format
    - `Accept` header negotiates format

### Non-Functional Requirements

#### Performance

- [ ] **NFR-001: Low Request Overhead**
  - **Metric**: Added latency from framework (decoding + validation + encoding)
  - **Target**: <1ms p50, <5ms p99 for simple CRUD operations
  - **Rationale**: Framework should be "fast enough to not think about"

- [ ] **NFR-002: Metadata Caching**
  - **Metric**: Reflection overhead per request
  - **Target**: Struct metadata cached, <0.1ms lookup per request
  - **Rationale**: Reflection is slow, caching mitigates cost

#### Security

- [ ] **NFR-101: No SQL Injection via ORM**
  - **Rationale**: GORM uses parameterized queries
  - **Compliance**: OWASP Top 10 - A03:2021

- [ ] **NFR-102: Password Hashing with bcrypt**
  - **Rationale**: User component stores hashed passwords
  - **Compliance**: OWASP password storage

- [ ] **NFR-103: Dependency Vulnerability Scanning**
  - **Rationale**: Third-party deps may have CVEs
  - **Compliance**: Run `govulncheck` in CI

#### Usability

- [ ] **NFR-201: Comprehensive README Files**
  - **Metric**: Developer can build first API in <60 minutes without external help
  - **Target**: README + examples sufficient for onboarding

- [ ] **NFR-202: Clear Error Messages**
  - **Metric**: Error messages include actionable next steps
  - **Target**: Validation errors show field, constraint, and location (e.g., `"body.email": must be valid email`)

#### Reliability

- [ ] **NFR-301: Graceful Shutdown**
  - **Metric**: In-flight requests complete before shutdown
  - **Target**: Zero dropped requests on `SIGTERM`

- [ ] **NFR-302: Test Coverage**
  - **Metric**: `go test -cover` for all packages
  - **Target**: 80%+ coverage for components, 60%+ for FX modules

### Constraints

**Technical Constraints**:
- Go 1.25+ required (uses new `range`-over-func, `ServeMux` patterns)
- Uber FX for DI (no support for other DI frameworks)
- Chi/Fiber/stdlib routers only (other routers via community adapters)
- Linux/macOS for development (Windows may have path issues)

**Business Constraints**:
- Solo maintainer (limited bandwidth for features)
- MIT license (permissive, commercial-friendly)
- No commercial support (community-driven only)

**Regulatory/Compliance**:
- None currently (framework is infrastructure, not end-user app)

## User Experience

### User Journey

```
[Install Framework] → [Create main.go] → [Define Structs] → [Register Routes] → [Run Server] → [Test API] → [Deploy]
```

**Detailed Steps**:
1. Developer runs `go get github.com/talav/talav`
2. Developer creates `main.go` with framework bootstrap
3. Developer defines input/output structs with tags
4. Developer registers routes via `zorya.Get/Post/Put/Delete`
5. Developer runs `go run main.go serve-http`
6. Developer tests API via curl/Postman
7. Developer visits `/docs` for interactive documentation
8. Developer deploys binary to production

### Key Screens/Interactions

1. **CLI Help Screen** (`myapp --help`)
   - Purpose: Show available commands
   - Key elements: Command list, flags, usage examples
   - Actions: Run `myapp <command>`

2. **OpenAPI JSON Spec** (`GET /openapi.json`)
   - Purpose: Provide machine-readable API spec
   - Key elements: Paths, schemas, responses, security
   - Actions: Import into API clients, generate SDKs

3. **API Documentation UI** (`GET /docs`)
   - Purpose: Interactive API testing
   - Key elements: Operation list, request builder, response viewer
   - Actions: Test endpoints, see examples

### Error States

| Scenario | User Experience | System Behavior |
|----------|----------------|-----------------|
| Validation failure | `422 Unprocessable Entity` with field-level errors | Return `{"status": 422, "errors": [{"code": "email", "location": "body.email", "message": "..."}]}` |
| Missing required field | `422 Unprocessable Entity` | Validation catches, returns error detail |
| Unauthorized request | `401 Unauthorized` | Security middleware blocks request |
| Forbidden (insufficient roles) | `403 Forbidden` | Security middleware blocks request |
| Internal server error | `500 Internal Server Error` with request ID | Log error with stack trace, return generic message |
| FX initialization failure | Panic with error message | Application fails to start (fail-fast) |

## Scope

### In Scope

- **Core Framework**: Application bootstrap, CLI, FX lifecycle
- **Zorya API Framework**: Type-safe handlers, validation, OpenAPI generation
- **HTTP Server**: Chi-based server with logging, request IDs
- **Components**: Config, logger, validator, ORM, schema, user, email, blob, seeder
- **Security**: Declarative route security, enforcers (simple, Casbin)
- **Documentation**: README files, examples, architecture rules for LLMs

### Out of Scope

- **Frontend Frameworks**: React, Vue, etc. (separate concern)
- **WebSockets/Real-Time**: gRPC, WebSocket support (future consideration)
- **Background Jobs**: Queue systems, cron schedulers (future consideration)
- **API Gateway Features**: Rate limiting (can be added via middleware), API keys (use security component)
- **Multi-Tenancy**: Database-per-tenant, schema-per-tenant (user responsibility)

### Future Considerations

- **GraphQL Support**: Zorya adapter for GraphQL (investigate feasibility)
- **Server-Sent Events**: Streaming improvements for long-lived connections
- **Metrics/Tracing**: Prometheus/OpenTelemetry integration
- **Admin UI**: Auto-generated CRUD interface for models

## Dependencies

### Internal Dependencies

- `pkg/component/schema` needed by `pkg/component/zorya` for request decoding
- `pkg/component/zorya` needed by `pkg/component/httpserver` for API framework
- `pkg/fx/fxcore` needed by all FX modules for command registration
- `pkg/fx/fxconfig` needed by most FX modules for configuration loading

### External Dependencies

- **Uber FX**: Maintained by Uber (active as of Jan 2026), risk: low
- **go-playground/validator**: Maintained by community (very active), risk: low
- **Chi router**: Maintained by go-chi org (active), risk: low
- **GORM**: Maintained by community (very active), risk: low

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Breaking API changes needed** | High | High | - Establish 1.0 stability guarantee<br>- Semantic versioning<br>- Migration guides |
| **Performance regression** | High | Medium | - Add benchmarks to CI<br>- Profile hot paths<br>- Set SLOs |
| **Security vulnerability** | High | Medium | - `govulncheck` in CI<br>- Pin dependencies<br>- Monitor advisories |
| **Solo maintainer burnout** | High | High | - Focus on quality over quantity<br>- Accept contributions<br>- Document everything |
| **FX adds too much complexity** | Medium | Medium | - Provide non-FX examples<br>- Document patterns<br>- Make FX optional (future) |

## Open Questions

- [ ] **Q1: Should security enforcement be opinionated or fully pluggable?**
  - **Impact**: Affects API surface and flexibility
  - **Owner**: Framework maintainer
  - **Deadline**: Beta release (Q2 2026)

- [ ] **Q2: How to handle database migrations in production?**
  - **Impact**: Users need production-safe migration strategy
  - **Owner**: ORM component maintainer
  - **Deadline**: Beta release (Q2 2026)

- [ ] **Q3: Should components be published as separate Go modules?**
  - **Impact**: Versioning and dependency management
  - **Owner**: Framework maintainer
  - **Deadline**: Before 1.0 (Q4 2026)

## Appendix

### Related Documents

- [Vision](../00-context/vision.md) - Why this product exists
- [System Map](../00-context/system-map.md) - Current architecture
- [Assumptions](../00-context/assumptions.md) - Risks and unknowns
- [Decision Log](../03-logs/decision-log.md) - Key decisions made

### Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2026-01-14 | 0.1 | Initial draft based on existing codebase | AI Assistant |
