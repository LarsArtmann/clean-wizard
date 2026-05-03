# Code Quality Scan — 2026-05-03

## Build Status
- `go build ./...` — ✅ PASS

## Test Status
- 2 test failures FIXED in this scan:
  - `TestSystemCacheCleaner_ValidateSettings` — hardcoded macOS cache types on Linux
  - `TestGoCacheCleaner_getGoBuildCacheLocations_Deduplication` — `os.TempDir()` == `/tmp` on Linux

## Lint Warnings Summary (from gopls/golangci-lint)

| Category | Count | Severity |
|----------|-------|----------|
| `err113` (dynamic errors) | ~40 | Warning |
| `gochecknoglobals` (global vars) | ~20 | Warning |
| `recvcheck` (mixed receivers) | ~10 | Warning |
| `funlen` (function too long) | ~5 | Warning |
| `infertypeargs` (unnecessary type args) | ~8 | Info |
| `unusedfunc` (unused functions) | ~3 | Info |
| `cyclop` (complexity > 10) | ~3 | Warning |
| `exhaustruct` (missing struct fields) | ~2 | Warning |
| `goconst` (repeated strings) | ~2 | Warning |
| Other | ~5 | Info/Warning |

## Files Over 350 Lines (Non-Test)

| File | Lines | Action |
|------|-------|--------|
| `internal/cleaner/compiledbinaries.go` | 585 | Split |
| `internal/cleaner/docker.go` | 524 | Split |
| `internal/cleaner/nodepackages.go` | 523 | Split |
| `cmd/clean-wizard/commands/githistory.go` | 516 | Split |
| `cmd/clean-wizard/commands/init.go` | 496 | Split |
| `internal/cleaner/systemcache.go` | 444 | Split |
| `internal/config/config.go` | 426 | Split |
| `internal/cleaner/githistory_executor.go` | 417 | Split |
| `internal/cleaner/githistory.go` | 416 | Split |
| `internal/cleaner/githistory_scanner.go` | 404 | Split |
| `internal/conversions/conversions.go` | 399 | Split |
| `internal/cleaner/projectexecutables.go` | 371 | Split |
| `cmd/clean-wizard/commands/cleaner_implementations.go` | 369 | Split |
| `internal/cleaner/fsutil.go` | 356 | Split |
| `internal/cleaner/golang_cache_cleaner.go` | 353 | At limit |

## Top Issues by Priority

### Critical (Fix Now)
1. **Test failures on Linux** — ✅ FIXED
2. **Unused function `readConfigFile`** in `config.go:52`
3. **Unused const `nixGenerationsMin`** in `sanitizer_nix.go:9`
4. **Unused const `bytesPerKB`** in `adapters/nix.go:28`

### High (Fix Soon)
5. **`err113` violations** — ~40 `fmt.Errorf` calls creating dynamic errors; should use `errors.New()` with sentinel errors or `errors.Join()`
6. **`exhaustruct`** — `domain.SizeEstimate` missing `Status` field in `golang_cache_cleaner.go:200`
7. **Mixed receivers** — `RiskLevelType`, `CleanStrategyType`, `SizeEstimateStatusType` use both pointer and value receivers

### Medium (Technical Debt)
8. **`funlen` violations** — `Clean`, `cleanGoCacheEnv`, `TestSystemCacheCleaner_ValidateSettings` too long
9. **`cyclop` violations** — `GetOperationType` (complexity 17), `cleanGoBuildCache` (11)
10. **`goconst`** — `"go-build*"` repeated 3 times in `golang_cache_cleaner.go`
11. **15 files over 350 lines** — need splitting

### Low (Polish)
12. **`infertypeargs`** — 8 places with unnecessary explicit type parameters
13. **`gochecknoglobals`** — enum string slices as globals (acceptable pattern for Go)
