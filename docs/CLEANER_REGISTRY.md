# CleanerRegistry Documentation

The `Registry` provides thread-safe management of cleaner instances, enabling polymorphic operations over all registered cleaners.

## Overview

```go
type Registry struct {
    cleaners map[string]Cleaner
    mu       sync.RWMutex
}
```

The Registry is the central hub for managing all cleaner implementations. It provides:

- **Thread-safe access** via `sync.RWMutex`
- **Polymorphic operations** over all cleaners
- **Platform-aware filtering** via `IsAvailable()` checks

## Creating a Registry

```go
registry := cleaner.NewRegistry()
```

## Core Methods

### Register

Adds a cleaner to the registry. If a cleaner with the same name exists, it will be overwritten.

```go
registry.Register("docker", dockerCleaner)
registry.Register("npm", npmCleaner)
```

### Get

Retrieves a cleaner by name. Returns the cleaner and `true` if found, `nil` and `false` otherwise.

```go
cleaner, found := registry.Get("docker")
if !found {
    return fmt.Errorf("cleaner not found")
}
```

### List

Returns all registered cleaners. Order is not guaranteed.

```go
allCleaners := registry.List()
for _, c := range allCleaners {
    fmt.Println(c.Name())
}
```

### Names

Returns all registered cleaner names.

```go
names := registry.Names()
// ["docker", "npm", "gradle", ...]
```

### Count

Returns the number of registered cleaners.

```go
count := registry.Count()
```

## Platform-Aware Operations

### Available

Returns all cleaners that are available on the current system by checking `IsAvailable(ctx)`.

```go
available := registry.Available(ctx)
for _, c := range available {
    fmt.Printf("%s is available\n", c.Name())
}
```

This is useful when you want to show users only the cleaners that can actually run on their system.

## Batch Operations

### CleanAll

Runs all available cleaners and aggregates results. Returns a map of cleaner name to result.

```go
results := registry.CleanAll(ctx)
for name, result := range results {
    if result.IsOk() {
        cleanResult := result.Unwrap()
        fmt.Printf("%s: freed %d bytes\n", name, cleanResult.BytesFreed)
    } else {
        err := result.UnwrapErr()
        fmt.Printf("%s: failed - %v\n", name, err)
    }
}
```

## Management Methods

### Unregister

Removes a cleaner from the registry.

```go
registry.Unregister("deprecated-cleaner")
```

### Clear

Removes all cleaners from the registry.

```go
registry.Clear()
```

## Cleaner Interface

All registered cleaners must implement the `Cleaner` interface:

```go
type Cleaner interface {
    // Name returns the unique identifier for this cleaner.
    Name() string

    // Clean executes the cleaning operation and returns the result.
    Clean(ctx context.Context) result.Result[domain.CleanResult]

    // IsAvailable checks if the cleaner can run on this system.
    IsAvailable(ctx context.Context) bool

    // Scan scans for cleanable items and returns them.
    Scan(ctx context.Context) result.Result[[]domain.ScanItem]
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"

    "github.com/LarsArtmann/clean-wizard/internal/cleaner"
)

func main() {
    ctx := context.Background()
    registry := cleaner.NewRegistry()

    // Register cleaners
    registry.Register("docker", NewDockerCleaner())
    registry.Register("npm", NewNpmCleaner())
    registry.Register("gradle", NewGradleCleaner())

    // Show available cleaners
    fmt.Printf("Registered %d cleaners\n", registry.Count())
    fmt.Printf("Available on this system: %d\n", len(registry.Available(ctx)))

    // Run all available cleaners
    results := registry.CleanAll(ctx)

    // Process results
    var totalFreed int64
    for name, result := range results {
        if result.IsOk() {
            cleanResult := result.Unwrap()
            totalFreed += cleanResult.BytesFreed
            fmt.Printf("✓ %s: %s freed\n", name, formatBytes(cleanResult.BytesFreed))
        } else {
            fmt.Printf("✗ %s: %v\n", name, result.UnwrapErr())
        }
    }

    fmt.Printf("\nTotal freed: %s\n", formatBytes(totalFreed))
}

func formatBytes(b int64) string {
    const unit = 1024
    if b < unit {
        return fmt.Sprintf("%d B", b)
    }
    div, exp := int64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
```

## Thread Safety

All Registry methods are thread-safe:

- **Read operations** (`Get`, `List`, `Names`, `Count`, `Available`) use `RLock`
- **Write operations** (`Register`, `Unregister`, `Clear`) use `Lock`

You can safely use the registry from multiple goroutines concurrently.
