// Package errorfamily provides structured error classification and handling for Go.
// Every error gets a behavioral Family (retry/no-retry), a machine-readable code,
// human-readable context, and optional diagnostics. Designed for CLI and HTTP boundaries.
//
// # The contract is the interfaces, not the struct
//
// The public protocol is four small interfaces in interfaces.go: [Coded],
// [Classified], [Contextual], and [Retryable]. Each error type implements the
// combination that makes sense for it; consumers ask for exactly what they need
// via errors.AsType[T].
//
// The concrete [Error] type is a reference implementation — convenient for simple
// needs, but not required. Projects with domain-specific error shapes (a finding
// with File/Line, an HTTP error with Status) build their own struct and implement
// only the interfaces they need. Share the protocol, not the implementation.
//
// # Stability
//
// The root package (classification core) is the stable foundation.
// The agent, diagnose, and bridge packages are experimental (v0.x).
package errorfamily
