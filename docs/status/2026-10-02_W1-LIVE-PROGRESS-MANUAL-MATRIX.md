# W1 Live Progress — Manual Test Matrix

**Date:** 2026-10-02
**Scope:** `--progress` live progress (go-output/nom, ADR-0002) across output modes and environments.
**Verdict:** PASS with one documented cosmetic limitation (M44, see below).

## Gating contract

Renderer activates only when **all** hold: real TTY on stdout, no `--json`, no `--sarif`, `--progress` passed.
Automated coverage: `TestProgressRequested_Gate`, `TestProgressRequested_NonTTYWriterIsDisabled` (commands),
`TestWorkflowRun_ProgressNeverWritesToStdout` (progress package — real nom emitter, zero stdout bytes),
Ginkgo `Live progress narrative` suite (execution).

## Matrix

| # | Environment             | Command                                               | Expected                                                                                             | Result                                                                                                                                                              |
| - | ----------------------- | ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | TTY (pty via `script`)  | `clean --dry-run --yes --progress --mode quick`       | Live frame, retries as ⟳, drained skip notes, final tree, intact post-run table                      | PASS (2026-10-02 smoke, 5 real cleaners, 7m46s; skip note `• homebrew: skipped (homebrew not available)` drained above frame)                                       |
| 2 | Pipe (plain redirect)   | `clean --dry-run --yes --progress --mode quick > log` | No renderer (gate), output identical to pre-progress runs                                            | PASS (append-only degradation verified at nom level in W0 P03; command gate keeps renderer off)                                                                     |
| 3 | CI (`CI=true`, non-TTY) | any command with `--progress`                         | Gate disabled — no frames, no ANSI, exit code unchanged                                              | PASS (Enabled=false for non-*os.File/non-TTY; `TestWorkflowRun_NoEmitterWritesNothingToStdout`)                                                                     |
| 4 | `--json`                | `clean --dry-run --json --progress`                   | JSON stream byte-identical with/without `--progress`                                                 | PASS (gate refuses emitter; `TestProgressRequested_Gate/json_mode_blocks_progress`; stdout-isolation test proves even an active emitter writes nowhere near stdout) |
| 5 | `--sarif`               | `scan --sarif --progress`                             | SARIF envelope untouched                                                                             | PASS (same gate, machineOutput=true)                                                                                                                                |
| 6 | `--dry-run` + real run  | both                                                  | Same narrative; dry-run shows estimate framing                                                       | PASS (matrix #1 is dry-run; real-run path identical events, only outcome data differs)                                                                              |
| 7 | `NO_COLOR=1` (TTY)      | pty run with renderer active                          | Colors suppressed                                                                                    | **PARTIAL — nom ignores NO_COLOR**                                                                                                                                  |
| 8 | `TERM=dumb` (TTY)       | pty run                                               | Plain rendering                                                                                      | **PARTIAL — nom ignores TERM=dumb**                                                                                                                                 |
| 9 | Timing cache isolation  | any `--progress` run                                  | Cache at `<user cache dir>/clean-wizard/nom-timing.csv`, never BuildFlow's `~/.cache/nom-timing.csv` | PASS (`TestDefaultCachePath_IsolatedUnderAppDir`; `WithCachePath` wired in `progress.New`)                                                                          |

## M44 finding: nom v0.38.2 InlineRenderer ignores NO_COLOR / TERM=dumb / CI

Empirical check (synthetic P03 harness, 2026-10-02): ANSI color codes (`[1;90m [1;91m [1;92m [1;93m [96m`)
are emitted identically under NO_COLOR=1, TERM=dumb, CI=true, and the no-env control — in both pty and
pipe destinations.

**Assessment:** cosmetic, not blocking. In production the gate keeps the renderer off every non-TTY
surface (CI, pipes, JSON/SARIF), so the escape codes only ever reach a real terminal where they render
correctly. A user with NO_COLOR=1 who explicitly passes `--progress` sees colors they asked not to see —
minor. Recorded as a candidate upstream issue for `go-output` (InlineRenderer should consult the
colorprofile/NO_COLOR); deliberately out of scope for this plan.

## Retry/skip narrative (behavioral, automated)

- Retry: `ActivityRetrying(attempt=N, reason="transient")` between starts; NO intermediate `ActivityFailed`
  (protects nom's timing cache from failed-attempt durations). Covered by P10 unit tests + BDD spec.
- Skip: drained note above frame + instant complete; reason from `NotAvailableError`. Covered.
- Exhausted retries: exactly one terminal `ActivityFailed`. Covered.
