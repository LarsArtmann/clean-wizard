# Status: go-output Adoption Plan — W1 Closed, W2+W3 Done, P20 Halfway (2026-10-02 10:23)

**Session scope:** continuation under the owner's full-execution directive ("NOW GET SHIT DONE — the WHOLE TODO LIST") for `docs/planning/2026-10-01_23-12_GO-OUTPUT-ADOPTION-PARETO-PLAN.md`. Entering state: W0 (P01–P05) closed, W1 code (P06–P09) done. Exiting state: P06–P19 complete, P20 halfway, P21–P24 + final verification remain.

## a) Fully done since last report

- **P10 — Bridge unit tests** (`internal/execution/progress_test.go`): thread-safe `recordingEmitter`; happy-path exact-sequence, nil-safety, retry (⟳ numbering + "transient" reason + zero intermediate failures), exhausted-retry (exactly one terminal `ActivityFailed`, after all retries), skip (one `ActivitySkipped`, `"some-tool not available"`), scan narrative, parallel `-race` safety, table-driven `emitTerminalOutcome` mapping, `reasonForSkip` fallback, `verboseLine` routing (Note vs writer). All green with `-race`.
- **TWO REAL W1 BUGS FOUND AND FIXED BY P10** (see (d)):
  1. `ActivityRegistered` double-emission (builder.go + executeWorkflow) with wrong ordering — removed the builder-side emission; `executeWorkflow` is now the single owner.
  2. `BuildScan` never wired Before/AfterStep hooks → scan steps jumped pending→completed with no `ActivityStarted`. Fixed; scan progress now narrates correctly.
