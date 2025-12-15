# Clean Wizard Documentation Hub

## 📚 Package Documentation

| Package                         | Description                                    | Documentation             |
| ------------------------------- | ---------------------------------------------- | ------------------------- |
| [domain](./domain.md)           | Core domain types and business logic           | [GoDoc](./domain.md)      |
| [adapters](./adapters.md)       | Modern library integrations and infrastructure | [GoDoc](./adapters.md)    |
| [config](./config.md)           | Configuration management and validation        | [GoDoc](./config.md)      |
| [cleaner](./cleaner.md)         | Cleaning operations and adapters               | [GoDoc](./cleaner.md)     |
| [conversions](./conversions.md) | Type conversions and builders                  | [GoDoc](./conversions.md) |
| [middleware](./middleware.md)   | Validation and request middleware              | [GoDoc](./middleware.md)  |
| [result](./result.md)           | Result type with railway programming           | [GoDoc](./result.md)      |

## 🚀 Quick Start Guide

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/LarsArtmann/clean-wizard/internal/cleaner"
    "github.com/LarsArtmann/clean-wizard/internal/domain"
    "github.com/LarsArtmann/clean-wizard/internal/adapters"
)

func main() {
    // Load environment configuration
    cfg, err := adapters.LoadEnvironmentConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Create rate limiter
    rateLimiter := adapters.NewRateLimiter(cfg.RateLimitRPS, 10)

    // Create cache manager
    cache := adapters.NewCacheManager(cfg.CacheTTL, cfg.CacheCleanupInterval)

    // Create Nix cleaner
    nixCleaner := cleaner.NewNixCleaner(true) // Dry run

    // Get store size
    ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
    defer cancel()

    // Wait for rate limiter
    if err := rateLimiter.Wait(ctx); err != nil {
        log.Fatalf("Rate limit error: %v", err)
    }

    // Get store size with caching
    cacheKey := "store_size"
    if cachedSize, found := cache.Get(cacheKey); found {
        fmt.Printf("Cached store size: %v bytes\n", cachedSize)
    } else {
        storeSize := nixCleaner.GetStoreSize(ctx)
        cache.Set(cacheKey, storeSize, cfg.CacheTTL)
        fmt.Printf("Store size: %v bytes\n", storeSize)
    }
}
```

### Error Handling

```go
import "github.com/LarsArtmann/clean-wizard/internal/result"

func handleCleaningResult(res result.Result[domain.CleanResult]) {
    if res.IsOk() {
        cleanResult := res.Value()
        fmt.Printf("Cleaned %d items, freed %d bytes\n",
            cleanResult.ItemsRemoved, cleanResult.FreedBytes)
    } else {
        fmt.Printf("Cleaning failed: %v\n", res.Error())
    }
}
```

### Configuration

```go
// Environment variables
export DEBUG=true
export SAFE_MODE=false
export MAX_DISK_USAGE_PERCENT=80
export NIX_PATH=/custom/nix/path

// Programmatic configuration
cfg := &domain.Config{
    Version:      "1.0.0",
    SafeMode:     true,
    MaxDiskUsage: 50,
    Protected:    []string{"/System", "/Applications"},
    Profiles: map[string]*domain.Profile{
        "nix-cleanup": {
            Name:        "Nix Cleanup",
            Description: "Clean Nix generations",
            Operations: []domain.CleanupOperation{
                {
                    Name:        "nix-generations",
                    Description: "Clean old Nix generations",
                    RiskLevel:   domain.RiskLow,
                    Enabled:     true,
                    Settings: &domain.OperationSettings{
                        NixGenerations: &domain.NixGenerationsSettings{
                            Generations: 3,
                            Optimize:    true,
                        },
                    },
                },
            },
            Enabled: true,
        },
    },
}
```

## 🔧 Advanced Features

### Rate Limiting

```go
// Create rate limiter (10 RPS, burst of 5)
rateLimiter := adapters.NewRateLimiter(10, 5)

// Wait for permission
err := rateLimiter.Wait(ctx)
if err != nil {
    return fmt.Errorf("rate limit exceeded: %w", err)
}

// Check immediate availability
if rateLimiter.Allow() {
    // Proceed with operation
}
```

### Caching

```go
// Create cache (5 minute TTL, 10 minute cleanup)
cache := adapters.NewCacheManager(5*time.Minute, 10*time.Minute)

// Store value
cache.Set("key", value, time.Minute)

