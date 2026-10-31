// Package humanizelint detects hand-rolled reimplementations of
// github.com/dustin/go-humanize and suggests the library function instead.
//
// The linter scans Go source files using AST pattern matching. Each rule looks
// for a cluster of signals that, together, strongly indicate a manual
// reimplementation of a specific go-humanize feature. Single weak signals are
// never enough — the rules require multiple corroborating signals within the
// same function to keep false positives near zero.
//
// # Rules
//
//	H001  manual-bytes-format     → humanize.Bytes / humanize.IBytes
//	H002  manual-comma-format     → humanize.Comma / humanize.Commaf
//	H003  manual-reltime-format   → humanize.RelTime / humanize.Time
//	H004  manual-plural           → english.Plural / english.PluralWord
//	H005  manual-si-format        → humanize.SI / humanize.SIWithDigits
//	H006  manual-ftoa             → humanize.Ftoa / humanize.FtoaWithDigits
//	H007  manual-parse-bytes      → humanize.ParseBytes
//	H008  manual-ordinal          → humanize.Ordinal
//	H009  manual-commaf           → humanize.Commaf / humanize.CommafWithDigits
//	H010  manual-comma-parse      → humanize.ParseComma / humanize.ParseCommaf
//	H012  manual-word-series      → english.WordSeries / english.OxfordWordSeries
//
// Usage as a library:
//
//	import (
//	    "github.com/larsartmann/go-finding"
//	    humanizelint "github.com/larsartmann/go-humanize-linter"
//	    "github.com/larsartmann/go-linter-sdk"
//	)
//
//	reg := humanizelint.DefaultRegistry()
//	report, err := reg.Run(ctx, ".")
//	// Ternary exit code: 0 = clean, 1 = high-confidence findings (must fix),
//	// 2 = only medium/low-confidence findings remain (triage).
//	fmt.Println(linter.ExitCodeByConfidence(report, finding.ConfidenceHigh))
package humanizelint
