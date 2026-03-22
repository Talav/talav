# Decision Log

> **Architectural & product decisions with context and outcomes**

---

## Purpose

This log captures significant decisions made during Talav framework development. Each decision documents:
- The problem/context that led to the decision
- Options considered with pros/cons
- Decision made and rationale
- Expected outcomes and actual results

---

### [DEC-001] - Merge Kernel into Application (Framework Simplification)

**Date:** 2026-01-13

**Status:** Implemented

**Decision Makers:** Framework maintainer

**Context:**
The framework had a `Kernel` struct that was a thin wrapper around `fx.App`, adding indirection without clear benefits. The `Application` struct delegated to `Kernel` for FX lifecycle management, but this split responsibilities unnecessarily.

**Problem Statement:**
Should we keep the `Kernel`/`Application` separation or merge them for simplicity?

**Options Considered:**

#### Option 1: Keep Kernel Separate
**Pros:**
- Separation of concerns (Application = CLI, Kernel = DI)
- Could support alternative DI frameworks in future

**Cons:**
- Added complexity with no current benefit
- Extra indirection in code
- More files to maintain
- No plans for alternative DI frameworks

#### Option 2: Merge Kernel into Application
**Pros:**
- Simpler codebase (fewer structs, less indirection)
- Clearer ownership of FX lifecycle
- Easier to understand for new contributors
- Removes unused abstractions

**Cons:**
- If we ever need alternative DI, would need refactoring

**Decision:**
We chose **Option 2: Merge Kernel into Application**

**Rationale:**
- YAGNI principle: No need for abstraction until we have 2+ implementations
- Simpler code is easier to maintain and understand
- FX is deeply integrated; swapping it out would require major refactoring regardless
- Can add abstraction later if needed (YAGNI)

**Implications:**
- `Kernel` type removed
- `Application` directly manages `fxApp *fx.App` field
- `initFX()` method moved from Kernel to Application
- Getter methods removed (tests access fields directly)

**Success Criteria:**
- Codebase is easier to understand
- No functional changes to API

**Actual Outcome:**
✅ Merged successfully in commit `77ff54f`. Reduced code by ~100 LOC. No breaking changes to user-facing API.

---

### [DEC-002] - Panic on FX Initialization Failure (Fail-Fast)

**Date:** 2026-01-13

**Status:** Implemented

**Decision Makers:** Framework maintainer

**Context:**
`NewApplication()` was deferring FX initialization errors via `initErr` field. Errors were only discovered when calling methods on `Application`, making debugging harder.

**Problem Statement:**
Should `NewApplication()` panic immediately on FX errors or store errors for later retrieval?

**Options Considered:**

#### Option 1: Store Errors, Return Them Later
**Pros:**
- Non-panicking API (idiomatic Go)
- Caller can handle errors gracefully

**Cons:**
- FX errors are configuration/wiring mistakes (programming errors)
- Delaying error discovery makes debugging harder
- Requires all methods to check `initErr` first

#### Option 2: Panic Immediately
**Pros:**
- Fail-fast on configuration errors
- Clear error message at startup
- Configuration errors should never reach production (caught in dev/test)
- Simplifies `Application` code (no error field)

**Cons:**
- Panicking is not idiomatic Go
- No way to recover from FX errors

**Decision:**
We chose **Option 2: Panic Immediately**

**Rationale:**
- FX initialization errors are **programming mistakes** (missing dependencies, circular deps, invalid config)
- These should never happen in production if tests/dev environments work
- Panicking makes the error obvious and forces immediate fix
- Aligns with "fail-fast" philosophy for configuration errors
- Go's `net/http.ListenAndServe` also panics on invalid config (precedent)

**Implications:**
- `NewApplication()` panics if `initFX()` fails
- `initErr` field removed from `Application`
- Users must catch panics if they want to handle FX errors programmatically

**Success Criteria:**
- FX errors are immediately visible
- No need to check error returns throughout `Application`

**Actual Outcome:**
✅ Implemented in commit `77ff54f`. Errors now fail immediately with clear message. No issues reported.

---

### [DEC-003] - Remove RegisterLifecycle (Component Autonomy)

**Date:** 2026-01-14

**Status:** Proposed

**Decision Makers:** Framework maintainer

**Context:**
`pkg/fx/fxhttpserver/module.go` has a `RegisterLifecycle()` function that uses FX lifecycle hooks to start/stop the HTTP server. However, the module comment states "no automatic lifecycle management - commands control lifecycle." The function is never called.

**Problem Statement:**
Should HTTP server lifecycle be managed by FX hooks or by commands?

**Options Considered:**

#### Option 1: FX Lifecycle Hooks (RegisterLifecycle)
**Pros:**
- Automatic startup/shutdown
- No need for commands

**Cons:**
- Violates component autonomy principle
- Server starts even if user just wants `--help`
- Couples server lifecycle to FX, not commands
- Contradicts module documentation

#### Option 2: Command-Based Lifecycle (Current)
**Pros:**
- Component autonomy: `serve-http` command owns lifecycle
- Lazy initialization: FX boots only when command runs
- Signal handling via command context (framework-provided)
- Aligns with documented architecture

**Cons:**
- Requires users to run commands (not really a con - this is intentional)

**Decision:**
We choose **Option 2: Command-Based Lifecycle**

