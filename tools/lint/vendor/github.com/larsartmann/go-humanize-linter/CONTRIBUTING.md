# Contributing

Thanks for your interest in contributing to go-humanize-linter!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

### Prerequisites

- **Go 1.27+**
- **Nix** (recommended) — all build commands set the required env vars automatically

### Required Environment Variables

All `go` commands require:

```bash
export GOEXPERIMENT=jsonv2
```

(Nix sets it automatically. All dependencies are public — no `GOPRIVATE` setup needed.)

### Build Commands (via Nix)

```bash
nix run .#test          # go test ./... -count=1
nix run .#test-race     # go test ./... -race -count=1
nix run .#build         # go build ./...
nix run .#lint          # golangci-lint run ./... (standard linters)
nix run .#custom-lint   # build custom-gcl and run with gohumanize plugin
nix run .#vet           # go vet ./...
nix run .#coverage      # go test with coverage report
```

### Build Commands (direct Go)

```bash
export GOEXPERIMENT=jsonv2
go build ./...
go test ./... -race -count=1
go vet ./...
```

## golangci-lint v2 Module Plugin Workflow

The linter ships as a golangci-lint v2 **module plugin**. To test the full
integration (including golangci-lint runtime discovery):

1. `golangci-lint custom` reads `.custom-gcl.yml` and builds `./custom-gcl`
2. The custom binary discovers `gohumanize` via `register.Plugin("gohumanize", newPlugin)` in `init()`
3. Run with the custom config: `./custom-gcl run -c .golangci.custom.yml ./...`

**One-liner via Nix:**

```bash
nix run .#custom-lint
```

**Key requirement:** The `.golangci.yml` must include a `linters.settings.custom.gohumanize.type: "module"` section or golangci-lint reports "unknown linters". See `.golangci.custom.yml` for the complete config.

**Note:** `nix run .#lint` uses stock golangci-lint (without the plugin) and filters the "Found unknown linters in //nolint directives" warning. This is expected — the project's own `//nolint:gohumanize` directives reference the module plugin name, which stock golangci-lint doesn't know.

## Testing Strategy

- **Unit tests** (`linter_test.go`) — run each rule against `testdata/` fixtures
- **White-box tests** (`pattern_helpers_test.go`, `pattern_commaf_test.go`) — test AST helpers directly
- **Analysistest** (`plugin/plugin_test.go`) — end-to-end plugin path via `analysistest.Run`
- **Anti-regression** (`TestRuleCountConsistency`) — guards against ghost rules (registered but never fire)
- **Self-scan** (`TestLintsItself_Clean`) — the linter must produce 0 findings on its own source

### Testdata Layout

```
testdata/
  h008_ordinal/main.go        # positive fixture (should trigger H008)
  h008_negative/main.go       # negative fixture (must NOT trigger H008)
  analysistest/
    h008positive/main.go      # analysistest fixture with // want "H008"
    clean/main.go             # must produce no diagnostics
```

## Adding a New Rule (H0NN)

1. **`pattern_h0NN.go`** — AST pattern helpers (`hasPattern`, etc.). Each helper takes `*ast.FuncDecl` and returns `bool`.

2. **`rule_h0NN.go`** — Rule factory (`RuleName()`) returning `linter.RuleFunc`, and detector function (`detectName`). The detector calls the pattern helpers and returns `[]finding.Finding`.

3. **`rules.go`** — Register the rule in **two** places:
   - `AllRules()` — add `RuleName()` to the slice
   - `allRuleDetectors()` — add `{"H0NN", detectName}` to the slice

   `TestRuleCountConsistency` will fail if you forget either one.

4. **`testdata/h0NN_name/main.go`** — Positive fixture that triggers the rule.

5. **`testdata/h0NN_negative/main.go`** — Negative fixture with similar-looking code that must NOT trigger.

6. **`testdata/analysistest/h0NNpositive/main.go`** — Analysistest fixture with `// want "H0NN"` on the function line.

