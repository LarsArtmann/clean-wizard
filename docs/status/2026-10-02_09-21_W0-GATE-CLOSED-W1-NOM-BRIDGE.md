# Status Report: W0 Gate Closed + W1 nom Bridge Implemented

**Date:** 2026-10-02 09:21
**Session type:** Full execution of the go-output adoption plan (`docs/planning/2026-10-01_23-12_GO-OUTPUT-ADOPTION-PARETO-PLAN.md`) under the owner's "execute the whole todo list" directive.
**Scope:** This session only. Covered P01–P09 of the plan (verification gate, ADR-0002, nom bridge implementation + live smoke). P10–P24 not started.

---

## a) FULLY DONE

### 1. W0 verification gate — all four unverified claims closed (P01–P04)

| Claim / question                                                | Verdict                                  | Evidence                                                                                                                                                                                                                                                                      |
| --------------------------------------------------------------- | ---------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "BuildFlow dogfoods nom" (worst slip of last session)           | **TRUE — stronger than claimed**         | BuildFlow has DIRECT source imports: `BuildFlow/execution/nom_status.go`, `tui_bridge*.go`, `nom_regression_test.go`, `nom_visibility_test.go`; `universal-workflow/go.mod:16` requires `go-output/nom v0.38.2` directly; `library-policy/go.mod` consumes v0.38.2 (indirect) |
| License coverage of submodules (b5)                             | Covered                                  | MIT at repo root (`go-output/LICENSE`); only `graph/` ships its own LICENSE; no per-module licenses elsewhere                                                                                                                                                                 |
| nom@v0.38.2 resolves + builds with clean-wizard's pins (b4/M08) | Green                                    | scratch module `/tmp/nom-scratch`: `go get nom@v0.38.2` + clean-wizard charm pins (huh v2.0.3, lipgloss v2.0.6, log v2.0.1, bubbles v2.2.1, bubbletea v2.0.10) + tidy + build all green                                                                                       |
| Failed→Retrying render narrative (b2)                           | VERIFIED rendered                        | P03 harness: completed line keeps `✔ docker ⟳1 (Transient)`; skip note drains above frame; terminal failure keeps ⚠ in final tree                                                                                                                                             |
| resultCollector retry semantics (b3)                            | Last-wins confirmed, tests exist + green | `TestRunCleaners_Retry` (1 entry, final outcome kept, 3 attempts) + `TestRunCleaners_SmartRetry_Transient` (exhausted → single Failed) — `recordFinal` overwrites by name                                                                                                     |

### 2. P03 mock render harness (f2/f5) — empirical verdicts, not reasoning

Ran `/tmp/nom-scratch/main.go` (14 flat cleaners + 1 transient retry + 1 skip + 1 terminal failure) in pipe AND pty modes:

- Flat fan is readable: running cleaner pinned on top, ✔/○ stacks, live counts box `⏵1 ✔2 ○11 701ms Σ14 (14%)`.
- Pipe mode degrades to append-only plain text (no cursor codes) — CI-safe.
- TTY mode does in-place cursor-up repaint (`ESC[17A` + clear-lines), restores cursor at Finish.
- **Timing cache facts (verified in nom source, now load-bearing for the bridge):**
  - Default path is `~/.cache/nom-timing.csv` → **collides with BuildFlow** (same default, bare activity names like `nix` in both apps). `WithCachePath` isolation is mandatory, not optional.
  - `handleActivityFailed` **records failed-attempt durations into the timing cache** → emitting ActivityFailed per retry attempt would skew medians. Bridge therefore emits `ActivityRetrying` directly and `ActivityFailed` only on exhausted retries.
  - go-workflow v0.1.13 calls `NextBackOff` ONLY when the backoff source did not return `backoff.Stop` (`retry.go:97-103`) → it is exactly the "a retry WILL happen" signal; `RetryEvent.Attempt` is 0-based (human attempt = +1).

### 3. P05 — ADR-0002 written, GATE = ADOPT

