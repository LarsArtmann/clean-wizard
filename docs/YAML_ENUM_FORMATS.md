# YAML Enum Format Guide

## Overview

Clean Wizard uses **type-safe enums** for all configuration fields that have a fixed set of values. These enums can be specified in YAML files using **either integers or strings**, providing flexibility and backward compatibility.

## Supported Formats

### 1. Integer Format (Recommended)

**Preferred format** - more compact, type-safe, and efficient:

```yaml
version: "1.0.0"
safe_mode: 1 # 0=DISABLED, 1=ENABLED, 2=STRICT
max_disk_usage: 50
protected:
  - "/System"
  - "/Library"
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup"
    enabled: 1 # 0=DISABLED, 1=ENABLED
    operations:
      - name: "nix-generations"
        description: "Clean Nix generations"
        risk_level: 0 # 0=LOW, 1=MEDIUM, 2=HIGH, 3=CRITICAL
        enabled: 1
        settings:
          nix_generations:
            generations: 1
            optimize: 0 # 0=DISABLED, 1=ENABLED
            dry_run: 1 # 0=DRY_RUN, 1=NORMAL, 2=FORCE
      - name: "docker"
        description: "Clean Docker resources"
        risk_level: 1
        enabled: 1
        settings:
          docker:
            prune_mode: 0 # 0=ALL, 1=IMAGES, 2=CONTAINERS, 3=VOLUMES, 4=BUILDS
```

**Benefits of integer format:**

- ✅ Compact configuration files
- ✅ Type-safe (no typos)
- ✅ Faster parsing
- ✅ Clear ordinal values
- ✅ Machine-readable

### 2. String Format (Human-Readable)

**Alternative format** - more readable and self-documenting:

```yaml
version: "1.0.0"
safe_mode: "ENABLED" # Can also use "enabled" (case-insensitive)
max_disk_usage: 50
protected:
  - "/System"
  - "/Library"
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup"
    enabled: "ENABLED" # "DISABLED" or "ENABLED"
    operations:
      - name: "nix-generations"
        description: "Clean Nix generations"
        risk_level: "LOW" # "LOW", "MEDIUM", "HIGH", "CRITICAL"
        enabled: "ENABLED"
        settings:
          nix_generations:
            generations: 1
            optimize: "DISABLED" # "DISABLED" or "ENABLED"
            dry_run: "NORMAL" # "DRY_RUN", "NORMAL", "FORCE"
      - name: "docker"
        description: "Clean Docker resources"
        risk_level: "MEDIUM"
        enabled: "ENABLED"
        settings:
          docker:
            prune_mode: "ALL" # "ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS"
```

**Benefits of string format:**

- ✅ Self-documenting configuration
- ✅ Human-readable
- ✅ Easier to understand at a glance
- ✅ Less likely to mix up similar values

### 3. Mixed Format (Valid but Not Recommended)

You can mix integer and string formats, but **this is discouraged**:

```yaml
safe_mode: 1 # Integer
risk_level: "LOW" # String
prune_mode: 0 # Integer
enabled: "ENABLED" # String
```

**Recommendation:** Choose **one format** (preferably integers) and use it consistently throughout your configuration.

## Complete Enum Reference

### Binary Enums (0 or 1)

| Enum Type          | Value 0  | Value 1 | Field Usage                                           |
| ------------------ | -------- | ------- | ----------------------------------------------------- |
| `CacheCleanupMode` | DISABLED | ENABLED | `go_packages.clean_cache`, `cargo_packages.autoclean` |
| `ProfileStatus`    | DISABLED | ENABLED | `profile.enabled`, `operation.enabled`                |
| `OptimizationMode` | DISABLED | ENABLED | `nix_generations.optimize`                            |

### SafeMode (0, 1, 2)

| Value    | Integer | String     | Description                |
| -------- | ------- | ---------- | -------------------------- |
| DISABLED | 0       | "DISABLED" | No safety checks           |
| ENABLED  | 1       | "ENABLED"  | Standard safety checks     |
| STRICT   | 2       | "STRICT"   | Maximum safety enforcement |

### RiskLevel (0, 1, 2, 3)

