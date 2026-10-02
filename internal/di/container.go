// Package di provides the dependency injection container for clean-wizard,
// built on samber/do v2. It centralizes all service registration and lifecycle
// management, replacing the former hand-wired dual-registry approach.
//
// The container is created once at application startup, services are registered
// via RegisterAllServices, and the returned cleanup function ensures graceful
// shutdown of all services implementing do.ShutdownerWithError.
package di

import (
	"log/slog"

	"github.com/LarsArtmann/clean-wizard/internal/logger"
	"github.com/samber/do/v2"
)

// Container wraps a samber/do injector with lifecycle management.
// It is the single entry point for resolving any service in the application.
type Container struct {
	injector do.Injector
}

// New creates a fresh Container with an empty injector.
// The returned cleanup function must be deferred to ensure all services
// implementing do.ShutdownerWithError are gracefully stopped. Shutdown
// failures are logged at Debug (they must never mask the command's real
// outcome, but should be visible when diagnosing leaked resources).
func New() (*Container, func()) {
	injector := do.New()

	return &Container{injector: injector}, func() {
		report := injector.Shutdown()
		if report == nil || len(report.Errors) == 0 {
			return
		}

		slogger := logger.StdLogger
		if slogger == nil {
			slogger = slog.Default()
		}

		slogger.Debug("DI shutdown reported service errors", "report", report.Error())
	}
}

// Injector exposes the underlying samber/do injector for direct use
// with do.Provide, do.Invoke, etc.
func (c *Container) Injector() do.Injector {
	return c.injector
}

// Shutdown gracefully shuts down all registered services.
func (c *Container) Shutdown() *do.ShutdownReport {
	return c.injector.Shutdown()
}
