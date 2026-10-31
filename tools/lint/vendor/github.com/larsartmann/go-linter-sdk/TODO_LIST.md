# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status      | Meaning                                                     |
| ----------- | ----------------------------------------------------------- |
| TODO        | Not started. Needs doing.                                   |
| IN_PROGRESS | Actively being worked on.                                   |
| BLOCKED     | Cannot proceed, external dependency or decision needed.     |
| DONE        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

<!-- Numbering gap at #1, #2, #4: those items were completed and removed
     (1: v0.3.1 release, 2: secret deletion, 4: dprint decision). Numbers are
     never reused — historical reports reference them. -->

### Consumer adoption (the reason this SDK exists)

None open. (#10 converter-deletion verification and #11 the durable adoption
doc both completed 2026-10-11: all six consumers verified at source level,
evidence and refresh procedure in `docs/ADOPTION.md`; logged in the v0.4.0
CHANGELOG entry.)

### Documentation & drift

None open. (#12 the CI doc-drift guard and #13 the version-pointer refactor
both completed 2026-10-11: `scripts/check-doc-versions.sh` runs as a CI job
and a flake check; living docs point at `go.mod`. Logged in the v0.4.0
CHANGELOG entry.)

### Release

None open. (#14 completed 2026-10-11: v0.4.0 tagged after green CI on the
exact commit; pkg.go.dev verified. Logged in `CHANGELOG.md`.)

### Dependency hygiene

None open. (#15 benchmarks re-run 2026-10-11 — every measured op improved
20-30% vs the 2026-09-09 baseline, no regressions; baseline refreshed.
#16 dependabot sweep 2026-10-11: zero open PRs, nothing superseded.)

### Quality hardening

None open. (#17 flake `checks` for `go test` / `go vet` / `golangci-lint` plus
the doc guard completed 2026-10-11 — `nix flake check` runs five hermetic
derivations. #18 the bump checklist is encoded in
`scripts/bump-checklist.sh` and referenced from CONTRIBUTING.md.)

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