`docs/adr/0002-adopt-go-output-nom-daghtml.md`: adopt `go-output` at exact pin `v0.38.2`; nom first (W1), daghtml second (W2), tui/graph conditional. Binding rules: gating `TTY && !--json && !--sarif && --progress`; mandatory timing-cache isolation; retry-event design (Retrying direct, Failed only on exhaust); skip = instant-complete + note. The three owner questions (g1–g3) resolved to the plan's recommended defaults under the owner's full-execution directive, recorded in the ADR.

### 4. W1 implementation (P06–P09) — nom bridge code complete + live smoke green

- **P06** `internal/execution/progress.go`: `ProgressEmitter` interface (10 methods — under the interfacebloat limit after merging `WorkflowCompleted`/`WorkflowFailed` into `WorkflowFinished(err)`), `noopProgressEmitter` default, `WithProgress` RunOption, `runConfig.emitter()` nil-safe accessor.
- **P07** lifecycle wiring: `WorkflowStarted` + upfront `ActivityRegistered` batch in `executeWorkflow` (names from new `collector.registeredNames()`); `ActivityStarted` in `makeBeforeHook` (fires per attempt — correct narrative); terminal per-step outcomes emitted ONCE after `Do()` from final collector state (`emitTerminalOutcome`: Completed/Skipped/Failed) — the single place that knows retries are exhausted.
- **P08** retry + skip: `retryOptions(cfg, em, name)` emits `ActivityRetrying(name, attempt+1, errorfamily.Classify(err).String())`; `ActivitySkipped` drains a note + instant-completes (g3 option c).
- **P09** `internal/progress` package: nom-backed `Emitter` (compile-time `var _ execution.ProgressEmitter`), `Enabled()` TTY gate, `DefaultCachePath()` = `<user cache dir>/clean-wizard/nom-timing.csv` with MkdirAll + temp fallback; `--progress` flag on clean + scan; shared `progressRequested()` gate helper; verbose output rerouted through `verboseLine` → `Note` (renderer lane) or `io.WriteString` (plain) — hooks.go forbidigo violations fixed in passing.
- **M34 live smoke on the real system (pty):** `clean --dry-run --yes --progress --mode quick` ran 5 real cleaners for 7m46s with a live frame (`⏵5 7m46s Σ5 (0%)`), drained `• homebrew: skipped (homebrew not available)` note, final ✔ tree, and the normal post-run table. The feature works end-to-end on real cleaners.
- Full short test suite **exit=0** (all packages, incl. commands after signature updates).

## b) PARTIALLY DONE

