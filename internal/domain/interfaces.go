package domain

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// Scanner interface for domain scanners
type Scanner interface {
	Scan(ctx context.Context, req ScanRequest) result.Result[ScanResult]
	SupportedTypes() []ScanType
	Name() string
}

// Cleaner interface for domain cleaners  
type Cleaner interface {
	Clean(ctx context.Context, req CleanRequest) result.Result[CleanResult]
	SupportedStrategies() []string
	Name() string
}

// Repository interface for data persistence
type Repository[T any] interface {
	Save(ctx context.Context, entity T) result.Result[void]
	Find(ctx context.Context, id string) result.Result[T]
	FindAll(ctx context.Context) result.Result[[]T]
	Delete(ctx context.Context, id string) result.Result[void]
}

// EventBus interface for domain events
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler EventHandler) error
}

// Event represents domain event
type Event interface {
	Type() string
	Data() interface{}
	Timestamp() time.Time
}

// EventHandler handles domain events
type EventHandler func(ctx context.Context, event Event) error

// void type for Result when no value is needed
type void struct{}

// TODO: Move to generics.go or similar utility file
// TODO: Implement proper event sourcing with CQRS
// TODO: Add validation middleware
// TODO: Implement retry policies
// TODO: Add circuit breaker pattern