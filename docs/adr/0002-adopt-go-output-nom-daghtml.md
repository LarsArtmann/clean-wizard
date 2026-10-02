# ADR-0002: Adopt `go-output` (nom live progress + daghtml HTML reports)

**Date:** 2026-10-01
**Status:** Accepted (opt-in first release; revisit triggers defined)
**Deciders:** Lars Artmann (full-execution directive 2026-10-01 — the three gate
questions from `docs/status/2026-10-01_21-03_GO-OUTPUT-ADOPTION-RESEARCH.md`
section (g) were resolved to the plan's recommended defaults under the owner's
"execute the whole plan" instruction)

---

## Context

clean-wizard's clean/scan runs are **completely silent** for their entire
duration (`clean.go:114` prints "Starting cleanup..." then nothing until the
post-run table). Minutes-long runs give the user zero feedback — the biggest UX
gap in the CLI.

The 2026-10-01 research session verified that `go-output`'s `nom` (live inline
progress) and `daghtml` (self-contained HTML report) modules are adoptable:

| Fact (all verified this session)                                                                                                           | Evidence                                                                                                                      |
| ------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| All modules resolve from the public proxy at `v0.38.2`                                                                                     | `go list -m`; scratch module build (P02)                                                                                      |
| Zero dependency conflicts — `charm.land/lipgloss/v2 v2.0.6` shared                                                                         | scratch go.mod MVS check with clean-wizard's exact pins (M08)                                                                 |
| **BuildFlow dogfoods nom/tui** — direct source imports + regression tests                                                                  | `BuildFlow/execution/nom_status.go`, `tui_bridge*.go`; `universal-workflow/go.mod` requires `go-output/nom v0.38.2` (P01/M01) |
| MIT license at repo root covers the submodules (no per-module LICENSE except graph)                                                        | `go-output/LICENSE`, P01/M03                                                                                                  |
| go-workflow v0.1.13 fires `BeforeStep`/`AfterStep` **per retry attempt**; `NextBackOff` receives every `RetryEvent`                        | go-workflow `workflow.go:438-475`; clean-wizard owns the hook at `internal/execution/retry.go:131`                            |
| resultCollector is **last-wins** under real retries (single entry, final outcome)                                                          | `TestRunCleaners_Retry`, `TestRunCleaners_SmartRetry_Transient` (P04, green)                                                  |
| Failed→Retrying narrative renders correctly: completed line keeps `✔ name ⟳1 (Transient)`                                                  | P03 mock harness, pipe + TTY captures                                                                                         |
| Flat-fan layout is readable: running pinned on top, ✔/○ stacks, live counts box with %                                                     | P03 pipe capture                                                                                                              |
| **Default timing cache path is `~/.cache/nom-timing.csv`** — shared with BuildFlow; **failed-attempt durations are recorded into medians** | `nom/timing_cache.go:18,68`, `nom/subscriber_handlers.go:155-160` (P03/M14)                                                   |

Risks: go-output is pre-1.0 (57 commits across these modules in 2 months;
v1.0.0 pending). Mitigation: exact pin + deliberate upgrades (fleet pattern:
BuildFlow and library-policy run the identical `v0.38.2`).

## Decision

**ADOPT** `github.com/larsartmann/go-output` at exact pin **`v0.38.2`**, in order:

1. **W1 — nom live progress** on `clean` (then `scan`), behind an **opt-in
   `--progress` flag** for the first release.
2. **W2 — daghtml post-run report** behind `--report <path>`.
3. **W3 — tui** (`--tui` spike) and **W4 — graph** (`scan --graph`) evaluated
   after W1 ships; neither blocks this decision.

### Binding design rules (from the verified facts)

