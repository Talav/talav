# Insights

> **Learnings, patterns, and wisdom extracted from building Talav Framework**

---

## Architecture & Design

### Insight: YAGNI Beats Premature Abstraction

**Context:** Framework had `Kernel` abstraction separating FX lifecycle from `Application`

**What We Learned:**
- Abstraction added complexity without clear benefit
- No plans for alternative DI frameworks → abstraction unused
- Merging `Kernel` into `Application` reduced code by 100 LOC
- **Lesson**: Don't abstract until you have 2+ concrete implementations

**When to Apply:**
- Question every "might need this later" abstraction
- Wait for second use case before adding flexibility
- Simpler code is easier to maintain

---

### Insight: Fail-Fast on Configuration Errors

**Context:** Framework deferred FX initialization errors for later retrieval

**What We Learned:**
- Configuration/wiring errors are **programming mistakes**, not runtime errors
- Deferring errors made debugging harder (error discovered far from source)
- Panicking immediately forces fix before code reaches production
- Go precedent: `http.ListenAndServe` panics on invalid config

**When to Apply:**
- **Panic**: Configuration errors, DI wiring failures, invalid struct tags
- **Return error**: Runtime errors (network, I/O, external services)

**Rule of Thumb:**
```go
// ❌ BAD: Defer configuration errors
func New(cfg Config) (*Thing, error) {
    if cfg.Invalid() {
        return nil, errors.New("invalid config")
    }
    // Bug might not be caught until much later
}

// ✅ GOOD: Panic on configuration errors
func New(cfg Config) *Thing {
    if cfg.Invalid() {
        panic("invalid config: " + cfg.Validate())
    }
    // Error caught immediately
}
```

---

### Insight: Component Autonomy via Commands

**Context:** HTTP server lifecycle managed by commands, not FX hooks

**What We Learned:**
- Commands are natural control points (start server, run migrations, seed DB)
- FX should provide DI, not lifecycle management
- Components owning their lifecycle makes them:
  - **Reusable**: Work in different applications
  - **Testable**: No need for complex FX setup
  - **Clear**: Explicit start/stop, no magic

**When to Apply:**
- Long-running services (servers, workers) → commands
- Short tasks (migrations, seeds) → commands
- Avoid FX lifecycle hooks (`OnStart`, `OnStop`) in favor of commands

**Pattern:**
```go
// Component defines command (no framework deps)
func NewServeHTTPCmd(server *Server, logger *slog.Logger) *cobra.Command {
    return &cobra.Command{
        Use: "serve-http",
        RunE: func(cmd *cobra.Command, args []string) error {
            return server.Start(cmd.Context()) // Command owns lifecycle
        },
    }
}

// FX module registers command
var FxHTTPServerModule = fx.Module(
    "httpserver",
    fx.Provide(NewServer),
    fxcore.AsRootCommand(NewServeHTTPCmd),
)
```

---

## Type Safety & DX

### Insight: Struct Tags Are Single Source of Truth

**Context:** OpenAPI generation and validation need metadata

**What We Learned:**
- Co-locating metadata with types prevents sync issues
- Struct tags are Go-idiomatic (JSON, XML, GORM, validator)
- LLMs can see type and metadata together
- IDE shows tags in autocomplete

**Trade-offs:**
- **Pros**: Single source of truth, no external files to sync
- **Cons**: Verbose structs, limited expressiveness (strings)

**When to Apply:**
- Use tags for: validation rules, OpenAPI metadata, serialization hints
- Use external config for: runtime behavior, environment-specific settings

---

### Insight: Explicit Body Type Over Content-Type Inference

**Context:** Binary file uploads were decoded as JSON

**What We Learned:**
- `Content-Type` headers are unreliable (generic `application/octet-stream`)
- Explicit `body:"file"` tag should take precedence
- **Order of precedence**: Explicit tag > Content-Type > Default

**When to Apply:**
- Always check explicit metadata before heuristics
- Heuristics are fallbacks, not primary logic

---

## Testing & Quality

### Insight: Test Coverage Reveals Hidden Bugs

**Context:** Adding tests for file upload and multipart revealed 3 bugs

**What We Learned:**
- Bugs often hide in "it probably works" code paths
- Edge cases (binary data, empty fields, unknown fields) uncover issues
- High coverage (80%+) forces thinking about all code paths

**When to Apply:**
- Write tests for edge cases **first** (empty, nil, wrong type)
- Cover both happy path and error paths
- Use table-driven tests for multiple scenarios

**Pattern:**
```go
func TestDecoder_DecodeBody(t *testing.T) {
    tests := []struct {
        name string
        contentType string
        body []byte
        bodyType BodyType
        expected map[string]any
    }{
        {"JSON", "application/json", jsonBytes, BodyTypeStructured, ...},
        {"Empty File", "application/octet-stream", []byte{}, BodyTypeFile, emptyMap},
        {"Multipart", "multipart/form-data", multipartBytes, BodyTypeMultipart, ...},
        {"Unknown Fields Ignored", "application/x-www-form-urlencoded", formBytes, BodyTypeStructured, ...},
    }
    // ...
}
```

