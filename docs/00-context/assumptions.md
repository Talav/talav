# Assumptions, Risks & Unknowns

## Assumptions

### User Assumptions

- [x] **Users prefer type safety over flexibility**: Developers want compile-time guarantees and IDE autocomplete, even if it requires more verbose struct definitions
  - *Validation*: Survey 10+ Go developers, 90% preferred typed handlers over `map[string]any`
  
- [ ] **Users will read documentation**: README files and examples are sufficient for onboarding
  - *Risk*: May need interactive tutorials or video walkthroughs
  
- [ ] **Users understand layered architecture**: Developers know the difference between components, FX modules, and application layer
  - *Risk*: May need enforced linting rules and better error messages

- [ ] **Users want OpenAPI generation**: Auto-generated OpenAPI specs are valuable enough to justify struct tag overhead
  - *Validation*: Track adoption of `/openapi.json` endpoint in example projects

### Technical Assumptions

- [ ] **Uber FX performance is acceptable**: DI container initialization time (<100ms) is fast enough for CLI applications
  - *Risk*: May need lazy initialization for large applications
  - *Validation*: Benchmark FX startup time with 50+ modules

- [x] **Go 1.25+ adoption is reasonable**: Requiring Go 1.25+ won't block adoption
  - *Validation*: Go 1.25 released Dec 2024, most teams upgrade within 6 months

- [ ] **Reflection overhead is negligible**: Using reflection for schema metadata doesn't significantly impact performance
  - *Validation*: Benchmark request handling with/without reflection (target: <1ms overhead)
  - *Current*: Metadata caching mitigates reflection cost, but needs verification

- [ ] **Chi router is sufficient for most use cases**: No need to support additional routers beyond Chi/Fiber/stdlib
  - *Risk*: Users on other routers (Gin, Echo) may fork or request adapters

### Business Assumptions

- [ ] **Open source model attracts contributors**: MIT license and GitHub hosting will build a community
  - *Risk*: Solo maintainer burnout without contributors
  - *Mitigation*: Set clear contribution guidelines, use GitHub Discussions

- [ ] **Developer-focused framework has market**: There's demand for a framework between "raw HTTP" and "full-stack Rails-like"
  - *Validation*: Track GitHub stars, issues, and adoption metrics

- [ ] **Documentation-driven development works with LLMs**: Comprehensive docs make AI assistants more effective
  - *Validation*: Test with Cursor/Copilot, measure code quality and consistency

## Risks

### High Priority Risks

| Risk | Impact | Probability | Mitigation Strategy |
|------|--------|-------------|---------------------|
| **Breaking API changes needed** | High - Early adopters abandon framework | High | - Establish 1.0 API stability guarantee<br>- Use semantic versioning strictly<br>- Provide migration guides |
| **Performance regression** | High - Users switch to faster alternatives | Medium | - Add performance benchmarks to CI<br>- Profile hot paths<br>- Set performance SLOs (e.g., <1ms overhead) |
| **Security vulnerability in dependencies** | High - Production applications compromised | Medium | - Use `govulncheck` in CI<br>- Pin dependency versions<br>- Monitor security advisories |
| **Solo maintainer burnout** | High - Framework becomes unmaintained | High | - Focus on quality over features<br>- Accept contributions early<br>- Document everything for future maintainers |

### Medium Priority Risks

| Risk | Impact | Probability | Mitigation Strategy |
|------|--------|-------------|---------------------|
| **FX DI adds too much complexity** | Medium - Users bypass FX, lose benefits | Medium | - Provide non-FX examples<br>- Document FX patterns clearly<br>- Consider FX-optional mode |
| **Layering rules too strict** | Medium - Users find workarounds | Medium | - Make linter configurable<br>- Provide escape hatches for advanced users<br>- Gather feedback early |
| **OpenAPI generation incomplete** | Medium - Users write specs manually | Low | - Prioritize common use cases<br>- Allow manual spec overrides<br>- Document limitations upfront |
| **Test coverage insufficient** | Medium - Bugs in production | Medium | - Set 80% coverage target<br>- Add integration tests<br>- Use fuzzing for critical paths |

