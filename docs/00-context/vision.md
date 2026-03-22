# Vision

## WHY does this product exist?

**Problem:**
- Modern Go API development requires assembling many components: routing, validation, serialization, dependency injection, configuration, logging, and OpenAPI generation
- Existing frameworks are either too prescriptive (full-stack frameworks) or too minimal (raw HTTP), with no middle ground
- Most frameworks tightly couple business logic to framework abstractions, making testing and reusability difficult
- HTTP API frameworks lack type-safe request/response handling, forcing developers to write repetitive marshaling/unmarshaling code
- Layered architecture enforcement is manual and error-prone, leading to circular dependencies and tight coupling

**Vision:**
- **Talav** is a modular Go framework for building production-ready HTTP APIs with a strict layered architecture
- Provides the "missing middle ground": opinionated structure without framework lock-in
- Enforces clean separation between components (business logic), FX modules (dependency wiring), and application layer (orchestration)
- Delivers type-safe, RFC-compliant APIs with minimal boilerplate through the Zorya API framework
- Makes LLM-assisted development productive through clear patterns, comprehensive documentation, and enforced boundaries

**Impact:**
- Developers build APIs 3-5x faster with type-safe handlers, automatic validation, and OpenAPI generation
- Codebases remain maintainable and testable due to strict layering and component autonomy
- Teams onboard new developers faster with consistent patterns and self-documenting architecture
- AI assistants are highly effective due to clear structure, comprehensive docs, and predictable patterns

## WHAT exists RIGHT NOW?

**Product Boundaries:**
- **In Scope**: Go HTTP API development with Chi/Fiber routers, Cobra CLI, Uber FX DI, structured logging, configuration management, ORM integration
- **Out of Scope**: Frontend frameworks, mobile development, non-HTTP protocols (gRPC, WebSockets - future consideration), serverless deployments (future)
- **User Journey**: Developer installs framework → creates application with modules → registers components → builds type-safe API endpoints → deploys CLI application

**Current Capabilities:**
- **Framework Core** (`pkg/module/framework`): Application bootstrapping, Cobra CLI integration, FX lifecycle management, signal handling
- **Zorya API Framework** (`pkg/component/zorya`): Type-safe HTTP handlers, automatic request/response marshaling, content negotiation (JSON/CBOR), RFC 9457 error handling, OpenAPI 3.1 generation, route security declarations
- **HTTP Server** (`pkg/component/httpserver`): Chi-based server with structured logging, request IDs, Zorya integration
- **Components**: Config (YAML/ENV), Logger (slog-based), Validator (go-playground/validator), ORM (GORM), Schema (request parameter decoding), Mapstructure (type conversion), Email (go-mail), Blob storage (gocloud.dev), Seeder (database seeding), Tag parser (struct tag parsing)
- **Security**: Declarative route security (roles, permissions, resources), Simple enforcer, Casbin enforcer integration
- **User System**: User component with repositories (user, role, permission), services, and FX integration
- **Media System**: Media component with storage abstraction and FX integration

**Key Metrics:**
- **Lines of Code Reduction**: 60-70% less boilerplate vs raw Chi/HTTP (measured by handler code)
- **Type Safety**: 100% type-safe request/response handling (no `interface{}` in user code)
- **Testing**: 80%+ test coverage target for all components
- **Performance**: <1ms request overhead from Zorya framework (measured via benchmarks)
- **Adoption**: Internal use in 3-5 projects (current stage: alpha/beta)

## Anchor Points

**Product Principles:**

1. **Layered Architecture is Non-Negotiable**
   - Components (`pkg/component/*`) are pure business logic with no framework/FX dependencies
   - FX modules (`pkg/fx/*`) wire components together but contain no business logic
   - Application layer (`pkg/module/*`, `main.go`) orchestrates everything
   - Violations fail linting and code review
   - *Why it matters*: Ensures testability, reusability, and prevents circular dependencies

2. **Type Safety First**
   - All HTTP request/response handling is type-safe via Zorya's generic `Register[I, O any]` API
   - Validation happens at compile-time (struct tags) + runtime (go-playground/validator)
   - No `interface{}`, `map[string]any`, or reflection in user code
   - *Why it matters*: Catches errors early, enables IDE autocomplete, makes refactoring safe

3. **RFC Compliance Over Convenience**
   - OpenAPI 3.1 spec (RFC 9457 for errors, RFC 9110 for HTTP semantics)
   - Structured errors with machine-readable codes and locations
   - Content negotiation per HTTP standards
   - *Why it matters*: APIs are predictable, interoperable, and production-grade by default

4. **Components Own Their Lifecycle**
   - Components define their own Cobra commands with no framework dependencies
   - Commands manage their own startup/shutdown via `cmd.Context()`
   - FX just provides dependency injection, not lifecycle management
   - *Why it matters*: Components are autonomous, reusable across projects, and easily testable

5. **Optimize for LLM-Assisted Development**
   - Clear module boundaries and import rules
   - Comprehensive README files with examples
   - Consistent patterns across all components
   - Struct tags over configuration files where possible
   - *Why it matters*: AI assistants can navigate, understand, and generate code effectively

**Target Users:**

- **Primary Persona**: Backend Go developers building REST APIs for web/mobile applications
  - Experience level: Intermediate to senior (2+ years Go)
  - Goals: Ship production-ready APIs quickly without sacrificing maintainability
  - Frustrations: Boilerplate code, framework lock-in, poor layering, manual OpenAPI maintenance
  - Needs: Type safety, automatic validation, testing utilities, clear patterns

- **Secondary Persona**: Teams using LLMs for development (Cursor, GitHub Copilot)
  - Experience level: Variable (junior to senior)
  - Goals: Leverage AI for faster development while maintaining code quality
  - Frustrations: AI generates inconsistent code, breaks layering, creates circular dependencies
  - Needs: Clear structure that AI can follow, enforced boundaries, comprehensive docs

**Strategic Constraints:**

- **Technical Constraints**:
  - Go 1.25+ required (uses new ServeMux patterns, range-over-func)
  - Chi v5 or Fiber v2 for routing (other routers via adapter pattern)
  - Uber FX for dependency injection (alternative DI frameworks not supported)
  - No external API gateway required (security enforcement is in-process)

- **Business Constraints**:
  - Solo-developer maintained (limited bandwidth)
  - Focus on depth (perfect a few use cases) over breadth (support everything)
  - MIT license (open source, permissive)
  - No commercial support commitment (community-driven)

- **Timeline/Resource Constraints**:
  - Alpha stage: Core features stable, API may change
  - Beta target: Q2 2026 (stabilize APIs, comprehensive testing)
  - 1.0 target: Q4 2026 (API stability guarantee, production-ready)
