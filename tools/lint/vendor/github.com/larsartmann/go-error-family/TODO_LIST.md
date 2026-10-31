# TODO List

Short- and mid-term actionable improvement tasks. Each item is bounded and
traceable to its source. When an item ships, remove it here and record it in
`CHANGELOG.md` under the version it shipped in. Harvest routing policy
(decision 2026-09-30): aggressive one-pass routing of every bounded item;
the next docs-health pass prunes or promotes.

**Last updated:** 2026-09-30 (HARVEST of `docs/status/2026-09-30_06-08_art-dupl-deep-sweep-triage.md` §f)

---

## 🟡 Security (upstream-pinned advisories)

- [ ] GHSA-rj75-hqrm-r3gf (postcss-selector-parser <7.1.6): upstream-pinned via @expressive-code/core 0.44.2 (Starlight chain); warning-severity build-time CPU advisory, gate stays green. Revisit when expressive-code bumps past 6.1.4 (2026-10-07: pnpm-audit unskipped after BuildFlow's lockfileDir fix; semver sweep + source-map-js override cleared the rest).

## Active

1. **Announce the Bridge Patterns guide** (source: ROADMAP theme 4) —
   REOPENED 2026-09-29: the GitHub Discussion venue was tried and withdrawn
   (a link-post in an otherwise empty Discussions tab serves no purpose;
   Discussions disabled again). The draft
   (`docs/planning/2026-09-28_bridge-patterns-announcement-draft.md`) is
   voice-checker-clean and self-reviewed; pick a channel with real audience
   (r/golang candidate) before publishing. Bounded: one channel decision +
   one publish.
2. **Heal the v0.11.0-tagged README on pkg.go.dev** (source: 2026-09-29
   docs-health pass) — the tag was cut before the announcement-withdrawal
   commit (`986a6a6`), so the README rendered at `@v0.11.0` still links the
   deleted Discussion #12. Master is already clean; any next release heals
   it. Bounded: decide standalone v0.11.1 vs ride-along on the next release.
3. **Review the post-v0.11.0 Dependabot wave** (source: 2026-09-29 report
   §f20) — submodule root-pin bumps (agent, bridge, diagnose, diagnose/git,
   diagnose/postgres, examples → v0.11.0) arrive on the weekly schedule;
   merge the green ones or hold for the next release. Bounded: one triage
   pass.
4. **Enforce the art-dupl baseline in CI** (source: 2026-09-28 Pareto report
   §f5; verification 2026-09-30 report §f6) — `.art-dupl-baseline.json` is
   committed but nothing runs `art-dupl check -t 1 .` in CI; a baseline nobody
   executes is prose, not a gate. 2026-09-30 verification: gate proven both
   ways (0 clean / 2 planted clone / 1 unknown-flag); `check` is plain-mode
   ONLY, so the CI job must run the bare command. Decision 2026-09-30 (user):
   CI enforces the baseline gate — `-t 1`, not the routine `-t 5`. Bounded:
   one `ci.yml` job; if it lands, add the planted-clone negative-path fixture
   as a check (2026-09-30 report §f28).
5. **Mechanically diff deep-sweep hashes against the baseline** (source:
   2026-09-30 report §f2) — the 2026-09-30 triage matched the 3 reported
   groups by file pair + code read and the hash-based gate was green, but
   scan-vs-baseline hashes were never compared programmatically. Bounded:
   one scan run + hash diff. Effort S.
6. **Decide and file the art-dupl flags-provenance upstream ask** (source:
   2026-09-30 report §f3/f27) — the baseline serves the plain gate only
   (decided 2026-09-30), but the JSON records no flags and `check` rejects
   `--suggest-generics` (exit 1). Two upstream questions: a `flags`
   provenance field, and `--type-aware` parity for `check`. Bounded:
   verify-before-filing pass + one issue. Effort M.
7. **Triage the gopls nilness warning at `error_test.go:567:44`** (source:
   2026-09-30 report §f5) — gopls reports `[nilness][nilpanic] panic with nil
   value`. Verdict needed: deliberate nil-panic test vs real smell; fix or
   suppress with rationale. Bounded: one file, one verdict. Effort S.
8. **Verify `art-dupl check` honors the scan's exclusion defaults** (source:
   2026-09-30 report §f8) — generated/test exclusions are auto-applied in
   scan mode; parity in check mode is unverified. Bounded: one repro
   attempt. Effort S.
9. **Spot-audit the `-t 1` suppression classes** (source: 2026-09-30 report
   §f9) — the deep sweep reports "221 non-actionable, 96 filtered suppressed"
   without enumerating them; confirm nothing actionable hides there
   (one-time audit; repeat only if the filter changes). Effort M.
10. **Define the deliberate re-baseline trigger** (source: 2026-09-30 report
    §f10/f29) — baseline `recordedAt` is 2026-09-28; document in AGENTS.md
    when to re-run `art-dupl baseline` (cross-module refactor, post-release
    drift check) so re-baselining is never silent. Bounded: one clause.
    Effort S.
11. **Package the pnpm-audit skip as an upstream BuildFlow issue** (source:
    2026-09-30 report §f13; standing AGENTS.md Known Limitations) —
    subdirectory lockfile discovery repro; verify-before-filing gate
    applies. Effort M.
12. **Package the branching-flow `IsIgnored` skip as an upstream BuildFlow
    issue** (source: 2026-09-30 report §f14; standing AGENTS.md Known
    Limitations) — phantom analyzer ignores its own `IsIgnored`; repro +
    file. Effort M.
13. **Run go-structure-linter CLI directly and diff against project config**
    (source: 2026-09-30 report §f15) — BuildFlow's embedded snapshot ignores
    `.go-structure-linter.yaml`; run the CLI directly and confirm the `flat`
    preset intent still holds. Effort M.
14. **Add usage examples for the under-adopted boundary APIs** (source:
    2026-09-30 report §f22; adoption reality audited 2026-07-23) — `LogError`
    ~3, `HTTPHandler` ~5, `errorfamilytest` ~3 external consumers; README-
    level examples for each. Docs-only — the bigger grow-vs-freeze decision
    stays gated by ROADMAP Open Question #3. Effort M.
15. **Define a fuzzing cadence for the 16 fuzz targets** (source: 2026-09-30
    report §f23) — targets exist (root 11, bridge 5) but no schedule;
    decide nightly job vs pre-release run and wire it. Bounded: one decision
    - wiring. Effort M.
