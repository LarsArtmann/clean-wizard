// Package execution orchestrates cleaner execution via Azure/go-workflow.
// It replaces the former sequential for-loop in the command layer with a
// proper DAG-based workflow engine that supports parallel execution,
// step hooks, and structured error collection.
//
// The execution layer is deliberately DI-agnostic: it receives a
// *cleaner.Registry and selected cleaner names as plain parameters,
// matching BuildFlow's execution package design where the workflow
// engine never imports the DI container.
package execution
