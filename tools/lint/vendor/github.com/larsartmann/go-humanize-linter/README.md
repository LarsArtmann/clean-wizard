# go-humanize-linter

An AST-based linter that detects hand-rolled reimplementations of [go-humanize](https://github.com/dustin/go-humanize) and suggests the library function instead.

Built on [go-linter-sdk](https://github.com/larsartmann/go-linter-sdk) and [go-finding](https://github.com/larsartmann/go-finding).

## Why?

go-humanize provides 30+ functions for formatting numbers, bytes, times, and English text. Across a typical Go codebase, developers reinvent the same features — byte-size formatting, comma separators, relative time, pluralization — dozens of times, often with subtle bugs (wrong rounding, missing edge cases, inconsistent unit labels).

This linter finds those reimplementations automatically.

## Rules

| Rule | Name                  | Detects                                                    | Suggests                                          |
| ---- | --------------------- | ---------------------------------------------------------- | ------------------------------------------------- |
| H001 | manual-bytes-format   | Byte-size formatting (1024 division + unit strings)        | `humanize.Bytes` / `humanize.IBytes`              |
| H002 | manual-comma-format   | Comma/thousands separator insertion (digit grouping loops) | `humanize.Comma`                                  |
| H003 | manual-reltime-format | Relative time formatting ("3 hours ago")                   | `humanize.RelTime` / `humanize.Time`              |
| H004 | manual-plural         | English pluralization (`if n == 1`)                        | `english.Plural` / `english.PluralWord`           |
| H005 | manual-si-format      | SI-prefix formatting ("1.2K", "3.4M")                      | `humanize.SI` / `humanize.SIWithDigits`           |
| H006 | manual-ftoa           | Trailing-zero stripping (`strings.TrimRight` nesting)      | `humanize.Ftoa`                                   |
| H007 | manual-parse-bytes    | Byte-size string parsing (HasSuffix chains, mult maps)     | `humanize.ParseBytes`                             |
| H008 | manual-ordinal        | Ordinal formatting (`switch n%10` with st/nd/rd/th)        | `humanize.Ordinal`                                |
| H009 | manual-commaf         | Float-with-comma formatting (`%.Nf` + separator loop)      | `humanize.Commaf`                                 |
| H010 | manual-comma-parse    | Comma-grouped number parsing (strip + `strconv` parse)     | `humanize.ParseComma` / `ParseCommaf`             |
| H012 | manual-word-series    | Word-series joining ("a, b, and c" assembly)               | `english.WordSeries` / `english.OxfordWordSeries` |

`Bytes`/`IBytes` also ship `BytesN`/`IBytesN` min-digits variants (go-humanize v1.1.0+); `ParseComma`/`ParseCommaf` require go-humanize v1.1.0+.

## Detection Strategy

### Examples of detected patterns

```go
// H001: "KMGTPE"[exp] trick
func formatBytes(bytes int64) string {
    const unit = 1024
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// H002: comma grouping loop
func formatNumber(n int) string {
    for i, r := range str {
        if (len(str)-i)%3 == 0 { result.WriteString(",") }
    }
}

// H003: relative time
func relativeTime(t time.Time) string {
    elapsed := time.Since(t)
    if elapsed < time.Hour { return strconv.Itoa(int(elapsed.Minutes())) + "m ago" }
}

// H004: pluralization
func pluralize(n int, singular, plural string) string {
    if n == 1 { return fmt.Sprintf("%d %s", n, singular) }
    return fmt.Sprintf("%d %s", n, plural)
}

// H010: comma-grouped number parsing
func parseCount(s string) (int64, error) {
    return strconv.ParseInt(strings.ReplaceAll(s, ",", ""), 10, 64)
}
```

## Usage

### As a library

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/larsartmann/go-finding"
    humanizelint "github.com/larsartmann/go-humanize-linter"
    "github.com/larsartmann/go-linter-sdk"
)

func main() {
    registry := humanizelint.DefaultRegistry()
    report, err := registry.Run(context.Background(), ".")
    if err != nil { panic(err) }

    for f := range report.All() {
        fmt.Printf("%s:%d [%s] %s\n", f.Position.File, f.Position.Line, f.Rule, f.Message)
    }

    os.Exit(linter.ExitCodeByConfidence(report, finding.ConfidenceHigh))
    // Ternary exit code: 0 = clean, 1 = high-confidence findings (must fix),
    // 2 = only medium/low-confidence findings remain (triage).
}
```

### CLI

```bash
# Scan a path (text output, exit 1 if findings, 0 if clean)
go-humanize-linter ./...

# JSON or SARIF output
go-humanize-linter --format json ./...
go-humanize-linter --format sarif ./... > results.sarif

# Write report to a file instead of stdout
go-humanize-linter --output report.txt ./...
go-humanize-linter --output results.json --format json ./...
go-humanize-linter --output results.sarif --format sarif ./...

# Enable / disable specific rules
go-humanize-linter --enable H001 --enable H003 ./...
go-humanize-linter --disable H004 ./...

# Load enable/disable rules from a YAML config file
go-humanize-linter --config .gohumanize.yaml ./...

# YAML config file format (.gohumanize.yaml):
#   enable:
#     - H001
#     - H003
#   disable:
#     - H004
# CLI flags are merged on top of config file values (union semantics).

# List rules or print version
go-humanize-linter --rules
go-humanize-linter --version

# Explain a rule or list which files would be scanned
go-humanize-linter --explain H001
go-humanize-linter --list-files ./...

# Filter by confidence (low, medium, high, full)
# Useful for triage: high/full findings are strong signals, medium/low need review.
go-humanize-linter --min-confidence high ./...

# Verify suppression directives are still needed
go-humanize-linter --verify-suppressions ./...

# Save a baseline and compare future runs against it
# First, save the current findings as a baseline:
go-humanize-linter --save-baseline baseline.json ./...
# Then in CI, fail only when findings are added or removed:
go-humanize-linter --behavior-delta baseline.json ./...
# Note: upgrading the linter (e.g. to v0.3.0, which adds H010) adds findings
# to the report — regenerate the baseline once after upgrading and review the
# delta before committing it.
```

### As a GitHub Action

```yaml
- uses: LarsArtmann/go-humanize-linter@v0.5.0
  with:
    path: ./...
    # enable: H001,H003     # only run these rules
    # disable: H004          # skip these rules
    # format: sarif          # text (default), json, or sarif
    # min-confidence: high   # low (default), medium, high, full
    # verify-suppressions: true  # report stale //nolint directives
    # behavior-delta: baseline.json  # fail if findings changed vs baseline
    # save-baseline: baseline.json   # save current findings as baseline
```

### Suppressing findings

Add a `//nolint:gohumanize` directive to suppress findings. The directive can
be placed on the function declaration, in the doc comment, or anywhere inside
the function body (on the specific statement that triggers the finding).

```go
// On the function declaration:
//nolint:gohumanize // intentional hand-rolled format
func prettySize(b int64) string {
    return fmt.Sprintf("%.1f %cB", float64(b)/1048576, "M")
}

// Inside the function body (on the triggering line):
func formatBytes(b int64) string {
    return fmt.Sprintf("%.1f %cB", float64(b)/1048576, "M") //nolint:gohumanize
}
```

Recognised forms:

| Directive                       | Effect                                           |
| ------------------------------- | ------------------------------------------------ |
| `//nolint`                      | Suppresses all linters                           |
| `//nolint:all`                  | Suppresses all linters                           |
| `//nolint:gohumanize`           | Suppresses all gohumanize rules on this function |
| `//nolint:gohumanize:H001`      | Suppresses only H001 (scoped)                    |
| `//nolint:gohumanize:H001,H002` | Suppresses H001 and H002 only                    |
| `//nolint:gohumanize,other`     | Suppresses gohumanize and another linter         |

A trailing `// reason` comment is allowed on any form.

## Build & Test

```bash
nix run .#test       # run tests
nix run .#lint       # run golangci-lint
nix run .#build      # build all packages
```

Direct Go commands require `GOEXPERIMENT=jsonv2` (a dependency uses `encoding/json/v2`).

## Requirements
