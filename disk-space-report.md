# Disk Space Report - Clean Wizard

**Generated:** 2026-02-22
**Total Project Size:** ~72MB

## Summary

The project is relatively small, but there are opportunities to reclaim ~60MB+ by cleaning git history.

---

## Disk Usage Breakdown

| Directory/File  | Size  | Status                           |
| --------------- | ----- | -------------------------------- |
| `.git/`         | 65MB  | ⚠️ Bloated (binaries in history) |
| `reports/`      | 2.6MB | ⚠️ Generated file                |
| `docs/`         | 2.1MB | OK                               |
| `internal/`     | 1.2MB | OK                               |
| `cmd/`          | 88KB  | OK                               |
| Everything else | <1MB  | OK                               |

---

## Issues & Recommendations

### 1. Git History Bloat (64MB) - HIGH IMPACT

**Problem:** Binaries were committed to git history and later removed. They still occupy space in `.git/objects`.

**Largest files in history:**

| File                                | Size        | Type      |
| ----------------------------------- | ----------- | --------- |
| `clean-wizard` (multiple versions)  | ~8.7MB each | Binary    |
| `bin/clean-wizard`                  | 8.0MB       | Binary    |
| `security-test`                     | 7.2MB       | Binary    |
| `simple-test-runner`                | 5.5MB       | Binary    |
| `.reports/html/styles/tailwind.css` | 2.9MB       | Generated |

**Recommendation:**

1. Add to `.gitignore`:
   ```
   clean-wizard
   bin/clean-wizard
   security-test
   simple-test-runner
   *.test
   ```
2. Rewrite history with `git filter-repo` to remove binaries (saves ~50MB+)
3. Force push (coordinate with team)

**Command to clean:**

```bash
# Install git-filter-repo first
git filter-repo --path clean-wizard --invert-paths
git filter-repo --path bin/clean-wizard --invert-paths
git filter-repo --path security-test --invert-paths
git filter-repo --path simple-test-runner --invert-paths
```

---

### 2. Generated Reports (2.6MB) - LOW IMPACT

**Problem:** `reports/current-dupl-analysis.txt` is a generated file.

**Recommendation:** Add to `.gitignore`:

```
reports/*.txt
reports/*.html
```

---

### 3. DS_Store Files (88KB) - LOW IMPACT

**Found 10 `.DS_Store` files:**

- `./.DS_Store` (12KB)
- `./internal/.DS_Store` (12KB)
- `./docs/.DS_Store` (12KB)
- `./cmd/.DS_Store` (8KB)
- `./test/.DS_Store` (8KB)
- `./tests/.DS_Store` (8KB)
- `./.git/.DS_Store` (8KB)
- `./.reports/.DS_Store` (8KB)

**Recommendation:** Already should be in `.gitignore`. Clean with:

```bash
find . -name ".DS_Store" -delete
```

---

### 4. Historical Status Reports (1.2MB) - OPTIONAL

**Location:** `docs/status/`

Contains 20+ status reports from past development sessions. Oldest from 2025-11-17.

**Options:**

- Keep as historical record
- Archive older reports (>3 months) to separate location
- Delete if no longer needed

---

## Quick Wins

| Action                   | Space Saved    | Effort |
| ------------------------ | -------------- | ------ |
| Delete `.DS_Store` files | 88KB           | 1 min  |
| Git history cleanup      | 50MB+          | 15 min |
| Ignore generated reports | Future savings | 1 min  |

---

## Recommended `.gitignore` Additions

```gitignore
# Binaries
clean-wizard
bin/clean-wizard
security-test
simple-test-runner
*.test
*.exe

# Generated reports
reports/*.txt
reports/*.html

# macOS
.DS_Store

# Temp files
*.tmp
*.log
```

---

## Conclusion

**Immediate action:** Clean git history to reclaim ~50MB+ (70%+ of total size).

The project itself is well-organized with no excessive dependencies or build artifacts. The main issue is historical binary commits that inflated the git repository.
