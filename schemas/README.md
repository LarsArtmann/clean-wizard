# JSON Schema for YAML Configuration

This directory contains JSON Schema definitions for validating clean-wizard YAML configuration files.

## config.schema.json

Comprehensive schema for validating clean-wizard YAML configurations with full support for type-safe enum fields.

### Features

- **Type-safe enum validation**: All enum fields support both string and integer representations
- **Comprehensive coverage**: Validates all configuration structures, nested objects, and array types
- **Descriptive error messages**: Each enum value includes human-readable descriptions
- **ISO 8601 timestamp validation**: Proper date-time format checking for timestamps
- **Duration pattern validation**: Validates custom duration formats (e.g., `7d`, `24h`, `30m`)

### Supported Enum Types

#### Binary Enums (0/1 values)
- `safe_mode`: DISABLED (0), ENABLED (1), STRICT (2)
- `enabled` (profile/operation): DISABLED (0), ENABLED (1)
- `optimize` (nix_generations): DISABLED (0), ENABLED (1)
- `cache_cleanup_mode`: DISABLED (0), ENABLED (1)

#### Multi-Value Enums
- `risk_level`: LOW (0), MEDIUM (1), HIGH (2), CRITICAL (3)
- `execution_mode` / `dry_run`: DRY_RUN (0), NORMAL (1), FORCE (2)
- `package_managers`: NPM (0), PNPM (1), YARN (2), BUN (3)
- `docker_prune_mode`: ALL (0), IMAGES (1), CONTAINERS (2), VOLUMES (3), BUILDS (4)
- `build_tool_type`: GO (0), RUST (1), NODE (2), PYTHON (3), JAVA (4), SCALA (5)
- `cache_type`: SPOTLIGHT (0), XCODE (1), COCOAPODS (2), HOMEBREW (3), PIP (4), NPM (5), YARN (6), CCACHE (7)
- `version_manager_type`: NVM (0), PYENV (1), GVM (2), RBENV (3), SDKMAN (4), JENV (5)
- `homebrew_mode`: ALL (0), UNUSED_ONLY (1)

### Usage with Validation Tools

#### Using `ajv` (Node.js)

```bash
npm install -g ajv-cli
ajv validate -s schemas/config.schema.json -d path/to/config.yaml
```

#### Using `jsonschema` (Python)

```bash
pip install jsonschema
python -m jsonschema -i config.yaml schemas/config.schema.json
```

#### Using `yamllint`

Add to your `.yamllint` configuration:

```yaml
extends: default
rules:
  truthy:
    allowed-values: ['true', 'false']
```

Then combine with JSON schema validation:

```bash
yamllint config.yaml && \
  python -m jsonschema -i config.yaml schemas/config.schema.json
```

### Integration with Editors

#### VS Code

Add to `.vscode/settings.json`:

```json
{
  "yaml.validate": true,
  "yaml.schemas": {
    "./schemas/config.schema.json": ["*.yaml", "*.yml"]
  },
  "yaml.customTags": []
}
```

#### IntelliJ IDEA / GoLand

1. Open Settings/Preferences
2. Go to Languages & Frameworks → Schemas and DTDs
3. Add JSON Schema: point to `schemas/config.schema.json`
4. Associate with `.yaml` and `.yml` files

### Continuous Integration

#### GitHub Actions

```yaml
name: Validate Configs

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install jsonschema
        run: pip install jsonschema pyyaml
      - name: Validate YAML configs
        run: |
          for file in *.yaml; do
            echo "Validating $file..."
            python -m jsonschema -i "$file" schemas/config.schema.json
          done
```

#### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
set -e

# Validate all YAML files
for file in $(git diff --cached --name-only --diff-filter=ACM | grep '\.yaml$'); do
  python -m jsonschema -i "$file" schemas/config.schema.json
done
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

### Schema Versioning

This schema follows semantic versioning matching the clean-wizard configuration version:
- Major version: Breaking changes to schema structure
- Minor version: New enum values or optional fields
- Patch version: Bug fixes, documentation updates

### Contributing

When adding new enum types or configuration fields:

1. Update the enum definitions in `internal/domain/*.go`
2. Update this schema with the new enum values and constraints
3. Update this README with documentation for new fields
4. Add test cases for schema validation
5. Update version if breaking changes were made

### Testing Schema Validity

Validate the schema itself:

```bash
npm install -g ajv-cli
ajv compile -s schemas/config.schema.json -r draft-07
```

### Common Validation Errors

#### Invalid Enum Value
```
Error: 0 is not a valid enum value for field "package_managers[0]"
Expected: one of NPM (0), PNPM (1), YARN (2), BUN (3)
```

**Solution**: Use valid enum integer or string representation.

#### Invalid Duration Format
```
Error: "7days" does not match pattern "^\\d+[dhms]$"
```

**Solution**: Use proper duration format: `7d` (not `7days`), `24h` (not `24hours`), `30m` (not `30minutes`).

#### Missing Required Field
```
Error: missing required property "profiles"
```

**Solution**: Add the required field to your configuration.

### Related Documentation

- [Configuration Guide](../../docs/config.md)
- [Domain Models](../../docs/domain.md)
- [Architecture Documentation](../../docs/architecture/)
