# Implementation Log

> **Chronological record of significant code changes and why they were made**

---

## 2026-01-14

### Framework: Removed RegisterLifecycle Dead Code

**What Changed:**
- Identified `RegisterLifecycle()` function in `pkg/fx/fxhttpserver/module.go` (lines 120-138) as unused
- Function contradicts module's documented "command-based lifecycle" approach

**Why:**
- Never called anywhere in codebase
- Violates component autonomy principle (components own lifecycle via commands)
- Causes confusion (docs say "commands control lifecycle", but code suggests FX hooks)

**Status:** Identified, pending removal

**Files:**
- `pkg/fx/fxhttpserver/module.go` (proposed deletion of lines 120-138)

---

## 2026-01-13

### Framework: Merged Kernel into Application

**What Changed:**
- Deleted `pkg/module/framework/kernel.go`
- Moved `modules`, `fxApp`, `commands` fields from `Kernel` to `Application`
- Moved `initFX()` method from `Kernel.Init()` to `Application.initFX()`
- Removed `initOnce` sync.Once (not needed, FX inits once during construction)

**Why:**
- `Kernel` was thin wrapper adding unnecessary indirection
- No plans to support alternative DI frameworks (YAGNI principle)
- Simpler code easier to maintain and understand

**Impact:**
- Reduced codebase by ~100 LOC
- No user-facing API changes
- Tests updated to access fields directly instead of via getters

**Commit:** `77ff54f`

**Files:**
- Deleted: `pkg/module/framework/kernel.go`
- Modified: `pkg/module/framework/application.go`, `pkg/module/framework/application_test.go`, `pkg/module/framework/README.md`

---

### Framework: Panic on FX Initialization Failure (Fail-Fast)

**What Changed:**
- `NewApplication()` now panics immediately if `initFX()` returns error
- Removed `initErr` field from `Application` struct
- Error is clear: `panic(fmt.Errorf("FX initialization failed: %w", err))`

**Why:**
- FX errors are programming mistakes (missing deps, circular deps, invalid config)
- Deferring errors made debugging harder
- Should never reach production if dev/test environments work
- Aligns with "fail-fast" for configuration errors

**Impact:**
- Errors discovered immediately at startup
- No need to check `initErr` throughout code
- Clearer error messages

**Commit:** `77ff54f`

**Files:**
- Modified: `pkg/module/framework/application.go`

---

### Framework: Removed Test Mode

**What Changed:**
- Removed `testMode` field from `Application` struct
- Removed `WithTestMode()` option
- Removed `TestApplication_TestMode` test

**Why:**
- `testMode` was unused (no code path checked it)
- Framework tests don't need special mode (use standard Go test patterns)
- Simplified API

**Commit:** `77ff54f`

**Files:**
- Modified: `pkg/module/framework/application.go`, `pkg/module/framework/options.go`, `pkg/module/framework/application_test.go`

---

### Framework: Removed Getter Methods

**What Changed:**
- Removed `FxApp()`, `Name()`, `Version()`, `Environment()`, `RootCmd()` methods
- Tests access fields directly (`app.name`, `app.version`, etc.)

**Why:**
- Getters only used by tests
- Go idiom: Unexported fields don't need getters unless external API needs them
- Tests can access unexported fields (same package)

**Commit:** `77ff54f`

**Files:**
- Modified: `pkg/module/framework/application.go`, `pkg/module/framework/application_test.go`

---

## 2026-01-13

### Schema: Fixed Tests for URL-Encoded and Multipart Forms

**What Changed:**
- `createBodyMetadata` now correctly sets `body` tag based on `BodyType` parameter
- `decodeURLEncodedForm` filters result to only include fields defined in metadata (ignores extra form fields)
- `processMultipartField` uses `paramName` (tag name) instead of `mapKey` (field name) for storing values

**Why:**
- Tests expect `body:"file"` and `body:"multipart"` tags, but helper was ignoring `BodyType` parameter
- URL-encoded form decoder was including all form fields, even ones not in struct
- Multipart decoder was using wrong keys, causing map lookups to fail

**Impact:**
- All schema tests now pass
- Form decoding matches OpenAPI-style behavior (ignore unknown fields)

**Commit:** `df03db3`

**Files:**
- Modified: `pkg/component/schema/decoder_body_test.go`, `pkg/component/schema/decoder_body.go`

---

### Schema: Fixed File Body Detection

**What Changed:**
- `decodeBody` now prioritizes explicit `bodyMeta.BodyType` checks (`BodyTypeFile`, `BodyTypeMultipart`) before relying on `Content-Type` header
- Empty file bodies return empty map instead of JSON decode error