- **M45 — zero-progress-bytes guarantee**, two layers: gate unit tests (`progressRequested` blocks json/sarif; non-TTY disabled) + the strongest form — `internal/progress/progress_test.go` runs a REAL nom emitter against a real workflow while `os.Stdout` is captured: zero stdout bytes, renderer sink provably active. Flip side test (no emitter → silent) included.
- **M44 — env-var spot-check**: empirical matrix on the P03 harness (pty + pipe × NO_COLOR/TERM=dumb/CI + control). **Finding: nom v0.38.2 InlineRenderer ignores all three — colors always emitted.** Cosmetic (gate keeps renderer off every non-TTY production surface); recorded in the matrix doc + flagged as upstream-issue candidates.
- **P11 — Ginkgo BDD spec** (`progress_bdd_test.go`, package `execution_test`): 5 narrative specs — plan-upfront-before-work, boundary close (Finish last), retry-not-failed narrative, skip-with-reason, machine-output silence (uninjected recorder sees zero events). 18/18 suite specs green.
- **P12 — W1 docs**: `docs/status/2026-10-02_W1-LIVE-PROGRESS-MANUAL-MATRIX.md` (9-row matrix, gating contract, M44 finding + assessment, retry/skip narrative pointers); FEATURES.md live-progress rows (clean + scan) + Last Updated bump; README `--progress` rows + scan flags table + "Live Progress" sales section.
- **P13 — daghtml mapping** (`internal/report/report.go`): `WorkflowResultToDAG` (status→color: success/--success, skip/--info, fail/--error + error dot; tooltips `freed 1.5 KiB | 3 items | 2.0 s` via SizeEstimate, NOT deprecated FreedBytes; registration order preserved), `WorkflowSummary` footer, `WriteWorkflowHTML`. Tests: mapping table, **golden nodes JSON** (`testdata/report-nodes.golden.json`, `-update-golden`), empty-result, self-contained-HTML (DOCTYPE, CSP, embedded JSON, no external scripts/links).
- **P14 — `--report <path>` on clean + scan**: `writeReportFile` helper (Rejection `clean.report_write` on file-create failure per config-save precedent, Corruption `*.report_render` on render failure); early Rejection on `--report`+`--json` (clean: `clean.report_json_conflict`) and `--report`+`--json/--sarif` (scan: `scan.report_machine_output_conflict`); confirmation line in human mode. Tests: exclusivity (both commands, sarif+json variants), self-contained artifact on disk, unwritable path classification.
- **P15 — README "HTML Reports" section** with examples and the exclusivity rule.
- **P16 — `--tui` spike** (`internal/progress/tui.go` + `--tui` EXPERIMENTAL flag on clean): `TUIEmitter` maps the full ProgressEmitter vocabulary onto `tui.BubbleTeaProgressReporter`; `NewTUIEmitter` wraps ctx + registers cancel (ctrl+c aborts the run gracefully); `DisplayModeNOM` pinned. Compile-time interface proof + no-panic dispatch tests green.
- **P17 — ADR-0002 Appendix A**: tui-vs-nom decision = **nom stays default; --tui experimental**. Reasons: (1) tui v0.38.2 constructs its internal NOM subscriber with the DEFAULT timing cache path (no `WithCachePath` injection) — violates our cache-isolation binding rule; (2) no real-run verdict (nom has a 7m46s real-run proof); (3) no user pull. Revisit triggers 5–6 appended.
- **P18 — `scan --graph mermaid|dot`**: `internal/report/graph.go` (`PipelineGraph` via go-output/graph, `WritePipelineGraph`, `IsSupportedGraphFormat`); golden tests for BOTH formats (`testdata/pipeline-{mermaid,dot}.golden`); scan wiring with early Rejections (`scan.graph_format`, `scan.graph_machine_output_conflict`) and print-and-exit before scanning; README flag rows + example.
- **P19 — architecture docs regenerated**: `docs/architecture-understanding/generate-module-graph.sh` (walks `go list` internal packages → Mermaid edges, rerunnable) + `2026-10-02_MODULE-GRAPH.md` (fresh graph + reading guide) explicitly superseding the 2026-05-03 D2 pair.
- **P20 (half) — container.go:27 shutdown errors**: cleanup closure now logs `do.ShutdownReport` errors at Debug via `logger.StdLogger` (slog.Default fallback), never masking command outcome. Tests: error case logs (requires `do.MustInvoke` first — do is lazy), clean case stays silent.
- **Deps pinned**: go-output `daghtml`, `tui`, `graph` all at v0.38.2 (exact pins per ADR-0002 rule 3).

## b) Partially done

- **P20**: code half done (above); the AGENTS.md memory entries (per-attempt hooks gotcha M88, go-output adoption pointer M89) are NOT yet written — that's the first item of the remaining work.
- **Race coverage**: execution + progress packages are `-race` green this session; the commands package ran `-race` once (191s) BEFORE the `--report`/`--graph`/`--tui` additions — needs one final re-run.
- **NO_COLOR record-keeping**: finding is in the manual matrix + ADR negative-consequence note; the actual upstream issues to go-output are drafted-in-head only, not filed.

## c) Not started

- **P21** HARVEST (plan → TODO_LIST.md/ROADMAP.md, dedupe, mark section (f) harvested)
- **P22** hygiene (hooks.go FreedBytes→SizeEstimate deprecation hints at hooks.go:49–50, container.go ireturn nolint-with-reason, retry.go mnd consts, M98 targeted `buildflow -s golangci-lint`, plus lint noise my new test files added: wsl_v5/makezero/varnamelen/golines warnings)
- **P23** commands print-helper migration (forbidigo/emoji cleanup)
- **P24** website GraphHTML embed
- **Final verification** (full short suite ~10 min, build, AGENTS.md update, git-log completeness check)

## d) What I fucked up (all caught and fixed; none outstanding)

The honest headline: **the W1 commits contained two real wiring bugs** — double `ActivityRegistered` emission (wrong ordering: before `WorkflowStarted` on the nom side) and scan steps with no start hooks. They shipped in the W1 wave and were only caught because P10 asserted exact event sequences. Nothing user-visible broke (the live smoke looked right), but "looks right" ≠ "is right"; this is exactly why the plan's test tasks exist, and why I refuse to call W1 "done" before P10 was green.

