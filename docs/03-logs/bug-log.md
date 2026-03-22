# Bug Log

> **Track bugs, fixes, root causes, and prevention strategies**

---

## Active Bugs

_No active bugs currently tracked_

---

## Resolved Bugs

### [BUG-001] - Schema Decoder Ignored `body:"file"` and `body:"multipart"` Tags

**Reported:** 2026-01-13  
**Fixed:** 2026-01-13  
**Severity:** High (tests failing)

**Symptoms:**
- Tests `TestDecoder_DecodeBody_File` and `TestDecoder_DecodeBody_Multipart` failing
- Binary file content (images, PDFs) decoded as JSON instead of raw bytes
- Multipart form data parsed incorrectly

**Root Cause:**
1. Test helper `createBodyMetadata()` was ignoring `bodyType` parameter, never setting `body` tag
2. `decodeBody()` logic prioritized `Content-Type` header over explicit `bodyMeta.BodyType`
3. Generic `Content-Type: application/octet-stream` triggered JSON decoder fallback

**Fix:**
1. Updated `createBodyMetadata()` to set `body` tag based on `BodyType` parameter
2. Changed `decodeBody()` to check `bodyMeta.BodyType` **before** checking `Content-Type`
3. Added explicit cases for `BodyTypeFile` and `BodyTypeMultipart` at top of decode logic

**Prevention:**
- Add integration test with actual binary file upload
- Document body type precedence: explicit tag > Content-Type > default

**Commit:** `df03db3`

**Files:**
- `pkg/component/schema/decoder_body_test.go`
- `pkg/component/schema/decoder_body.go`

---

### [BUG-002] - URL-Encoded Form Decoder Included Unknown Fields

**Reported:** 2026-01-13  
**Fixed:** 2026-01-13  
**Severity:** Medium (unexpected behavior)

**Symptoms:**
- `TestDecoder_DecodeBody_URLEncodedForm` failing
- Form data included fields not defined in struct
- Expected behavior: Ignore unknown fields (OpenAPI-style strict parsing)

**Root Cause:**
- `decodeURLEncodedForm()` parsed all form fields into map
- No filtering against struct metadata
- Result included `extra_field` not in struct definition

**Fix:**
- Added filtering loop in `decodeURLEncodedForm()`
- Only include fields present in `schemaMeta.Fields`
- Use `ParamName` (tag name) for matching form keys

**Prevention:**
- Add test case explicitly checking unknown fields are ignored
- Document "strict mode" behavior in README

**Commit:** `df03db3`

**Files:**
- `pkg/component/schema/decoder_body.go`

---

### [BUG-003] - Multipart Decoder Used Field Name Instead of Tag Name

**Reported:** 2026-01-13  
**Fixed:** 2026-01-13  
**Severity:** High (data loss)

**Symptoms:**
- `TestDecoder_DecodeBody_Multipart` failing
- Multipart form values not accessible in decoded map
- Map keys didn't match expected names

**Root Cause:**
- `processMultipartField()` used `mapKey` (Go field name) as map key
- Should have used `paramName` (struct tag name)
- Form data uses tag names, not field names

**Fix:**
- Changed `result[mapKey]` to `result[paramName]` in `processMultipartField()`
- Now consistent with other decoders (query, header, cookie)

**Prevention:**
- Add test verifying tag name vs field name mapping
- Document that struct tags define wire format, not field names

**Commit:** `df03db3`

**Files:**
- `pkg/component/schema/decoder_body.go`

---

### [BUG-004] - FX Module Resolution Fetched Old Remote Versions

**Reported:** 2026-01-12  
**Fixed:** 2026-01-12  
**Severity:** Critical (build failure)

**Symptoms:**
- `go mod tidy` failed with "package X does not exist in module Y"
- `undefined: TypeName` errors for types that exist in local workspace
- Go was fetching remote versions instead of using workspace modules

**Root Cause:**
- `go.mod` files had stale `require` directives pointing to old pseudo-versions
- Go module resolution preferred remote versions over workspace
- Remote versions had different package structure than local

**Fix:**
- Updated all `require` directives to latest workspace pseudo-versions
- Ran `go work sync` to update `go.work.sum`
- Ran `go mod tidy` in all modules to refresh dependencies

**Prevention:**
- Run `scripts/tidy.sh` after any module restructuring
- Add CI check to ensure workspace versions are used
- Document workspace module workflow

**Commit:** `9da34ad`

**Files:**
- Multiple `go.mod` files across `pkg/fx/*`, `pkg/module/*`

---

### [BUG-005] - `reflect.Ptr` Deprecation Warning

**Reported:** 2026-01-13  
**Fixed:** 2026-01-13  
**Severity:** Low (linter warning)

**Symptoms:**
- Linter warning: "Constant reflect.Ptr should be inlined"
- Using deprecated `reflect.Ptr` constant

**Root Cause:**
- `reflect.Ptr` renamed to `reflect.Pointer` in Go 1.18
- Old code still used deprecated name

**Fix:**
- Replaced `reflect.Ptr` with `reflect.Pointer`

**Prevention:**
- Keep linter updated
- Run `golangci-lint` in CI

**Commit:** `0155cc2`

**Files:**
- `pkg/component/schema/decoder_body.go`

---

## Bug Analysis

### By Category

| Category | Count | Notes |
|----------|-------|-------|
| Schema Decoding | 3 | Body type detection, field mapping |
| Module Resolution | 1 | Go workspace versions |
| Linter Warnings | 1 | Deprecated API usage |

### By Severity

| Severity | Count |
|----------|-------|
| Critical | 1 |
| High | 2 |
| Medium | 1 |
| Low | 1 |

### Common Patterns

1. **Test Coverage Gaps**: Bugs discovered when tests were added (file upload, multipart)
   - **Lesson**: Write tests for edge cases (binary data, unknown fields)

2. **Tag Name vs Field Name Confusion**: Multiple bugs related to using field name instead of tag name
   - **Lesson**: Be explicit about wire format (tags) vs struct definition (fields)

3. **Go Module Workspace Complexity**: Go's module resolution can be confusing with workspaces
   - **Lesson**: Document workspace patterns, automate with scripts

---

## Prevention Strategies

### Already Implemented

✅ **Linter in CI**: Run `golangci-lint` on every commit  
✅ **Test Coverage Target**: 80%+ for components  
✅ **Scripts for Common Tasks**: `scripts/tidy.sh`, `scripts/lint.sh`

### Planned

- [ ] **Integration Tests**: Add tests with real HTTP requests (not just unit tests)
- [ ] **Fuzzing for Schema Decoder**: Test with random/malformed inputs
- [ ] **Performance Benchmarks**: Catch regressions in request decoding
- [ ] **Dependency Scanning**: Use `govulncheck` in CI

---

## Related Documents

- [Implementation Log](implementation-log.md) - Code changes that fixed bugs
- [Insights](insights.md) - Learnings from bug patterns
