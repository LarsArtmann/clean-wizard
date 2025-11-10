package config

import (
	"time"
)

// ConfigChangeResult represents configuration change validation result
type ConfigChangeResult struct {
	IsValid bool           `json:"is_valid"`
	Changes []ConfigChange  `json:"changes"`
	Errors  []string       `json:"errors"`
	Warnings []string      `json:"warnings"`
	Risk    string         `json:"risk"`
	Timestamp time.Time    `json:"timestamp"`
}

// ConfigChange represents a configuration change
type ConfigChange struct {
	Field     string      `json:"field"`
	Operation string      `json:"operation"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Risk      string      `json:"risk"`
}

// ProfileOperationResult represents profile operation validation result
type ProfileOperationResult struct {
	IsValid    bool   `json:"is_valid"`
	Operation  string `json:"operation"`
	Profile    string `json:"profile"`
	Risk       string `json:"risk"`
	Suggestions []string `json:"suggestions"`
	Errors     []string `json:"errors"`
	Timestamp  time.Time `json:"timestamp"`
}

// AddError adds an error to the result
func (ccr *ConfigChangeResult) AddError(err string) {
	ccr.Errors = append(ccr.Errors, err)
	ccr.IsValid = false
}

// AddWarning adds a warning to the result
func (ccr *ConfigChangeResult) AddWarning(warn string) {
	ccr.Warnings = append(ccr.Warnings, warn)
}

// HasErrors returns true if result has errors
func (ccr *ConfigChangeResult) HasErrors() bool {
	return len(ccr.Errors) > 0
}

// HasWarnings returns true if result has warnings
func (ccr *ConfigChangeResult) HasWarnings() bool {
	return len(ccr.Warnings) > 0
}

// ChangeCount returns number of changes
func (ccr *ConfigChangeResult) ChangeCount() int {
	return len(ccr.Changes)
}

// AddError adds an error to the profile operation result
func (por *ProfileOperationResult) AddError(err string) {
	por.Errors = append(por.Errors, err)
	por.IsValid = false
}

// AddSuggestion adds a suggestion to the profile operation result
func (por *ProfileOperationResult) AddSuggestion(suggestion string) {
	por.Suggestions = append(por.Suggestions, suggestion)
}

// HasErrors returns true if profile operation result has errors
func (por *ProfileOperationResult) HasErrors() bool {
	return len(por.Errors) > 0
}