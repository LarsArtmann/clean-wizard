# Clean Wizard - Project Instructions

## Build & Test

```bash
go build ./...
go test ./... -short
```

## Project Structure

- `cmd/clean-wizard/` - CLI entry point and commands
- `internal/cleaner/` - Cleaner implementations (13 cleaners)
- `internal/domain/` - Domain types and enums
- `internal/config/` - Configuration loading and validation
- `docs/` - Documentation

## Key Files

- `TODO_LIST.md` - Current source of truth for pending work
- `FEATURES.md` - Feature status documentation