| Value    | Integer | String     | Description              |
| -------- | ------- | ---------- | ------------------------ |
| LOW      | 0       | "LOW"      | Minimal risk operations  |
| MEDIUM   | 1       | "MEDIUM"   | Moderate risk operations |
| HIGH     | 2       | "HIGH"     | High risk operations     |
| CRITICAL | 3       | "CRITICAL" | Critical risk operations |

### ExecutionMode / DryRun (0, 1, 2)

| Value   | Integer | String    | Description                     |
| ------- | ------- | --------- | ------------------------------- |
| DRY_RUN | 0       | "DRY_RUN" | Simulate without actual changes |
| NORMAL  | 1       | "NORMAL"  | Normal execution                |
| FORCE   | 2       | "FORCE"   | Force execution without prompts |

### PackageManagerType (0, 1, 2, 3)

| Value | Integer | String | Description          |
| ----- | ------- | ------ | -------------------- |
| NPM   | 0       | "NPM"  | npm package manager  |
| PNPM  | 1       | "PNPM" | pnpm package manager |
| YARN  | 2       | "YARN" | Yarn package manager |
| BUN   | 3       | "BUN"  | Bun package manager  |

### DockerPruneMode (0, 1, 2, 3, 4)

| Value      | Integer | String       | Description                   |
| ---------- | ------- | ------------ | ----------------------------- |
| ALL        | 0       | "ALL"        | Prune all Docker resources    |
| IMAGES     | 1       | "IMAGES"     | Prune only Docker images      |
| CONTAINERS | 2       | "CONTAINERS" | Prune only Docker containers  |
| VOLUMES    | 3       | "VOLUMES"    | Prune only Docker volumes     |
| BUILDS     | 4       | "BUILDS"     | Prune only Docker build cache |

### BuildToolType (0, 1, 2, 3, 4, 5)

| Value  | Integer | String   | Description                      |
| ------ | ------- | -------- | -------------------------------- |
| GO     | 0       | "GO"     | Go build tools                   |
| RUST   | 1       | "RUST"   | Rust/Cargo build tools           |
| NODE   | 2       | "NODE"   | Node.js build tools              |
| PYTHON | 3       | "PYTHON" | Python build tools               |
| JAVA   | 4       | "JAVA"   | Java build tools (Maven, Gradle) |
| SCALA  | 5       | "SCALA"  | Scala build tools (SBT)          |

### CacheType (0-7)

| Value     | Integer | String      | Description              |
| --------- | ------- | ----------- | ------------------------ |
| SPOTLIGHT | 0       | "SPOTLIGHT" | macOS Spotlight cache    |
| XCODE     | 1       | "XCODE"     | Xcode derived data cache |
| COCOAPODS | 2       | "COCOAPODS" | CocoaPods cache          |
| HOMEBREW  | 3       | "HOMEBREW"  | Homebrew cache           |
| PIP       | 4       | "PIP"       | Python pip cache         |
| NPM       | 5       | "NPM"       | Node.js npm cache        |
| YARN      | 6       | "YARN"      | Yarn cache               |
| CCACHE    | 7       | "CCACHE"    | ccache compiler cache    |

### VersionManagerType (0-5)

| Value  | Integer | String   | Description              |
| ------ | ------- | -------- | ------------------------ |
| NVM    | 0       | "NVM"    | Node Version Manager     |
| PYENV  | 1       | "PYENV"  | Python Version Manager   |
| GVM    | 2       | "GVM"    | Go Version Manager       |
| RBENV  | 3       | "RBENV"  | Ruby Version Manager     |
| SDKMAN | 4       | "SDKMAN" | SDKMAN for Java/Kotlin   |
| JENV   | 5       | "JENV"   | Java Environment Manager |

### HomebrewMode (0, 1)

| Value       | Integer | String        | Description                 |
| ----------- | ------- | ------------- | --------------------------- |
| ALL         | 0       | "ALL"         | Clean all Homebrew packages |
| UNUSED_ONLY | 1       | "UNUSED_ONLY" | Clean only unused packages  |

## Best Practices

### 1. Choose One Format and Stick to It

```yaml
# ✅ GOOD - Consistent integer format
safe_mode: 1
risk_level: 0
prune_mode: 0
enabled: 1

# ❌ AVOID - Inconsistent mixed format
safe_mode: 1
risk_level: "LOW"
prune_mode: 0
enabled: "ENABLED"
```

### 2. Use Integer Format for New Configurations

