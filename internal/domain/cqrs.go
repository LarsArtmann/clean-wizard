package domain

import (
	"context"
	"time"
)

// TODO: Implement proper event sourcing with immutable events
// TODO: Add snapshot support for performance optimization
// TODO: Implement proper aggregate boundaries

// Event represents domain event
type Event interface {
	ID() string
	AggregateID() string
	Type() string
	Data() interface{}
	Version() int
	Timestamp() time.Time
	OccuredAt() time.Time
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	EventId      string    `json:"id"`
	AggregateId  string    `json:"aggregate_id"`
	EventType    string    `json:"type"`
	EventData    interface{} `json:"data"`
	EventVersion int       `json:"version"`
	EventTime    time.Time `json:"timestamp"`
	OccurredAt    time.Time `json:"occured_at"`
}

func (e BaseEvent) ID() string                     { return e.EventId }
func (e BaseEvent) AggregateID() string           { return e.AggregateId }
func (e BaseEvent) Type() string                 { return e.EventType }
func (e BaseEvent) Data() interface{}            { return e.EventData }
func (e BaseEvent) Version() int                 { return e.EventVersion }
func (e BaseEvent) Timestamp() time.Time          { return e.EventTime }
func (e BaseEvent) OccuredAt() time.Time          { return e.OccurredAt }

// EventStore interface for event persistence
type EventStore interface {
	Save(ctx context.Context, events []Event) error
	Load(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error)
	GetEvents(ctx context.Context, aggregateID string) ([]Event, error)
	GetEventsByType(ctx context.Context, eventType string, limit int) ([]Event, error)
}

// Command represents intent to change system state
type Command interface {
	ID() string
	AggregateID() string
	Type() string
	Data() interface{}
	Timestamp() time.Time
}

// BaseCommand provides common command functionality
type BaseCommand struct {
	CommandId      string    `json:"id"`
	CommandAggregateId string    `json:"aggregate_id"`
	CommandType    string    `json:"type"`
	CommandData    interface{} `json:"data"`
	CommandTime   time.Time `json:"timestamp"`
}

func (c BaseCommand) ID() string                    { return c.CommandId }
func (c BaseCommand) AggregateID() string            { return c.CommandAggregateId }
func (c BaseCommand) Type() string                  { return c.CommandType }
func (c BaseCommand) Data() interface{}             { return c.CommandData }
func (c BaseCommand) Timestamp() time.Time          { return c.CommandTime }

// Query represents read operation
type Query interface {
	ID() string
	Type() string
	Data() interface{}
	Timestamp() time.Time
}

// BaseQuery provides common query functionality
type BaseQuery struct {
	QueryId      string    `json:"id"`
	QueryType    string    `json:"type"`
	QueryData    interface{} `json:"data"`
	QueryTime    time.Time `json:"timestamp"`
}

func (q BaseQuery) ID() string                    { return q.QueryId }
func (q BaseQuery) Type() string                  { return q.QueryType }
func (q BaseQuery) Data() interface{}             { return q.QueryData }
func (q BaseQuery) Timestamp() time.Time          { return q.QueryTime }

// QueryHandler handles read queries
type QueryHandler[T any] interface {
	Handle(ctx context.Context, query Query) (T, error)
}

// CommandHandler handles write commands
type CommandHandler interface {
	Handle(ctx context.Context, command Command) error
}

// EventHandler handles domain events
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	CanHandle(eventType string) bool
}

// TODO: Implement proper command bus with middleware
// TODO: Add query bus with caching capabilities
// TODO: Implement saga pattern for distributed transactions
// TODO: Add proper retry policies and circuit breakers