7. **`plugin/plugin_test.go`** — Add `"./h0NNpositive"` to the `analysistest.Run` call in `TestAnalyzerAnalysistest`.

8. **`linter_test.go`** — Add `TestRuleName_Positive` and `TestRuleName_Negative`.

9. **`cmd/go-humanize-linter/main.go`** — Add to `ruleExplanations` map using the exported `humanizelint.RuleIDH0NN` constant.

10. **`docs/rules/H0NN.md`** — Rule documentation (follow `H001.md` format).

11. **`README.md`** — Add to the rules table.

## Adding a CLI Flag

1. **`cmd/go-humanize-linter/main.go`** — Add the flag definition, wire it to the processing logic, and update `flag.Usage`.
2. **`cmd/go-humanize-linter/main_test.go`** — Add test coverage.
3. **`README.md`** — Document the flag in the usage section.

## Suppression Behavior

### How `//nolint` works

`//nolint:gohumanize` directives suppress findings. The directive can be
placed:

1. On the function declaration line (trailing comment)
2. In the doc comment (immediately before the function)
3. On the line before the declaration
4. **Inside the function body** (anywhere between `{` and `}`)

This last placement is important for AI-generated code and code review: the
directive can be placed on the exact statement that triggers the finding.

### Scoped suppression

`//nolint:gohumanize:H001` suppresses only H001, not all rules. This works
identically in both the CLI and the plugin path. Internally, `checkFuncDecls`
passes the rule ID to `isSuppressedRule(funcSuppressions(...), ruleID)`.

### Three suppression paths

There are three code paths that must stay consistent:

1. **CLI** (`walker.go:checkFuncDecls`) — uses `funcSuppressions` + `isSuppressedRule`
2. **Plugin** (`rules.go:HumanizeDetector.Run`) — same functions
3. **Verification** (`suppression.go:extractFunctionSuppressions`) — same `commentAssociatedWithFunc` helper

All three call `commentAssociatedWithFunc` for the association logic. If you
change the association rules, update all three paths.

### Namespace

The suppression namespace is `gohumanize` (the analyzer name), NOT
`go-humanize-linter` (the module path). `//nolint:go-humanize-linter` silently
does nothing. The `--verify-suppressions` flag catches this.

See ADR 0006 for the design decision on in-body (line-range) suppression.

Please use GitHub Issues to report bugs or request features.

## Behavior Delta Workflow

`--save-baseline` and `--behavior-delta` together let you detect when a change
to the linter (or to the codebase under test) alters the set of findings.
This is invaluable for refactor PRs that touch detection logic — reviewers
can see exactly which findings shifted without diffing raw output.

### Saving a baseline

Run the linter once against your target codebase, saving the findings to a
JSON file:

```bash
go-humanize-linter --save-baseline baseline.json ./...
```

The baseline file is a JSON array of `{rule, file, line}` entries. The
suggestion message text is intentionally excluded — only the
(rule, file, line) tuple is compared, so rewording a message never causes a
delta. See `docs/adr/0004-behavior-delta.md` for the rationale.

### Comparing against a baseline

After making changes, re-run with `--behavior-delta`:

```bash
go-humanize-linter --behavior-delta baseline.json ./...
```

Exit codes:

| Exit | Meaning                                                     |
| ---- | ----------------------------------------------------------- |
| 0    | No delta — the (rule, file, line) set matches the baseline. |
| 1    | Findings added or removed since the baseline.               |
| 2    | Linter error (baseline file missing, unreadable, etc.).     |

The delta is printed to stderr in human-readable form. Findings whose message
changed but whose (rule, file, line) tuple is unchanged are **not** reported
as a delta — only structural changes (added/removed findings) trigger exit 1.

### CI integration

A typical PR-check workflow:

1. The `main` branch CI saves `baseline.json` as a build artifact.
2. PR builds download `baseline.json` and run with `--behavior-delta`.
3. A non-zero exit fails the check, surfacing the finding diff to the reviewer.

This catches silent regressions (a refactor that drops a finding) and silent
noise additions (a detector broadening that introduces new findings) equally.