1. **W1 exit criteria NOT met yet:** P10 (fake-emitter unit tests incl. race), P11 (Ginkgo BDD spec), P12 (manual matrix + flag docs + FEATURES.md/README) are not done; the non-TTY zero-output-diff criterion is not yet byte-verified (M45 test missing).
2. **Lint verification of touched files not run** this session (LSP diagnostics proved stale mid-session; a targeted `buildflow -s golangci-lint` pass is queued as P22/M98). Pre-existing warnings remain (FreedBytes deprecated in hooks.go, mnd in retry.go, ireturn in container.go, forbidigo in commands).
3. **M44 env spot-checks partial:** observed in the pipe capture that `NO_COLOR=1` did NOT fully strip styling codes (append-only degradation still held — the critical part). CI env and TERM=dumb not exercised. Needs the M44 pass before P12 signs the matrix.
4. **bubbletea/v2 version drift noted:** clean-wizard go.mod now has `bubbletea/v2 v2.0.10` (last session's research said v2.0.9). Scratch MVS passed against it; not treated as a problem — flagged as a question below.
5. **Daemon commit behavior:** working tree is clean and the last three commits are daemon auto-commits — session work appears committed, but I did not verify each file landed in a commit (buildflow skill warns pma has blind spots).

## c) NOT STARTED

- P10 bridge unit tests (fake subscriber/emitter: happy, retry, skip sequences; race; status-mapping tables)
- P11 Ginkgo BDD live-progress narrative spec
- P12 manual test matrix documentation + `--progress` flag help review + FEATURES.md + README
- M45 zero-progress-bytes test for `--json`/`--sarif`
- W2 entirely: P13 WorkflowResult→daghtml mapping + goldens, P14 `--report` flag, P15 README section
- W3: P16 `--tui` spike, P17 tui-vs-nom decision (ADR-0002 appendix)
- W4: P18 `scan --graph mermaid|dot`, P19 architecture-docs regeneration
- X: P20 AGENTS.md memory entries + `container.go:27` shutdown-error Debug log, P21 docs-health HARVEST, P22 hygiene (FreedBytes→SizeEstimate, ireturn, mnd), P23 commands print-helper cleanup, P24 website DAG embed
- ADR-0002 not yet linked from project AGENTS.md (planned under P20)

## d) TOTALLY FUCKED UP

1. **First live smoke run hung for minutes (user noticed):** I launched `clean --dry-run --yes --progress` under a pty WITHOUT `--mode`/`--profile` — the huh interactive selector rendered and waited for keyboard input that never came. The research itself documented this exact branch (`clean_select.go:34-43`); I knew it and still walked into it. Killed the job; reran with `--mode quick` which worked. Rule adopted: smoke tests of clean MUST pin a non-interactive surface.
2. **Write-tool rejection spiraled into a false renderer bug:** the P03 harness `main.go` was created via bash heredoc, so the write tool rejected my overwrite ("modified since last read") — and I built and ran the OLD stub, saw empty pipe output, and spent a debugging cycle reading the renderer's plain-text draw path before checking the file content. The empty output was the stub exiting silently. Cost: ~2 tool rounds + wrong first hypothesis.
3. **Trusted stale LSP diagnostics over the compiler:** mid-edit diagnostics kept showing forbidigo/typecheck findings that no longer existed (or had moved). I re-checked files repeatedly; `go build` was green the whole time. The compiler (and later the real lint run) is the source of truth.
4. **Self-inflicted formatting slip:** a careless multiedit on hooks.go removed the import-group blank line; caught and restored, but it was avoidable noise.

## e) WHAT WE SHOULD IMPROVE

1. **Smoke tests must be non-interactive by construction** — `--mode`/`--json` pinned, or feed stdin; never rely on reflexes to notice a hang.
2. **One mechanism per file:** files created via bash heredoc must be re-read before tool-writes (or keep single-writer discipline). This caused failure d2.
3. **Timebox long commands up front:** the full short suite legitimately takes ~10 min on this machine; the first run should have been backgrounded with a log file (as the rerun was), not foregrounded.
4. **Automate the W1 invariance check now:** a golden test that byte-compares `--json`/`--sarif` output with and without `--progress` would close the last unverified safety rail (M45) and prevent regressions forever.
5. **Upstream hygiene candidate:** the NO_COLOR quirk (colors survive in plain-text logs) is worth verifying against nom's envdetect and possibly filing upstream — via verify-before-filing, only after reproduction.
6. **ADR binding rules work:** the timing-cache pollution fact (found in P03) directly changed the bridge design before any consumer code existed — the gate sequence paid for itself.

## f) Next tasks (up to 50 — brainstorm fuel, not commitments)

**Close W1 (do first):**

1. P10: fake ProgressEmitter test harness (record events, thread-safe) — S
2. P10: happy-path event sequence assertions (Started→Registered→Started→Completed order) — S
3. P10: retry-path assertions (Retrying between Starts, no intermediate Failed, attempt numbering) — S
4. P10: skip-path assertions (note + instant complete, no failure) — S
5. P10: terminal-failure assertions (single Failed after exhaust) — S
6. P10: race test — parallel cleaners emitting concurrently — M
7. P10: status-mapping table tests (succeeded/skipped/failed → emitter calls) — S
8. M45: `--json`/`--sarif` emit zero progress bytes (byte-compare goldens, with/without flag) — S
9. M44: NO_COLOR / TERM=dumb / CI=true env spot-checks — S
10. P11: Ginkgo Describe skeleton for live progress — S
11. P11: It: all cleaners pending upfront — S
12. P11: It: retry renders ⟳ attempt suffix — S
13. P11: It: skip renders note per decision — S
14. P11: It: json mode silent — S
15. P11: run full suite; de-flake — M
16. P12: manual matrix doc (TTY/pipe/CI/NO_COLOR/--json/--sarif/dry-run/real) — S
17. P12: `--progress` flag help text final review — XS
18. P12: FEATURES.md live-progress entry — S
19. P12: README quick section (screenshot-worthy example) — S
20. Verify daemon commits contain all session files (`git log --stat` spot-check) — XS

**Cross-cutting X wave (cheap, high durability):**
21. P20: AGENTS.md — go-workflow per-attempt hook semantics gotcha — XS
22. P20: AGENTS.md — timing-cache isolation rule + WithCachePath pointer — XS
23. P20: AGENTS.md — go-output adoption pointer (ADR-0002 + this report) — XS
24. P20: AGENTS.md — nom emit-contract (Retrying direct, Failed on exhaust) — XS
25. M90: container.go:27 shutdown-error Debug log + test (go-dag-app nugget) — XS
26. P21: docs-health HARVEST of this section (f) into TODO_LIST.md / ROADMAP.md — S
27. P22: hooks.go FreedBytes → SizeEstimate (2 sites) — XS
28. P22: container.go ireturn — nolint-with-reason or refactor — XS
29. P22: retry.go mnd magic numbers → named consts — XS
30. P22: targeted `buildflow -s golangci-lint` verification on touched files — S

**W2 (daghtml report):**
31. P13: WorkflowResult → daghtml.DAG builder — S
32. P13: color/tooltip mapping ("freed X | items | duration") — S
33. P13: golden-file test for report builder — S
34. P13: failure-node case (Error dot + tooltip) — S
35. P13: dry-run variant sanity check — XS
36. P14: `--report <path>` flag on clean + scan, exclusive with --json/--sarif (Rejection on conflict) — S
37. P14: classified error handling on report write failure — S
38. P14: integration test: dry-run --report → tmpfile exists + valid HTML — S
39. P15: README "HTML reports" section — XS

**W3 (tui):**
40. P16: `--tui` spike over the same emitter (BubbleTeaProgressReporter) — M
41. P16: SetCancelFunc → command ctx (graceful abort) — S
42. P16: spike notes + go/no-go — S
43. P17: tui-vs-nom default decision in ADR-0002 appendix — S

**W4 (graph/docs/website):**
44. P18: `scan --graph mermaid|dot` from registry + goldens — S
45. P19: regenerate architecture docs from code (replace 5-month-stale D2) — M
46. P24: website cleaner-pipeline DAG embed via GraphHTML — M

**Follow-ups from this session:**
47. Investigate NO_COLOR quirk in nom envdetect; reproduce, then consider upstream issue — S
48. Confirm bubbletea/v2 v2.0.9→v2.0.10 fleet bump was deliberate (see g2) — XS
49. Consider `ActivityProgress` events (nom sub-step messages) for long cleaners (nix, docker) as a W1.5 enhancement — M
50. Flip g1 (progress default-on for TTY) decision point after one release of `--progress` field evidence — owner

## g) Questions I cannot answer myself

1. **Real-run durations confirm the feature's value — 7m46s of silence on a quick-mode dry-run.** Given that, do you want to flip g1 now (progress default-on for interactive runs, opt-out via `--no-progress`) in this same release, or keep ADR-0002's opt-in `--progress` for the first release and revisit with field evidence?
2. **bubbletea/v2 v2.0.10:** clean-wizard's go.mod pins v2.0.10 while the 2026-10-01 research snapshot recorded v2.0.9 (identical pin claim was part of the zero-conflict verification). The scratch MVS passed with v2.0.10, so there is no conflict — but was this bump a deliberate fleet action I should record as fact, or drift the fleet should reconcile?
3. **Commit policy for the implementation waves:** the auto-commit daemon has committed this session's work as `chore: auto-commit N changed file(s)` entries. For ADR-0002 and the W1 feature work, do you want explicit conventional commits (like the planning session's `bc2f5c1`) instead, and should W2+ follow the same rule?

---

_W0 gate fully closed; W1 code complete and smoke-verified on a real 7m46s run. W1 tests/docs and W2+ remain per the Pareto plan. Waiting for instructions._
