# go-business-rules Integration Assessment & Adoption

**Date:** 2026-09-23
**Library:** `github.com/LarsArtmann/go-business-rules/v2@v2.2.0` (public, severity-aware validation)
**Status:** Phase 1 ADOPTED — config validation engine rewritten on businessrules rules

---

## Executive Summary

`go-business-rules` adds severity levels (Info / Warning / Error / Critical) to
validation: instead of pass/fail, rules report the *degree* of failure. This
maps directly onto clean-wizard's config validation, which had a hand-rolled
mini-framework with the same shape but fewer guarantees.

**Adopted (this change):**

- All five config validation levels (structure, field, cross-field, business
  logic, security) are now expressed as declarative `businessrules.Rule`s in
  `internal/config/validation_rules.go` and executed through the library's
  validator. The imperative per-level append code
  (`validator_structure.go`, `validator_crossfield.go`, parts of
  `validator_business.go`) is deleted.
- 4-level severity: `SeverityCritical` added to
  `operations.ValidationSeverity` (previously 3 levels, no critical). A parent
  directory reference (`..`) in a protected path now escalates to
  **critical** instead of plain error.
- Validation warnings are no longer silently dropped at the load boundary:
  `validateLoadedConfig` logs every warning (field, message, suggestion) via
  `logger.Warn`. Previously warnings existed only inside `ValidationResult`
  and never reached CLI users.
- Rule metadata (`WithTags` = level + semantic kind, `WithDescription` = fix
  suggestion) replaces ad-hoc string fields; the bridge surfaces both in
  `ValidationError.Rule` / `.Suggestion` — JSON output shape is unchanged.
- The declared-but-unenforced `MinProtectedPaths` constraint is now enforced.
- Profile iteration is sorted → deterministic violation order (verified by
  `TestValidateConfig_DeterministicViolationOrder`).

**Behavioral compatibility:** field names, messages, suggestions, warning
buckets, and `IsValid` semantics are preserved; the existing config/BDD/CLI
test suites pass unchanged. One intentional escalation: `..` path traversal
is now `severity: "critical"`.

## Why the library fit

| Clean-wizard pain (before)                         | businessrules answer (after)                          |
| -------------------------------------------------- | ----------------------------------------------------- |
| Homegrown `ValidationRule[T]` (min/max/pattern/regex cache) — a less-tested re-implementation | Pre-built, fuzz/property-tested rule builders |
| 3-level severity strings, no critical              | 4-level `Severity`, `result.Errors()/Warnings()/Critical()` filtering |
| `ValidationError` vs `ValidationWarning` duplicated shapes appended by hand | One violation model; severity decides the bucket |
| `..` traversal = same severity as a bad percentage  | `SeverityCritical` — blocking failures ranked          |
| Warnings computed then discarded at `config.Load`  | Severity-aware result is complete; boundary logs them  |
| Unenforced declared constraints (`MinProtectedPaths`) | Every declared constraint becomes a rule           |
| Non-deterministic map-iteration violation order    | Sorted profile iteration + sequential rule execution   |

Both projects require `GOEXPERIMENT=jsonv2`, so the library's json/v2
marshaling is a non-event here.

## Deliberately NOT adopted (yet)

| Capability                          | Reason                                                                                  |
| ----------------------------------- | --------------------------------------------------------------------------------------- |
| `Stream(ctx)` concurrent evaluation | Config rules are cheap (µs); sequential gives deterministic order for free              |
| `listeners/otel` / cqrslite bridges | No OpenTelemetry or event bus in clean-wizard today                                     |
| Composite builders (`When`/`All`/`Or`) | Condition checks inside closures read clearer at current rule count                 |
| `operations`-level `ValidateSettings` rewrite | `internal/domain` is stdlib-only by design (`docs/PACKAGE_BOUNDARY.md`); library belongs in `internal/config` |
| `ValidateField` severity retention  | Returns `error` by contract; only used by tests/middleware today                        |

## Future opportunities

- Rewrite `ValidationMiddleware.validateChangeBusinessRules` ifs as tagged
  rules (same engine, change-level severities).
- Emit `RuleEvaluated` events into verbose CLI output (`-v`) for free
  per-rule timing.
- If `go-finding` ever unifies with the error-family taxonomy, map
  validation severity → `errorfamily.Family` centrally.

## Verification

- `go build ./...` + `go test ./... -short` green (config, cmd, execution, di,
  domain, bdd suites).
- New tests: `internal/config/validation_rules_test.go` (severity bucketing,
  critical escalation, structured metadata preservation, warning bridging,
  deterministic order, safe-mode context metadata).
- `buildflow -s golangci-lint`: zero findings on changed files.
- `nix fmt`: clean.