// Retrieve value
if value, found := cache.Get("key"); found {
    fmt.Printf("Found: %v\n", value)
}
```

### HTTP Client

```go
// Create HTTP client with custom settings
client := adapters.NewHTTPClient().
    WithTimeout(30*time.Second).
    WithRetry(3, 1*time.Second, 10*time.Second).
    WithAuth("Bearer", "token")

// Make request
resp, err := client.Get(ctx, "https://api.example.com/data")
if err != nil {
    return fmt.Errorf("HTTP request failed: %w", err)
}

if resp.IsSuccess() {
    fmt.Printf("Response: %s\n", resp.Body)
}
```

## 🏗️ Architecture

### Domain-Driven Design

- **Domain**: Core business logic and types
- **Adapters**: External integrations and infrastructure
- **Configuration**: Settings management and validation
- **Cleaners**: Domain-specific cleaning operations
- **Conversions**: Type-safe transformations and builders
- **Middleware**: Cross-cutting concerns (validation, etc.)
- **Result**: Railway programming for error handling

### Type Safety

- Compile-time enums for risk levels, strategies, etc.
- Railway programming pattern with Result types
- Strong typing throughout the codebase
- Elimination of `map[string]any` usage

## 🧪 Testing

### Unit Tests

```bash
go test ./...
go test ./internal/domain -v
go test ./internal/adapters -v
```

### Integration Tests

```bash
go test ./tests/bdd -v
go test ./internal/config -v -run TestIntegration
```

### Benchmarking

```bash
go test ./internal/cleaner -bench=.
go test ./internal/conversions -bench=Benchmark
```

## 📊 Performance

### Benchmarks

- **Nix Generation Listing**: ~10ms for 100 generations
- **Configuration Validation**: ~1ms for complex configs
- **Cache Operations**: ~100μs for in-memory cache
- **Rate Limiting**: ~10μs for token bucket operations

### Memory Usage

- **Base Application**: ~20MB RSS
- **Cache (1000 items)**: +~10MB
- **Nix Store Data**: ~5MB per 100 generations

## 🛡️ Security

### Safe Mode

All cleaning operations run in safe mode by default:

```go
// Safe mode requires explicit confirmation
cfg := &domain.Config{
    SafeMode: true,
    // ... other settings
}

// Dry run for testing
nixCleaner := cleaner.NewNixCleaner(true) // dryRun = true
```

### Risk Management

- **RiskLow**: Safe for automated execution
- **RiskMedium**: Requires user confirmation
- **RiskHigh**: Requires explicit approval
- **RiskCritical**: Blocked in safe mode

## 🔍 Monitoring & Observability

### Logging

```go
import "github.com/rs/zerolog"

log := zerolog.New(os.Stderr).With().
    Timestamp().
    Logger()

log.Info().
    Str("operation", "nix_cleanup").
    Int("generations_removed", 3).
    Int64("bytes_freed", 1073741824).
    Msg("Nix cleanup completed")
```

### Metrics

```go
// Track cleaning metrics
metrics := struct{
    OperationsTotal     int64 `json:"operations_total"`
    ItemsCleanedTotal  int64 `json:"items_cleaned_total"`
    BytesFreedTotal    int64 `json:"bytes_freed_total"`
}{
    OperationsTotal:    100,
    ItemsCleanedTotal:  1500,
    BytesFreedTotal:   1024 * 1024 * 1024 * 50, // 50GB
}
```

## 🚨 Troubleshooting

### Common Issues

#### Rate Limiting

```go
if errors.Is(err, context.DeadlineExceeded) {
    fmt.Println("Operation timed out, may be rate limited")
}
```

#### Cache Misses

```go
if _, found := cache.Get(key); !found {
    fmt.Println("Cache miss, fallback to direct operation")
}
```

#### Configuration Errors

```go
if err := cfg.ValidateEnvironmentConfig(); err != nil {
    fmt.Printf("Configuration error: %v\n", err)
}
```

### Debug Mode

```bash
export DEBUG=true
export LOG_LEVEL=debug
./clean-wizard scan
```

## 📝 Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/LarsArtmann/clean-wizard.git
cd clean-wizard

# Install dependencies
go mod download

# Run tests
go test ./...

# Run with debug
go run ./cmd/clean-wizard --debug
```

### Code Style

- Use Go fmt and go vet
- Maintain 80%+ test coverage
- Document all public APIs
- Use railway programming (Result types)
- Follow domain-driven design principles

## 📄 License

MIT License - see LICENSE file for details.