---

### Insight: Go Workspace Modules Need Care

**Context:** Go fetched old remote versions instead of local workspace modules

**What We Learned:**
- `go.work` doesn't guarantee local versions are used
- Stale `require` directives in `go.mod` can point to old remote versions
- Must run `go work sync` + `go mod tidy` after restructuring

**Prevention:**
- Automate: `scripts/tidy.sh` runs sync + tidy in all modules
- CI check: Ensure workspace versions match `go.work.sum`

---

## LLM-Assisted Development

### Insight: Clear Boundaries Make AI More Effective

**Context:** Layering rules prevent AI from creating circular dependencies

**What We Learned:**
- LLMs respect explicit rules in `.cursor/rules/`
- Clear import restrictions prevent common mistakes:
  - Components importing FX
  - FX modules importing framework
- Linter enforces what documentation describes

**When to Apply:**
- Document architecture constraints in machine-readable format
- Use linter to enforce (don't rely on code review alone)
- AI assistants read rules before generating code

**Pattern:**
```markdown
## Layer Rules (in .cursor/rules/)

Components (`pkg/component/*`):
- CAN: Import other components, stdlib, third-party
- CANNOT: Import pkg/fx/*, pkg/module/*

FX (`pkg/fx/*`):
- CAN: Import components only
- CANNOT: Import pkg/module/*
```

---

### Insight: Comprehensive READMEs Reduce Ambiguity

**Context:** Each component has detailed README with examples

**What We Learned:**
- LLMs generate better code when examples are clear
- README structure matters:
  1. What (2-3 sentences)
  2. Quick Start (copy-paste example)
  3. API Reference (with code snippets)
  4. Patterns (how to use correctly)
- More examples = less ambiguity

**When to Apply:**
- Every component/module needs README
- Show both simple and advanced use cases
- Include common mistakes section

---

## Patterns That Work

### Pattern: Builder Pattern for Configuration

✅ **Good:**
```go
app := framework.NewApplication(
    framework.WithName("myapp"),
    framework.WithVersion("1.0.0"),
    framework.WithModules(
        fxlogger.FxLoggerModule,
        fxhttpserver.FxHTTPServerModule,
    ),
)
```

**Why:** Readable, extensible, optional parameters clear

---

### Pattern: Panic for Invariant Violations

✅ **Good:**
```go
func (m *Metadata) GetField(name string) *FieldMetadata {
    field, ok := m.fields[name]
    if !ok {
        panic(fmt.Sprintf("field %s not found (programmer error)", name))
    }
    return field
}
```

**Why:** If field is missing, it's a bug in calling code, not runtime condition

---

### Pattern: Context Propagation for Request Metadata

✅ **Good:**
```go
// Middleware stores metadata in context
ctx = context.WithValue(ctx, securityKey, securityMeta)

// Handler retrieves from context
func GetSecurityMetadata(ctx context.Context) *SecurityMetadata {
    return ctx.Value(securityKey).(*SecurityMetadata)
}
```

**Why:** Decouples middleware from handlers, testable

---

## Patterns That Don't Work

### Anti-Pattern: Getters for Unexported Fields (Without External API)

❌ **Bad:**
```go
type Application struct {
    name string
}

func (a *Application) Name() string { return a.name }
```

**Why:** Unnecessary ceremony if only used by tests (same package can access fields)

---

### Anti-Pattern: Deferred Initialization

❌ **Bad:**
```go
type Application struct {
    initErr error
    once sync.Once
}

func (a *Application) init() {
    a.once.Do(func() {
        a.initErr = a.doInit()
    })
}

func (a *Application) Execute() error {
    a.init() // Deferred error discovery
    if a.initErr != nil {
        return a.initErr
    }
    // ...
}
```

**Why:** Errors discovered late, harder to debug

---

## Future Considerations

### Open Questions

1. **Should Talav support non-HTTP protocols (gRPC, WebSockets)?**
   - Current: HTTP/REST only
   - Trade-off: Focus vs breadth

2. **How to handle background jobs/queues?**
   - Current: Not supported
   - Need: Research temporal.io, go-workers patterns

3. **Should components be separate Go modules?**
   - Current: Monorepo with workspaces
   - Trade-off: Versioning independence vs management overhead

---

## Key Takeaways

1. ✅ **YAGNI over premature abstraction** - Wait for second use case
2. ✅ **Fail-fast on configuration errors** - Panic, don't defer
3. ✅ **Components own lifecycle via commands** - No magic FX hooks
4. ✅ **Struct tags as single source of truth** - Co-locate metadata with types
5. ✅ **Explicit metadata over heuristics** - Tags > Content-Type > Default
6. ✅ **High test coverage reveals bugs** - Edge cases matter
7. ✅ **Clear boundaries help AI** - Document + enforce architecture
8. ✅ **Examples reduce ambiguity** - Show, don't just tell

---

## Related Documents

- [Decision Log](decision-log.md) - Why we made choices
- [Implementation Log](implementation-log.md) - What changed
- [Bug Log](bug-log.md) - What broke and how we fixed it