**Why:**
- Binary files (images, PDFs) have generic `Content-Type: application/octet-stream`
- Logic was falling back to JSON decoder for binary content
- Tests had `body:"file"` tag but code wasn't respecting it

**Commit:** `df03db3`

**Files:**
- Modified: `pkg/component/schema/decoder_body.go`

---

### Linter: Fixed reflect.Ptr Deprecation Warning

**What Changed:**
- Replaced `reflect.Ptr` with `reflect.Pointer` constant

**Why:**
- `reflect.Ptr` deprecated in Go 1.18+
- Linter warning: "Constant reflect.Ptr should be inlined"

**Commit:** `0155cc2`

**Files:**
- Modified: `pkg/component/schema/decoder_body.go`

---

## 2026-01-12

### FX Modules: Fixed Module Version Resolution

**What Changed:**
- Updated `require` directives in `go.mod` files to use latest workspace pseudo-versions
- Example: `github.com/talav/talav/pkg/component/user v0.0.0-20260113034123-9da34ad44376`
- Ran `go work sync` and `go mod tidy` in all modules

**Why:**
- Go was fetching older remote versions instead of using local workspace modules
- Caused `undefined` errors for types/functions that exist in workspace but not in remote version
- Example: `pkg/component/user/repository` package didn't exist in old remote version

**Impact:**
- All modules now use local workspace versions
- Build errors resolved
- No need for `replace` directives

**Commit:** `9da34ad`

**Files:**
- Modified: `pkg/fx/fxuser/go.mod`, `pkg/module/media/go.mod`, `pkg/component/schema/go.mod`, and others

---

## 2025-12-20

### FxHTTPServer: Added HTTP Server Command

**What Changed:**
- Created `pkg/component/httpserver/cmd/serve.go` with `NewServeHTTPCmd()`
- Registered command in `fxhttpserver.FxHTTPServerModule` via `fxcore.AsRootCommand`
- Command calls `server.Start(cmd.Context())` to start server

**Why:**
- Component autonomy: HTTP server owns its lifecycle via command
- No FX lifecycle hooks needed
- Signal handling via framework-provided context

**Commit:** `bd850c3`

**Files:**
- Added: `pkg/component/httpserver/cmd/serve.go`
- Modified: `pkg/fx/fxhttpserver/module.go`

---

## 2025-12-15

### Framework: Added Application Module

**What Changed:**
- Created `pkg/module/framework/` with `Application`, `Kernel` structs
- Bootstrap helpers: `RunDefault()`, `Run()`, `RunWithContext()`
- Cobra CLI integration with command injection from FX modules
- Signal handling for graceful shutdown

**Why:**
- Needed entry point for applications built with Talav
- Cobra provides CLI structure
- FX provides DI
- Framework ties them together

**Commit:** `77ff54f`

**Files:**
- Added: `pkg/module/framework/application.go`, `pkg/module/framework/kernel.go`, `pkg/module/framework/bootstrap.go`, `pkg/module/framework/options.go`

---

## 2025-11-30

### Zorya: Added Security Declaration Support

**What Changed:**
- Added `Secure()` wrapper for declaring route security requirements
- Security options: `Auth()`, `Roles()`, `Permissions()`, `Resource()`, `ResourceFromParams()`, `Action()`
- Security metadata stored in context via auto-injected middleware
- Enforcement delegated to `security.SecurityMiddleware`

**Why:**
- Clean separation: Zorya declares requirements, security component enforces
- No circular dependencies (Zorya doesn't import security)
- Declarative API: `zorya.Secure(zorya.Roles("admin"))`

**Commit:** `8abaf62`

**Files:**
- Modified: `pkg/component/zorya/route.go`, `pkg/component/zorya/security.go`, `pkg/component/zorya/group.go`

---

## 2025-11-15

### Zorya: Initial Alpha Release

**What Changed:**
- Type-safe HTTP handler API: `zorya.Register[I, O any](api, route, handler)`
- Request parameter decoding via `schema.Codec`
- Response encoding with content negotiation
- OpenAPI 3.1 schema generation
- RFC 9457 error handling
- Chi/Fiber/stdlib adapters

**Why:**
- Core value prop: Type-safe APIs with minimal boilerplate
- OpenAPI auto-generation eliminates manual spec maintenance
- RFC compliance ensures production-ready APIs

**Commit:** `f48b304`

**Files:**
- Added: `pkg/component/zorya/*` (20+ files)

---

## Earlier Changes

See git log for detailed history: `git log --oneline --all`

---

## Notes

- This log captures **significant** changes that impact architecture, API, or behavior
- Routine bug fixes, refactoring, and minor tweaks are not logged here (see git history)
- Each entry should answer: What changed? Why? What was the impact?