- **Gating (safety rail #1):** the renderer activates only when
  `TTY && !--json && !--sarif && --progress`. Every other mode emits byte-identical
  output to today.
- **Version pin policy:** exact `v0.38.2`; upgrades are deliberate one-line
  changes with a changelog read (fleet-pinned, like BuildFlow).
- **Timing cache isolation (mandatory):** construct the subscriber with
  `WithCachePath(<os.UserCacheDir()>/clean-wizard/nom-timing.csv)`. The default
  `~/.cache/nom-timing.csv` collides with BuildFlow's cache, and activity names
  are bare (e.g. `nix` in both apps) — without isolation the two apps poison
  each other's ETAs.
- **Retry event design:** emit `ActivityRetrying{Attempt, Reason}` directly for
  transient failures WITHOUT an intermediate `ActivityFailed` — because
  `handleActivityFailed` records the failed attempt's duration into the timing
  cache (verified), polluting medians. Emit `ActivityFailed` only when retries
  are exhausted (final outcome). `Reason` comes from the error family name.
- **Skipped cleaners (g3 = option c):** register every selected cleaner upfront
  (`ActivityRegistered`), and an unavailable cleaner is instantly completed with
  an `EnqueueLines` note `• <name>: skipped (unavailable)` — nom has no skip
  status; a vanished row would look like a bug.
- **No cleaner/DI/registry changes:** the bridge lives in `internal/execution`
  seams (hooks, retry options) plus a RunOption-injected, nil-safe emitter.

### Owner answers recorded (g1–g3)

| Question                                 | Answer adopted                                                                                   |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------ |
| g1 progress default-on vs opt-in         | **Opt-in `--progress`** first release; flip to default-on only after a release of field evidence |
| g2 pre-1.0 pin acceptable vs wait v1.0.0 | **Pin `v0.38.2` now**; deliberate upgrade policy; revisit at go-output v1.0.0                    |
| g3 skipped-cleaner rendering             | **Option (c)**: instant-complete + "skipped (unavailable)" note                                  |

## Alternatives considered

- **Status quo (silent runs):** rejected — the UX gap is the strongest
  user-facing complaint surface; all adoption risks have verified mitigations.
- **tui first (BubbleTea full app):** rejected for now — inline nom is
  non-interactive-safe (pipes/CI degrade automatically), while a full TUI
  changes the input contract; tui stays a W3 spike over the same event bridge.
- **Wait for go-output v1.0.0:** rejected — v1.0.0 has no date; the fleet
  (BuildFlow, universal-workflow, library-policy) already runs v0.38.2 in
  production paths, which is stronger evidence than a version number.
- **go-dag-app for orchestration:** out of scope — rejected in ADR-0001;
  revisit triggers remain cold.

## Consequences

- Positive: minutes-long runs become legible (progress + ETA + retry
  visibility); post-run HTML artifact for humans; shared-event vocabulary with
  BuildFlow's execution layer.
- Negative/accepted: one pre-1.0 dependency pinned exactly; `nom`'s
  bold-color codes survive `NO_COLOR` in some paths (cosmetic, verified P03;
  plain-text append-only degradation still holds in pipes/CI).
- Neutral: timing cache introduces persistent state under the user cache dir;
  isolated to `clean-wizard/` namespace.

## Revisit triggers

1. go-output releases `v1.0.0` → re-read changelog, consider unpinned minor.
2. nom changes the event sum type or `WithCachePath` semantics in a breaking
   way → re-run the P03 harness before upgrading.
3. `--progress` proves stable for one release → decide default-on (g1 flip) in
   a follow-up ADR appendix.
4. BuildFlow diverges to a newer go-output with features clean-wizard needs →
   fleet-wide coordinated bump via buildflow update.

## Verification artifacts

- P01 consumer/license check: commands + outputs in session log 2026-10-01.
- P02 scratch build: `/tmp/nom-scratch` (nom@v0.38.2 + clean-wizard pins, MVS green).
- P03 render harness: `/tmp/nom-scratch/main.go` (pipe + pty captures `out-pipe.txt`, `out-tty.txt`).
- P04 collector semantics: `internal/execution/integration_test.go:107,205` (green).

## Appendix A (2026-10-02): tui-vs-nom default decision (W3 spike, P16–P17)

**Decision: nom (InlineRenderer) remains the default and only non-experimental
live-progress mode. The BubbleTea TUI ships as `--tui` (clean only),
EXPERIMENTAL, opt-in. Revisit before promoting.**

### Spike evidence

- Adapter: `internal/progress/tui.go` maps the identical ProgressEmitter
  vocabulary onto `tui.BubbleTeaProgressReporter` (compile-time contract
  proven; event dispatch tests green, `-race`).
- Cancel wiring: `NewTUIEmitter` wraps the command context and registers the
  cancel func, so ctrl+c aborts the run gracefully (plan M76).
- Display mode pinned to `DisplayModeNOM` — the TUI renders the same tree
  narrative as the inline frame, plus zoom/scroll/click-to-highlight.

### Why not default

1. **Timing-cache isolation violation (ADR-0002 binding rule):** tui v0.38.2
   constructs its internal NOM subscriber via bare `nom.NewNOMSubscriber()`
   (reporter.go, `NewProgressModel`) — the default `~/.cache/nom-timing.csv`
   path BuildFlow also uses. There is no option to inject `WithCachePath`.
   Adopting tui as default would break the cache-isolation rule we just paid
   to establish.
2. **No real-run verdict:** the nom frame was validated on a 7m46s real clean
   run; the TUI was validated at contract level only. Promoting a default on
   less evidence than the incumbent would be ass-backwards.
3. **No user pull yet:** the nom frame already answers the "silent minutes"
   problem. Interactive extras (zoom, scroll) are nice-to-have, not pain.

### Revisit triggers (append to the list above)

5. go-output/tui exposes a cache-path/subscriber injection option → re-run the
   spike; if a real-run comparison shows a genuine UX win, promote via a new
   ADR appendix.
6. A user (or owner) actually wants click-to-highlight/zoom during long runs.