**Rationale:**
- Component autonomy is a core principle (components manage own lifecycle)
- Commands provide natural control points (start server, run migrations, etc.)
- FX should only provide DI, not lifecycle management
- `RegisterLifecycle` is dead code (never called)

**Implications:**
- Remove `RegisterLifecycle()` function
- Update any docs mentioning automatic lifecycle
- Ensure all components follow command-based pattern

**Success Criteria:**
- HTTP server starts only via `serve-http` command
- FX provides DI, commands manage lifecycle

**Actual Outcome:**
🚧 Pending implementation

---

### [DEC-004] - Use Chi Router as Default (Over Fiber, Stdlib)

**Date:** 2025-11-20

**Status:** Accepted

**Decision Makers:** Framework maintainer

**Context:**
Needed to choose a default HTTP router for the framework. Zorya supports multiple routers via adapters, but need to pick one as "recommended" for docs/examples.

**Options Considered:**

#### Option 1: Chi
**Pros:**
- Idiomatic Go (stdlib-like API)
- Lightweight and fast
- Excellent middleware ecosystem
- Context-based pattern matching
- Go 1.22+ `ServeMux` patterns now supported

**Cons:**
- Not fastest in benchmarks (Fiber wins)

#### Option 2: Fiber
**Pros:**
- Fastest router (Express.js-like API)
- Large ecosystem

**Cons:**
- Uses fasthttp (not `net/http` stdlib)
- Different idioms (not as Go-like)

#### Option 3: Stdlib ServeMux (Go 1.22+)
**Pros:**
- No dependencies
- Built-in to Go

**Cons:**
- Limited middleware support
- No route groups
- Less ergonomic than Chi/Fiber

**Decision:**
We chose **Option 1: Chi**

**Rationale:**
- Idiomatic Go: Feels like stdlib, easy for Go developers to adopt
- Middleware: Excellent ecosystem (httplog, cors, etc.)
- Zorya works with any router, but Chi is best balance of performance, idioms, and ecosystem
- Can still use Fiber/stdlib via adapters

**Implications:**
- Examples use Chi
- `fxhttpserver` module defaults to Chi
- Docs show Chi first, then mention adapters

**Success Criteria:**
- Developers find Chi easy to use
- Performance is acceptable

**Actual Outcome:**
✅ Chi is default. No complaints. Performance adequate for 99% of use cases.

---

### [DEC-005] - Struct Tags Over Configuration Files (DX Choice)

**Date:** 2025-10-15

**Status:** Accepted

**Decision Makers:** Framework maintainer

**Context:**
OpenAPI generation and request validation need metadata about fields. Could use struct tags or external config files (YAML/JSON).

**Problem Statement:**
Should metadata be in struct tags or external config files?

**Options Considered:**

#### Option 1: Struct Tags
**Pros:**
- Co-located with types (single source of truth)
- No file sync issues
- IDE autocomplete for tags
- Go's idiomatic approach

**Cons:**
- Can get verbose
- Limited expressiveness (string-based)

#### Option 2: External Config Files
**Pros:**
- More expressive (YAML/JSON)
- Easier to edit for non-coders

**Cons:**
- Files go out of sync with code
- Harder to discover (separate files)
- More moving parts

**Decision:**
We chose **Option 1: Struct Tags**

**Rationale:**
- Go idiom: Struct tags are standard (JSON, XML, GORM, validator)
- Single source of truth: Type and metadata in same place
- LLM-friendly: AI can see type and metadata together
- Discoverability: IDE shows tags in autocomplete

**Implications:**
- All metadata in tags: `schema`, `validate`, `openapi`, `body`, `status`
- Tag parser component handles parsing
- More verbose structs, but clearer intent

**Success Criteria:**
- Developers find tags intuitive
- LLMs generate correct tags

**Actual Outcome:**
✅ Tags work well. Users occasionally complain about verbosity, but appreciate co-location.

---

## Decision Categories

### Technical Architecture

| ID | Decision | Date | Status |
|----|----------|------|--------|
| DEC-001 | Merge Kernel into Application | 2026-01-13 | Implemented |
| DEC-002 | Panic on FX initialization failure | 2026-01-13 | Implemented |
| DEC-003 | Remove RegisterLifecycle | 2026-01-14 | Proposed |
| DEC-004 | Use Chi router as default | 2025-11-20 | Accepted |

### Product Strategy

| ID | Decision | Date | Status |
|----|----------|------|--------|
| DEC-005 | Struct tags over config files | 2025-10-15 | Accepted |

### Process & Workflow

| ID | Decision | Date | Status |
|----|----------|------|--------|
| TBD | - | - | - |

---

## Superseded Decisions

### [DEC-002-SUPERSEDED] - Store FX Errors, Return Later

**Originally decided:** 2025-12-10  
**Superseded by:** DEC-002 (Panic immediately) on 2026-01-13  
**Reason for change:** Error deferral made debugging harder; FX errors are configuration mistakes that should fail-fast  
**Learning:** Configuration/wiring errors belong in the "panic" category, not "return error" category

---

## Decision Review Schedule

| Decision ID | Next Review Date | Owner |
|-------------|------------------|-------|
| DEC-003 | 2026-02-01 | Framework maintainer |
| DEC-004 | 2026-06-01 | Framework maintainer |

---

## Related Documents

- [Implementation Log](implementation-log.md) - Code changes from decisions
- [Insights](insights.md) - Learnings from decisions
- [PRD](../01-product/prd.md) - Product requirements influenced by decisions