### Low Priority Risks

| Risk | Impact | Probability | Mitigation Strategy |
|------|--------|-------------|---------------------|
| **Go 1.26+ breaks compatibility** | Low - Framework requires updates | Low | - Monitor Go release notes<br>- Test with release candidates<br>- Pin to Go minor version |
| **Uber FX development stalls** | Low - Need alternative DI solution | Low | - Monitor FX repo activity<br>- Document DI patterns<br>- Consider wire (Google) as fallback |

## Unknowns

### Critical Unknowns

- [ ] **Question**: How do users want to handle database migrations in production?
  - **Why it matters**: Wrong approach causes data loss or downtime
  - **How to resolve**: Survey users, implement 2-3 migration strategies (auto, manual, CLI command)
  - **Decision by**: Beta release (Q2 2026)

- [ ] **Question**: Should security enforcement be pluggable or opinionated?
  - **Why it matters**: Affects API design and flexibility
  - **Current state**: Middleware-based (pluggable), but may need more guidance
  - **How to resolve**: Build 3 real-world examples (simple roles, Casbin RBAC, custom logic)
  - **Decision by**: Beta release (Q2 2026)

- [ ] **Question**: How to handle long-running background jobs (queues, cron)?
  - **Why it matters**: Many APIs need async processing, current focus is HTTP-only
  - **How to resolve**: Research patterns (temporal.io, go-workers, embedded scheduler)
  - **Decision by**: Post-1.0 (2027)

### Important Unknowns

- [ ] **Question**: Should Zorya support GraphQL or stay REST-only?
  - **Why it matters**: GraphQL adoption growing, but adds complexity
  - **How to resolve**: Prototype GraphQL adapter, measure effort vs benefit
  - **Decision by**: Post-1.0 (2027)

- [ ] **Question**: How to version APIs (`/v1/users` vs `Accept: application/vnd.api.v1+json`)?
  - **Why it matters**: Versioning strategy affects API design
  - **Current state**: No opinionated approach, users handle manually
  - **How to resolve**: Document 2-3 versioning patterns, pick one as default
  - **Decision by**: Beta release (Q2 2026)

- [ ] **Question**: Should components be published as separate Go modules or monorepo?
  - **Why it matters**: Affects dependency management and versioning
  - **Current state**: Monorepo with Go workspaces
  - **How to resolve**: Research Go module best practices, gather community feedback
  - **Decision by**: Before 1.0 (Q4 2026)

- [ ] **Question**: How to handle file uploads >1GB (streaming, chunked, resumable)?
  - **Why it matters**: Current multipart handling loads entire file into memory
  - **How to resolve**: Implement streaming upload API, test with large files
  - **Decision by**: Post-1.0 (2027)

## Validation Log

| Date | Assumption | Method | Result | Action Taken |
|------|------------|--------|--------|--------------|
| 2026-01-10 | Users want type-safe handlers | Informal survey (10 devs) | 9/10 preferred typed over `map[string]any` | ✅ Proceed with generic `Register[I, O]` API |
| 2026-01-12 | Go 1.25+ adoption acceptable | Check Go release history | Go 1.24→1.25 took 5 months avg | ✅ Require Go 1.25+ |
| TBD | Reflection overhead negligible | Benchmark suite | Pending | Run benchmarks before beta |
| TBD | Documentation sufficient for onboarding | User testing (5 developers) | Pending | Test with developers unfamiliar with framework |
| TBD | OpenAPI generation meets needs | Example project analysis | Pending | Build 3 real-world APIs, check spec completeness |

---

**Note**: Assumptions should be validated continuously. Unknowns should be resolved by their target decision date. Risks should be reviewed monthly.
