# Enum Quick Reference

Quick reference guide for all type-safe enums in the clean-wizard project.

## Overview

All enums implement the `TypeSafeEnum[T]` interface:

```go
type TypeSafeEnum[T any] interface {
    String() string
    IsValid() bool
    Values() []T
}
```

Enums support both **YAML** and **JSON** serialization with case-insensitive string parsing.

---

## Core Enums

### RiskLevelType

**File:** `internal/domain/type_safe_enums.go`

Risk assessment for cleanup operations.

| Value | String     | Icon | Description                       |
| ----- | ---------- | ---- | --------------------------------- |
| 0     | `LOW`      | 🟢   | Safe to clean, minimal impact     |
| 1     | `MEDIUM`   | 🟡   | Moderate risk, review recommended |
| 2     | `HIGH`     | 🟠   | High risk, may affect projects    |
| 3     | `CRITICAL` | 🔴   | Critical, could cause data loss   |

**Usage:**

```go
risk := domain.RiskLevelMediumType
if risk.IsHigherThan(domain.RiskLevelLowType) {
    // Handle higher risk
}
fmt.Println(risk.Icon()) // Output: 🟡
```

**YAML:**

```yaml
risk_level: MEDIUM # or "medium", 1
```

**JSON:**

```json
{ "risk_level": "MEDIUM" }
```

---

### ValidationLevelType

**File:** `internal/domain/type_safe_enums.go`

Validation strictness levels.

| Value | String          | Description                            |
| ----- | --------------- | -------------------------------------- |
| 0     | `NONE`          | No validation                          |
| 1     | `BASIC`         | Basic validation checks                |
| 2     | `COMPREHENSIVE` | Thorough validation                    |
| 3     | `STRICT`        | Strict validation with detailed errors |

**YAML:**

```yaml
validation_level: STRICT # or "strict", 3
```

---

### ChangeOperationType

**File:** `internal/domain/type_safe_enums.go`

Types of changes tracked during operations.

| Value | String     | Description       |
| ----- | ---------- | ----------------- |
| 0     | `ADDED`    | Item was added    |
| 1     | `REMOVED`  | Item was removed  |
| 2     | `MODIFIED` | Item was modified |

---

### CleanStrategyType

**File:** `internal/domain/type_safe_enums.go`

Cleaning strategy modes.

| Value | String         | Icon | Description               |
| ----- | -------------- | ---- | ------------------------- |
| 0     | `aggressive`   | 🔥   | Delete aggressively       |
| 1     | `conservative` | 🛡️   | Delete conservatively     |
| 2     | `dry-run`      | 🔍   | Preview only, no deletion |

**Usage:**

```go
strategy := domain.StrategyDryRunType
if strategy == domain.StrategyDryRunType {
    // Only simulate, don't delete
}
```

**YAML:**

```yaml
strategy: dry-run # or "dryrun", "conservative", "aggressive", 0-2
```

---

### SizeEstimateStatusType

**File:** `internal/domain/type_safe_enums.go`

Status of size calculations.

| Value | String    | Description                    |
| ----- | --------- | ------------------------------ |
| 0     | `KNOWN`   | Size was calculated accurately |
| 1     | `UNKNOWN` | Size is an estimate            |

**Purpose:** Replaces boolean `Unknown` field, making impossible states unrepresentable.

---

## Operation Settings Enums

### CacheCleanupMode

**File:** `internal/domain/operation_settings.go`

Enable/disable cache cleanup.

| Value | String     | Description            |
| ----- | ---------- | ---------------------- |
| 0     | `DISABLED` | Cache cleanup disabled |
| 1     | `ENABLED`  | Cache cleanup enabled  |

**Usage:**

```go
mode := domain.CacheCleanupEnabled
if mode.IsEnabled() {
    // Perform cleanup
}
```

---

### DockerPruneMode

**File:** `internal/domain/operation_settings.go`

Docker resource pruning options.

| Value | String       | Description            |
| ----- | ------------ | ---------------------- |
| 0     | `ALL`        | Prune all resources    |
| 1     | `IMAGES`     | Prune images only      |
| 2     | `CONTAINERS` | Prune containers only  |
| 3     | `VOLUMES`    | Prune volumes only     |
| 4     | `BUILDS`     | Prune build cache only |

**YAML:**