My own churn this session (all transient, all fixed in-place): a first-draft test invented a nonexistent API (`execution.ResolveForTest`) and deref'd a bool; reimplemented `filepath.Base/Dir` by hand; guessed `format.Bytes`/`format.Duration` output instead of reading them first ("1.5 kB"/"2s" vs actual "1.5 KiB"/"2.0 s"); used `_` as a parameter type in a fake; swapped `os.Stdout` from two parallel tests (real data race, caught by `-race`); duplicate `machineOutput` declaration during the graph wiring; a `do.Provide`d service never invoked (lazy DI → shutdown test silently no-oped until `do.MustInvoke`). Also unresolved-niggles: golangci counts 11 methods on `ProgressEmitter` (I count 10 — worth a look in P22) and my new test files carry fresh wsl_v5/varnamelen warnings.

## e) Improvements for the codebase (beyond the plan's tasks)

1. Two upstream candidates for go-output: InlineRenderer should honor NO_COLOR/TERM=dumb (colorprofile), and tui should expose cache-path/subscriber injection. Draft via verify-before-filing before touching their trackers.
2. `--report` could include scan mode/profile in its subtitle (currently only clean does).
3. Pipeline `--graph` could render unavailable cleaners as distinct (dashed) nodes — currently available-only.
4. `runCleanCommand` is at 16 positional params; a `cleanRunConfig` struct would stop the whack-a-mole of updating four call sites per new flag.
5. The `-update-golden` flag pattern is duplicated across report tests — one shared test helper would do.

## f) Next steps (ordered; first ≈10 are the plan's remainder)

1. P20/AGENTS.md: per-attempt hooks gotcha + go-output adoption pointer (ADR-0002, manual matrix) + Last Reviewed bump.
2. P22: hooks.go FreedBytes→SizeEstimate; container.go ireturn nolint-with-reason; retry.go mnd consts.
3. P22: `go mod tidy` via buildflow (promote daghtml/graph/tui from `// indirect` to direct) + targeted `buildflow -s golangci-lint` on all touched files (M98) — includes investigating the interfacebloat 11-vs-10 count.
4. P22: quiet the lint warnings my new test files introduced.
5. P23: print-helper migration in commands (inventory → helpers → clean.go/scan.go → remaining → lint).
6. P21: HARVEST into TODO_LIST.md/ROADMAP.md (tui promotion triggers, upstream issues, --report/--graph nice-to-haves) + dedupe + mark research section (f) harvested.
7. P24: website GraphHTML embed + build verify.
8. Final: `GOEXPERIMENT=jsonv2 go build ./...` + full short suite + one `-race` pass over commands/progress/report/di/execution.
9. Final: verify daemon captured everything (`git log --stat`), update AGENTS.md, close todos.
10. Post-plan (owner-gated): file the two go-output upstream issues; --tui pty smoke on a real run to firm up the spike verdict; g1 default-on flip decision after one stable release.

## g) Questions for the owner (3)

1. **g1 flip timing**: `--progress` is opt-in per ADR-0002. After this plan ships and survives one release, do you want the follow-up appendix to flip it default-on for interactive TTY runs (keep `--no-progress` escape), or keep opt-in until user feedback exists?
2. **bubbletea/v2 v2.0.10**: go.mod now carries bubbletea v2.0.10 (research notes said v2.0.9). Deliberate bump on your side, or daemon/transitive drift I should re-pin to match the fleet?
3. **Commit style for the remainder**: the daemon is committing everything as `chore: auto-commit N file(s)`. For P21–P24 + final verification, keep that, or should I make explicit conventional commits per task (`feat(report): ...`, `docs(adr): ...`) so the wave boundaries are readable in history?
