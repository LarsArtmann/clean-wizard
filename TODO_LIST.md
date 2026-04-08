# TODO LIST

**Last Updated:** 2026-04-05
**Focus:** Actionable items for the next 2-4 weeks

---

## Pending Tasks

| #   | Task                                         | Impact | Effort | Notes                              |
| --- | -------------------------------------------- | ------ | ------ | ---------------------------------- |
| 1   | Add tests for getRegistryName reverse lookup | MED    | LOW    | Related to metadata consolidation  |
| 2   | Add profile command tests                    | MED    | MED    | No test files for commands package |
| 3   | Add scan command tests                       | MED    | MED    | No test files for commands package |
| 4   | Add clean command tests                      | MED    | HIGH   | No test files for commands package |
| 5   | Set up CI pipeline                           | HIGH   | MED    | At minimum: go build + go test     |
| 6   | Fix pre-commit hook timeout                  | MED    | LOW    | golangci-lint times out            |

---

## Quick Actions

- Run build: `go build ./...`
- Run tests: `go test ./... -short`
- Lint: `golangci-lint run`

---

**Status:** 6 actionable items pending
