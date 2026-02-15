# Project Split Executive Report: Clean Wizard

## Introduction
This report outlines a strategic proposal to refactor the existing `clean-wizard` project into multiple, highly focused, and independently deployable projects. The current `clean-wizard` project, a comprehensive cleaning system for various development and system caches, has grown in complexity, encompassing core cleaning logic, configuration management, a command-line interface, and extensive internal utilities. This monolithic structure, while functional, presents challenges in terms of maintainability, reusability, testing, and scalability.

The proposed split aims to enhance modularity by establishing clear boundaries between distinct functionalities, allowing for independent development, deployment, and versioning of each component.

## Current Project Overview
The `clean-wizard` project is a robust tool designed to identify and remove redundant files and caches from different environments (e.g., Go, Docker, Nix, Cargo, Node.js, Homebrew, system temporary files). Its key features include:
*   **Multi-Cleaner Support:** A wide array of specialized cleaners for various technologies.
*   **Advanced Configuration:** A sophisticated configuration system with validation and sanitization.
*   **Command-Line Interface:** A user-friendly CLI for initiating and managing cleaning operations.
*   **Domain Modeling:** A well-defined internal domain model for cleaning operations, results, and settings.
*   **Extensive Testing:** Comprehensive unit, integration, and BDD test suites.
*   **Detailed Documentation:** Rich internal and external documentation, architectural analyses, and planning documents.

## Rationale for Splitting
The primary motivations for splitting the `clean-wizard` project are:
1.  **Decoupling Concerns:** Separating distinct functionalities into their own projects ensures each component adheres to the Single Responsibility Principle.
2.  **Improved Maintainability:** Smaller codebases are easier to understand, debug, and maintain. Changes in one area are less likely to introduce regressions in unrelated parts.
3.  **Enhanced Reusability:** Core components, such as cleaning logic and configuration parsers, can be reused in other projects or integrated into different interfaces (e.g., a GUI, a web service) without carrying the weight of the entire CLI application.
4.  **Independent Development & Deployment:** Teams can work on specific sub-projects without affecting others, allowing for faster iterations and independent release cycles.
5.  **Simplified Testing:** Each project can be tested in isolation, leading to more focused and efficient test suites.
6.  **Scalability:** The modular architecture better supports future growth and the addition of new features or cleaners.

## Proposed Project Structure

The `clean-wizard` project can be logically decomposed into five highly focused, independent projects:

### 1. `clean-wizard-core-cleaners` (Library/Engine)
*   **Purpose:** This project will encapsulate the fundamental cleaning logic for all supported environments. It will define the `Cleaner` interface and provide concrete implementations for each specific cleaner (e.g., Go cache, Docker images, Nix store paths). It will be a pure Go library focused solely on *what* to clean and *how*.
*   **Key Components (from original project):**
    *   `internal/cleaner/` (all cleaner implementations, interfaces, registries)
    *   Relevant parts of `internal/domain/` (e.g., `interfaces.go`, cleaner-specific types)
    *   Potentially core execution utilities if tightly coupled to the cleaning process (e.g., `internal/adapters/exec.go`, `internal/adapters/nix.go`)
*   **Dependencies:** Minimal external dependencies, potentially just core Go libraries and specific tooling for each cleaner (e.g., Docker client library).
*   **Output:** A Go module providing a programmatic interface to various cleaning capabilities.

### 2. `clean-wizard-config` (Library/Schema)
*   **Purpose:** This project will provide a robust, type-safe, and validated configuration management system. It will handle loading configuration files (e.g., YAML, JSON), validating them against a schema, and applying sanitization rules. This library can be used by any application needing to configure cleaning operations.
*   **Key Components:**
    *   `internal/config/` (all configuration loading, parsing, validation, sanitization logic)
    *   Relevant parts of `internal/domain/` (e.g., `config.go`, `config_methods.go`, `operation_settings.go`)
    *   `schemas/config.schema.json` (the formal configuration schema)
    *   `internal/middleware/validation.go`
*   **Dependencies:** Configuration parsing libraries (e.g., YAML unmarshaller), validation frameworks.
*   **Output:** A Go module offering configuration loading, validation, and sanitization functionalities.

### 3. `clean-wizard-cli` (Application)
*   **Purpose:** This will be the standalone command-line interface application. Its sole responsibility will be to provide the user-facing interaction, orchestrating operations by consuming the `clean-wizard-core-cleaners` and `clean-wizard-config` libraries. It will handle parsing CLI arguments, loading configurations, invoking cleaners, and presenting results in a user-friendly format.
*   **Key Components:**
    *   `cmd/clean-wizard/main.go`
    *   `internal/interface/cli/`
    *   `internal/api/` (for mapping CLI commands to core logic)
    *   `internal/format/` (for consistent output presentation)
*   **Dependencies:** `clean-wizard-core-cleaners`, `clean-wizard-config`, CLI parsing libraries (e.g., Cobra, spf13/cobra).
*   **Output:** A single executable binary.

### 4. `clean-wizard-reporting` (Library/Service)
*   **Purpose:** This project will focus on the aggregation, analysis, and presentation of cleaning results. It could generate detailed reports, integrate with monitoring dashboards, or provide data for analytics. This project would consume results from cleaning operations and transform them into actionable insights.
*   **Key Components:**
    *   `internal/result/` (types and logic for handling cleaning results)
    *   `internal/format/` (specifically for result formatting and potentially different output types like JSON, Markdown, etc.)
*   **Dependencies:** `clean-wizard-core-cleaners` (for result types).
*   **Output:** A Go module for result processing, or potentially a microservice if advanced reporting capabilities (e.g., web interface, database storage) are required.

### 5. `clean-wizard-schema-definitions` (Shared/Utility)
*   **Purpose:** This is a very lightweight, language-agnostic project that will house all formal schema definitions (e.g., JSON Schema, OpenAPI, TypeSpec) for core data structures, configurations, and API contracts. Its primary goal is to ensure consistency across all `clean-wizard` projects and enable code generation for different languages if needed.
*   **Key Components:**
    *   `schemas/config.schema.json`
    *   `api/typespec/clean-wizard.tsp`
*   **Dependencies:** None.
*   **Output:** A collection of schema files, serving as the single source of truth for data structures.

## Example New Directory Structure

```
/Users/larsartmann/projects/
├── clean-wizard-core-cleaners/
│   ├── go.mod
│   ├── internal/cleaner/
│   ├── internal/domain/
│   └── ... (relevant files for core cleaning)
├── clean-wizard-config/
│   ├── go.mod
│   ├── internal/config/
│   ├── internal/domain/
│   ├── schemas/config.schema.json
│   └── ... (relevant files for configuration)
├── clean-wizard-cli/
│   ├── go.mod
│   ├── cmd/clean-wizard/main.go
│   ├── internal/interface/cli/
│   ├── internal/api/
│   ├── internal/format/
│   └── ... (relevant files for CLI)
├── clean-wizard-reporting/
│   ├── go.mod
│   ├── internal/result/
│   ├── internal/format/
│   └── ... (relevant files for reporting)
└── clean-wizard-schema-definitions/
    ├── schemas/config.schema.json
    ├── api/typespec/clean-wizard.tsp
    └── README.md
```

## Conclusion
Splitting the `clean-wizard` project into these focused modules will significantly improve its architectural health, making it more manageable, testable, extensible, and reusable. This decomposition aligns with modern software engineering best practices for building scalable and maintainable systems. The proposed structure provides a clear roadmap for refactoring and continued development of the `clean-wizard` ecosystem.
