# Assumptions & Open Questions

## Validated

| Date | Assumption | Result |
|------|------------|--------|
| 2026-01-10 | Users want type-safe handlers over `map[string]any` | 9/10 devs surveyed preferred typed — proceeding with generic `Register[I, O]` API |
| 2026-01-12 | Go 1.25+ is a reasonable requirement | Go 1.24→1.25 upgrade cycle ~5 months — acceptable |

## Open Assumptions

**FX startup performance is acceptable**
DI container init must stay under ~100ms for CLI use. Not yet benchmarked with a realistic module count (30+).
- *Validate*: benchmark `fx.New()` with full module set before beta

**Reflection overhead in schema decoding is negligible**
Metadata caching is in place but not measured. Target: <1ms per request overhead from Zorya.
- *Validate*: add benchmark to CI before beta

**Monorepo with Go workspaces scales to 1.0**
Currently working well. Open question is whether components should be versioned independently or as a single release.
- *Decide by*: Q4 2026 (before 1.0)

## Open Questions

**Database migrations** — no opinionated answer yet. Wrong choice causes production downtime.
Options: auto-migrate on startup, separate CLI command, or external tool (goose/atlas).
- *Decide by*: Q2 2026 (beta)

**API versioning** — currently delegated to the user. Two viable approaches: URL prefix (`/v1/`) or content negotiation (`Accept: application/vnd.api.v1+json`). Picking one as a recommended default would reduce decision fatigue.
- *Decide by*: Q2 2026 (beta)

**Security enforcement guidance** — the middleware is pluggable but there are no canonical examples for common patterns (simple roles, Casbin RBAC, JWT). Without examples, users will implement it inconsistently.
- *Decide by*: Q2 2026 (beta)