**Rationale:** Integer format is the recommended default because it:

- Takes less space in config files
- Prevents typos (you can't mistype `0`)
- Is unambiguous and machine-readable
- Aligns with internal representation

```yaml
# ✅ Recommended - Integer format
profiles:
  daily:
    operations:
      - name: "docker"
        settings:
          docker:
            prune_mode: 0 # Clear and unambiguous
```

### 3. Use String Format for Documentation or Teaching

**Rationale:** String format is excellent for examples and documentation because it:

- Self-documents the value
- Easier to understand without looking up references
- More approachable for new users

```yaml
# ✅ Good for documentation/examples
profiles:
  daily:
    operations:
      - name: "docker"
        settings:
          docker:
            prune_mode: "ALL" # Self-explanatory
```

### 4. Use Case-Insensitive String Matching

String format is case-insensitive, so all these are valid:

```yaml
prune_mode: "ALL"      # ✅ Uppercase (preferred)
prune_mode: "all"      # ✅ Lowercase (valid)
prune_mode: "All"      # ✅ Mixed case (valid)
prune_mode: "aLl"      # ✅ Valid but ugly (don't do this)
```

**Recommendation:** Use uppercase strings for consistency with enum definitions.

## Migration Guide

### From String to Integer Format

If you have existing string-format configs and want to migrate to integer format:

**Before (string format):**

```yaml
safe_mode: "ENABLED"
risk_level: "LOW"
prune_mode: "ALL"
```

**After (integer format):**

```yaml
safe_mode: 1
risk_level: 0
prune_mode: 0
```

**Quick Migration Tips:**

1. Use the enum reference tables above to map strings to integers
2. Test your migrated config with `clean-wizard scan --dry-run`
3. Commit the migration in a single change for easier review

### From Integer to String Format

If you prefer more human-readable configs:

**Before (integer format):**

```yaml
safe_mode: 1
risk_level: 0
prune_mode: 0
```

**After (string format):**

```yaml
safe_mode: "ENABLED"
risk_level: "LOW"
prune_mode: "ALL"
```

**Note:** Both formats work perfectly - choose based on your team's preference.

## Validation

Clean Wizard validates enum values at load time:

```yaml
# ❌ INVALID - Invalid enum string
risk_level: "EXTREME"
# Error: invalid risk level: EXTREME

# ❌ INVALID - Out-of-range integer
safe_mode: 99
# Error: invalid safe mode value: 99

# ❌ INVALID - Wrong type
enabled: true
# Error: cannot parse profile status: expected string or int
```

## Tools and Editor Support

### JSON Schema Validation

The project includes a JSON schema at `schemas/config.schema.json` that validates enum fields. Configure your editor to use it for real-time validation.

### VS Code

Add to `.vscode/settings.json`:

```json
{
  "yaml.validate": true,
  "yaml.schemas": {
    "./schemas/config.schema.json": ["*.yaml", "*.yml"]
  }
}
```

### IntelliJ IDEA / GoLand

1. Open Settings/Preferences
2. Go to Languages & Frameworks → Schemas and DTDs
3. Add JSON Schema: `schemas/config.schema.json`
4. Associate with `.yaml` and `.yml` files

## FAQ

### Q: Which format should I use?

**A:** Use integer format for production configs. It's more compact and type-safe. Use string format for examples and documentation.

### Q: Can I mix formats in the same file?

**A:** Yes, but it's not recommended. Choose one format and use it consistently.

### Q: Will my existing configs break?

**A:** No. Both formats are fully supported and will continue to work.

### Q: Which format is more performant?

**A:** Integer format is slightly faster for parsing (~5-10% faster), but the difference is negligible for typical config files.

### Q: Are there any plans to deprecate one format?

**A:** No. Both formats are first-class citizens and will be supported indefinitely.

### Q: How do I know which integer corresponds to which string?

**A:** Refer to the enum reference tables in this document, or check the source code in `internal/domain/operation_settings.go`.

### Q: Can I use lowercase strings?

**A:** Yes, string matching is case-insensitive. However, uppercase (`"LOW"`) is preferred for consistency.

## Related Documentation

- [Configuration Guide](CONFIGURATION.md)
- [Domain Models](domain.md)
- [JSON Schema](../schemas/config.schema.json)
- [Architecture Documentation](architecture/README.md)
