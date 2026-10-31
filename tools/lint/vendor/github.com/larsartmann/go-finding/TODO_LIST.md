# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).
> Owner questions (ApplySimpleFixes philosophy, brew/nix, report policy)
> live in [ROADMAP.md](ROADMAP.md) "Open questions" — not here.

> **Rebuilt 2026-09-23 (evening)** after the whole-list execution round: the
> v1.13.0 train was COMPLETED (all 5 tags + GitHub Releases, Latest = v1.13.0,
> proxy + consumer-compat verified), issue #36 closed, the release-integrity
> gates landed (changelog-drift, README version sweep, CI nix job,
> consumer-compat job, one-tag-per-push procedure, 5-tag preflight), the
> testing bundle landed (99% root coverage, fuzz campaigns, ordering/conflict
> pins), GOEXPERIMENT=jsonv2 was dropped (json/v2 GA in Go 1.27), and the four
> planned consumer filings resolved without filing (two claims verified stale,
> one fixed directly in go-business-rules 431d8cd, art-dupl GAP-2 already
> filed as #1). Execution record: the annotation log in
> `docs/planning/2026-09-23_15-01_SUPERB-pareto-plan-release-truth-and-gates.md`.
>
> **Updated 2026-09-27:** dependabot train resolved (#37 merged; #38's go.work
> floor + vendorHash fixed and merged — all 5 modules now on the bumped deps
> with `go 1.27.1`; #39 superseded), the `.goreleaser.yml` nix-pipe
> `{{binary}}` template bug is FIXED (renders as `{{ .Binary }}`; drop
> `--skip=nix` from the local procedure after the next successful render),
> consumer-compat now `go vet`s every resolved library module, the fuzz
> campaigns' corpus entries are committed as seeds (145 across
> FuzzEditListProvider/FuzzTextEditValidate/FuzzParseLSPDiagnosticTags),
> pkg.go.dev v1.13.0 pages render-verified (core + pipeline), the sampling
> NO-GO backstop review date is set (2027-01-31), the stale
> hierarchical-errors clone question resolved itself (directory gone), and
> pre-push-verify gained a `/mnt/buildcache` df guard.

---

## 🔴 HIGH Priority — next train (v1.14.0) & gates that need a tag

| Task                                                                                                    | Impact | Effort | Notes                                                                                                                                                                                                                                                      |
| ------------------------------------------------------------------------------------------------------- | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Cut the v1.14.0 train for the multi-edit feature (Edits, EditListProvider, ResolveFlightRecorderConfig) | Crit   | L      | DONE 2026-10-11: pipeline/analysis tagged at 9c675c7 (pre-DetectorV2, consumer-proven against published core v1.14.0), cmd tagged at the resync commit 13cc091 so `go install` ships the new engine; master CI 20/20 green for the first time since Oct 5. |
| CLI: delegate `flightRecorderFileConfig.resolve()` to `pipeline.ResolveFlightRecorderConfig`            | Med    | S      | Prepared design in place; UNBLOCKED — pipeline/v1.14.0 tagged 2026-10-11. Kills the duration-parse split brain; wording family then unified via the pinned pipeline messages.                                                                              |
| Restore sigstore signing on the next workflow-run release train                                         | Med    | S      | v1.13.0 shipped WITHOUT `.sigstore.json` bundles (recovery release ran GoReleaser locally where cosign keyless has no OIDC). Checksums.txt is unsigned too.                                                                                                |
| erraudit T13/T14 migration onto tagged analysis/pipeline APIs 🔒(after v1.14.0)                         | High   | M      | Blocked on the Edits bridge being TAGGED (not in v1.13.0). Also file the upstream erraudit issue: `nolint-audit` false-reports type-aware directives as stale (verified 2026-09-23, 24/24).                                                                |
| Build `scripts/release-train.sh` + post-release verification (T11)                                      | Med    | L      | The prep→verify→tag→serialized-push→rerun→resync→postflight sequence ran manually three times now; encode it (proxy ×5, clean-dir `go get` ×5, assets, pkg.go.dev backoff).                                                                                |

## toolsdk/v1.16.0 train — fixes on master since 10429a5 (updated 2026-10-11)

| Task                                                                                                               | Impact | Effort | Notes                                                                                                                                                                                                                                                                  |
| ------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Push master (8+ commits: #44/#43/#42 fixes, `Spec.Outputs`, dryrun tests, docs)                                    | Crit   | S      | Nothing is consumable until pushed; verify a ci.yml run exists for the pushed SHA (one-push-one-event gotcha).                                                                                                                                                         |
| Verify downstream spec compatibility BEFORE tagging — DONE 2026-10-11                                              | —      | —      | All 12 blank-imported BuildFlow specs pass the new `Trigger.Validate` (resolved from go-finding master via BuildFlow `tools/go.mod` replace). One dirty spec found and fixed in its own repo: oxlint-auto-configure `Language: "javascript"` → `"js"` (startup panic). |
| Cut toolsdk/v1.16.0: release-preflight `--bench --stress`, stress `ginkgo -r --race --repeat=20`, one-tag-one-push | Crit   | M      | Carries: Trigger validation (#44), typed accessors (#43), `RepairResult.ModifiedFiles`/`Facts` (#42), `Spec.Outputs` (write contract, consumed by BuildFlow's DAG edge derivation), `DryRunFromContext` nil-context fix, round-trip fidelity + dryrun tests.           |
| Correct the "11 specs" figure issue #44 carries (actual: 12)                                                       | Low    | S      | Re-derived from BuildFlow `tools/providers/sdk_imports.go` on 2026-10-11; correct in the close comment or release notes.                                                                                                                                               |
| Post-tag: proxy `go get` verification + pkg.go.dev render check for toolsdk/v1.16.0                                | Med    | S      | Standard train tail; consumer bumps (BuildFlow drops its `../go-finding/toolsdk` replace + deletes the `legacyWriteContracts` shim once go-auto-upgrade declares `Outputs`).                                                                                           |
| Add toolsdk to preflight's GOWORK=off build/tidy loops (script skips it today)                                     | Low    | S      | AGENTS.md documents the manual workaround; fix the script instead.                                                                                                                                                                                                     |

## 🟠 HIGH-MED Priority — follow-ups & ecosystem

| Task                                                       | Impact | Effort | Notes                                                                                                                                                                      |
| ---------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Opportunistic consumer bumps beyond the leads              | Low    | Med    | go-linter-sdk + golangci-lint-auto-configure are on v1.13.0 (done); the remaining v1.8.0-era rows bump per consumer. gomend + licenseforge stay blocked upstream (#1/#46). |
| Watch erraudit upstream for the nolint-audit staleness fix | Low    | Low    | Once fixed, the advisory listing in `scripts/error-audit.sh --nolint-audit` becomes usable as a removal oracle again.                                                      |
| Re-check library-policy after hook issue #74 resolves      | Low    | Low    | #74 still OPEN (verified 2026-09-23).                                                                                                                                      |
| Diet AGENTS.md under 220 lines (go-structure-linter gate)  | Med    | Med    | 283 lines vs max 220 (2026-10-11): move bulk (e.g. the Key Files table) to docs/ and keep pointers. Each cut needs a memory-loss check — curated context, not filler.      |

## 🟡 MEDIUM Priority — testing & code quality

| Task                                                                                 | Impact | Effort | Notes                                                                                                                                                                                           |
| ------------------------------------------------------------------------------------ | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Dependabot watch: gomega bumps merged; pin-vs-dependabot decision for actions/* SHAs | Low    | Low    | #25 merged; sbom bump was manual review.                                                                                                                                                        |
| Periodic `/mnt/buildcache` hygiene                                                   | Low    | Low    | The shared GOCACHE hit 100% (167G) and broke every build; `go clean -cache` recovered. A df guard (<5 GiB = FAIL) now runs in pre-push-verify (2026-09-27); a periodic prune/cron remains open. |

## 🟢 LOW Priority — launch & product

| Task                                                                      | Impact | Effort | Notes                                                                                                                                 |
| ------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| Submit Awesome Go entry (LAUNCH3)                                         | Med    | Low    | Entry text ready in `docs/brainstorming/launch-announcement-draft.md` (finalized 2026-09-23); needs an awesome-go fork + PR. 🔒 owner |
| Create `LarsArtmann/homebrew-tap` + `HOMEBREW_TAP_GITHUB_TOKEN` (LAUNCH4) | Med    | S      | goreleaser brews section renders but ships `skip_upload: true` until the tap repo + secret exist. ROADMAP open question. 🔒 owner     |
| CONTRIBUTING / release-runbook page for public audience                   | Low    | Med    | Release procedure is internal-facing.                                                                                                 |
| v2.0 design spike session                                                 | Low    | High   | Parked in [ROADMAP.md](ROADMAP.md) "Hardening": Position sentinel, FixStrategy union, TagSet, sub-structs.                            |
| Sampling NO-GO review date (e.g. 2027-01 calendar entry)                  | Low    | Low    | DONE 2026-09-27: backstop date 2027-01-31 set in ROADMAP.                                                                             |

---

_Assisted-by: Crush <crush@charm.land>_
