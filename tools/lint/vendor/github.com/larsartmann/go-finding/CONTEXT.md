# go-finding Domain Language

**Project:** go-finding — Unified data model and pipeline for static analysis tools.

---

## Core Concepts

### Finding

A single issue detected by a static analysis tool. The central data type of the library.

### Position / Range

Location in source code. `Position` is a point (file, line, column, offset). `Range` is a span from Start to End.

### Severity

How serious a Finding is: `info`, `warning`, `error`, `critical`.

### FixStrategy

How a Finding can be remediated: `none`, `suggest`, `direct`, `ai`.

### Category

Domain classification: `correctness`, `security`, `performance`, `style`, `unused`.

### Suppression

A Finding may be suppressed (ignored) either in source code or via config, optionally with an expiry.

### RelatedRef

Links between Findings (e.g., "clone-of", "causes", "wraps").

### Report

A container for Findings from a single tool run, with aggregated Summary statistics.

---

## Pipeline Concepts

### Pipeline

Orchestrates the `detect → triage → fix → verify` loop across multiple iterations.

### Detector

Interface for anything that can find issues: `Name() string`, `Detect(ctx) ([]Finding, error)`.

### Triage

Categorizing Findings by FixStrategy into Direct, Suggest, and None buckets.

### FixApplier

Applies `direct` fixes to source files with backup/rollback support.

### ConflictDetector

Identifies overlapping fixes that cannot be safely applied together.

### Verifier

Re-runs Detectors after fixes to verify what was resolved vs what remains or is new.

### Metrics

Optional timing and count collection for pipeline stages and detectors.

### RetryDetector

Wrapper that retries a Detector with exponential backoff.

### Partial Detection

Graceful degradation: collect findings from successful detectors even when others fail.

---

## Output Formats

### SARIF

Static Analysis Results Interchange Format (2.1.0). Lossless round-trip via Properties.

### LSP

Language Server Protocol Diagnostic format. Lossless round-trip via `LSPDiagnosticData`.

### Diagnostic

go/analysis integration. Converts `analysis.Diagnostic` to Finding.

---

## Key Invariants

- Findings are immutable data, not state machines.
- Core types depend only on stdlib.
- Conversions preserve all data via Metadata / Properties.
- Pipeline is NOT safe for concurrent `Run()` — create a new one per invocation.
