# TODO List

> Short- and mid-term improvement tasks for go-humanize-linter.
>
> **Open work only.** When a task is finished, delete it here and record it in
> `CHANGELOG.md`. Every item below has been verified against the code as not-yet-done.
>
> Legend — Effort: `XS` ≤15 min · `S` ≤1 h · `M` ≤4 h · `L` ≥½ day.

## Open tasks

| #   | Task                                                                                     | Tier   | Effort | Status                                                                                        | Evidence                                                |
| --- | ---------------------------------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| T34 | CI guard: fail loud when go.mod `go` directive leaves the allowed range (policy in code) | High   | XS     | open — bumper struck 4×; incident #4 closed by adoption in v0.5.0, guard still missing        | `ci.yml` (no guard step); AGENTS.md gotcha              |
| T35 | `BenchmarkH012Detector` mirroring the H010 benchmark shape                               | Medium | S      | open — H010 has one, H012 shipped without                                                     | `bench_test.go:37` (H010 precedent)                     |
| T36 | Corpus re-sweep under gogenfilter v3.6.1 (detection table may have moved)                | High   | M      | open — pin landed 2026-09-19, not re-validated                                                | commit `9c76bd4`; sweep loop in docs/validation         |
| T37 | `scripts/post-release-check.sh` — bounded proxy/go-get/pkg.go.dev checks                 | Medium | M      | open — hand-polled every release so far                                                       | `docs/status/2026-09-19_12-08` NEXT#4                   |
| T38 | Depth-2 nested-repo sweep support (`games/KeyCountdown`-style layouts)                   | Medium | M      | open — corpus sweeps miss nested modules                                                      | `docs/validation/2026-09-18_h010-sweep.md` (scope note) |
| T39 | Multi-platform release artifacts (darwin/windows/arm64; today: linux-amd64 only)         | Medium | M      | open — carried since v0.2.0 planning                                                          | `.github/workflows/release.yml`; 09-02/09-09 reports    |
| T40 | Weekly self-scan + corpus-sweep ritual (CI schedule or checklist)                        | Medium | S      | open — proposed 3×, never started; red-main drifted 5 days once                               | `docs/status/2026-09-18_22-01` NEXT#29                  |
| T29 | H011 rule: manual SI-string parsing (ParseSI)                                            | Medium | M      | deferred — zero corpus demand (2026-09-18 research); revisit on next upstream go-humanize tag | `rules.go` (H011 reserved); ROADMAP H011+               |

## Standing decisions (not tasks — do not re-litigate without the user)

- Declined 2026-09-19: T18 plugin-index PR, T20 gogenfilter release, branch
  protection on main, and any foreign-repo actions. Consumer repos
  (KeyCountdown, Kernovia, CreditReformBilanzampel) stay unfixed as the
  demonstration corpus.
- Go directive policy (2026-09-19): revert-on-sight for background bumps —
  incident #4 (`b7f00c2`) violated it; closed 2026-10-23 by ADOPTING Go 1.27
  and repinning ci.yml/release.yml/action.yml together (v0.5.0). The policy
  stands for future bumps; the T34 CI guard is still open.
- Baseline flags: upgrading to a release that adds rules (e.g. H010 in
  v0.3.0) adds findings — regenerate `--behavior-delta` baselines once after
  upgrading (documented in README).
