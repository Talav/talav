# Vision

## Why

Go has no good middle ground for HTTP API development. Raw `net/http` + Chi gives you flexibility but zero structure. Full-stack frameworks give you structure but couple your business logic to the framework. You end up either writing the same wiring code on every project, or locked into abstractions you can't escape.

Talav is an attempt to fix that: a small, opinionated framework that enforces a layered architecture and provides type-safe HTTP handling — without owning your domain logic.

> For what is currently built, see [system-map.md](system-map.md).

## Design Principles

**1. Hard layering**

Three layers, enforced by linting:

- `pkg/component/*` — business logic, zero framework/FX imports
- `pkg/fx/*` — dependency wiring only, no business logic
- `pkg/module/*` / `main.go` — domain modules and application assembly (`framework/` handles bootstrap; `user/`, `media/`, `security/` compose components into deployable slices)

Violations fail the pre-commit hook. There are no exceptions.

**2. Type-safe HTTP I/O**

Request decoding and response encoding use generics. No `interface{}`, no `map[string]any`, no manual JSON parsing in handler code. The compiler catches shape mismatches; the runtime catches value errors (validation tags).

**3. Standards over convenience**

Errors follow RFC 9457. OpenAPI output targets 3.1. Content negotiation follows RFC 9110. When there's a conflict between "easier to implement" and "correct per spec", the spec wins.

**4. Components own their lifecycle**

A component defines its own Cobra command and manages its own start/stop via `cmd.Context()`. FX wires dependencies; it does not own the lifecycle. This keeps components portable — you can use them outside of FX if needed.

## Constraints

- **Go 1.25+** — uses range-over-func and newer stdlib patterns
- **Uber FX** — not pluggable; the DI model is load-bearing
- **Chi v5 or Fiber v2** — other routers via adapter, but these are the tested paths
- **MIT license**

## Stage

Alpha — core patterns are stable, public API is not. Breaking changes are possible before 1.0.