```yaml
docker:
  prune_mode: ALL # or "IMAGES", "CONTAINERS", etc., 0-4
```

---

### BuildToolType

**File:** `internal/domain/operation_settings.go`

Supported build tool ecosystems.

| Value | String   | Description          |
| ----- | -------- | -------------------- |
| 0     | `GO`     | Go modules           |
| 1     | `RUST`   | Rust/Cargo           |
| 2     | `NODE`   | Node.js              |
| 3     | `PYTHON` | Python               |
| 4     | `JAVA`   | Java (Maven, Gradle) |
| 5     | `SCALA`  | Scala (SBT)          |

---

### CacheType

**File:** `internal/domain/operation_settings.go`

System cache types for cleanup.

| Value | String       | Platform | Description           |
| ----- | ------------ | -------- | --------------------- |
| 0     | `SPOTLIGHT`  | macOS    | Spotlight index cache |
| 1     | `XCODE`      | macOS    | Xcode derived data    |
| 2     | `COCOAPODS`  | macOS    | CocoaPods cache       |
| 3     | `HOMEBREW`   | macOS    | Homebrew cache        |
| 4     | `PIP`        | All      | Python pip cache      |
| 5     | `NPM`        | All      | Node.js npm cache     |
| 6     | `YARN`       | All      | Yarn cache            |
| 7     | `CCACHE`     | All      | Compiler cache        |
| 8     | `XDG_CACHE`  | Linux    | XDG cache directory   |
| 9     | `THUMBNAILS` | Linux    | Thumbnail cache       |

---

### PackageManagerType

**File:** `internal/domain/operation_settings.go`

Node.js package managers.

| Value | String | Description          |
| ----- | ------ | -------------------- |
| 0     | `NPM`  | npm registry         |
| 1     | `PNPM` | pnpm (efficient)     |
| 2     | `YARN` | Yarn package manager |
| 3     | `BUN`  | Bun runtime          |

---

## Serialization Formats

### YAML Format

All enums support:

- **Case-insensitive strings:** `LOW`, `low`, `Low` → all valid
- **Integer values:** `0`, `1`, `2` → mapped to enum values
- **Numeric strings:** `"0"`, `"1"` → converted to integers

**Example:**

```yaml
profile:
  risk_level: HIGH # String (case-insensitive)
  validation_level: 2 # Integer
  strategy: "dry-run" # String
```

### JSON Format

JSON serialization uses string representation only:

```json
{
  "risk_level": "HIGH",
  "validation_level": "COMPREHENSIVE",
  "strategy": "conservative"
}
```

---

## Common Patterns

### Validation

```go
func processRisk(risk domain.RiskLevelType) error {
    if !risk.IsValid() {
        return fmt.Errorf("invalid risk level: %d", risk)
    }
    // ... process
}
```

### Iterating Values

```go
for _, level := range domain.RiskLevelLowType.Values() {
    fmt.Printf("- %s (%s)\n", level.String(), level.Icon())
}
```

### Configuration Parsing

```go
// YAML automatically parsed via UnmarshalYAMLEnum helper
type Config struct {
    RiskLevel domain.RiskLevelType `yaml:"risk_level"`
    Strategy  domain.CleanStrategyType `yaml:"strategy"`
}
```

---

## Error Messages

Enums provide helpful error messages for invalid values:

```
invalid risk level value: INVALID

Valid options:
  Strings: LOW, MEDIUM, HIGH, CRITICAL
  Integers: 0, 1, 2, 3

See docs/YAML_ENUM_FORMATS.md for more details
```

---

## Quick Lookup Table

| Enum                   | File                  | Values Count |
| ---------------------- | --------------------- | ------------ |
| RiskLevelType          | type_safe_enums.go    | 4            |
| ValidationLevelType    | type_safe_enums.go    | 4            |
| ChangeOperationType    | type_safe_enums.go    | 3            |
| CleanStrategyType      | type_safe_enums.go    | 3            |
| SizeEstimateStatusType | type_safe_enums.go    | 2            |
| CacheCleanupMode       | operation_settings.go | 2            |
| DockerPruneMode        | operation_settings.go | 5            |
| BuildToolType          | operation_settings.go | 6            |
| CacheType              | operation_settings.go | 10           |
| PackageManagerType     | operation_settings.go | 4            |
