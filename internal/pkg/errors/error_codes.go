package errors

// ErrorCode represents standardized error codes.
type ErrorCode int

const (
	// General errors.
	ErrUnknown ErrorCode = iota
	ErrInvalidInput
	ErrNotFound
	ErrPermissionDenied
	ErrTimeout

	// Configuration errors.
	ErrConfigLoad
	ErrConfigSave
	ErrConfigValidation

	// Nix-specific errors.
	ErrNixNotAvailable
	ErrNixCommandFailed
	ErrNixStoreCorrupted

	// Cleaning errors.
	ErrCleaningFailed
	ErrCleaningTimeout
	ErrCleanupRollback
)

// errorCodeStrings maps ErrorCode values to their string representations.
var errorCodeStrings = map[ErrorCode]string{
	ErrUnknown:           "UNKNOWN",
	ErrInvalidInput:      "INVALID_INPUT",
	ErrNotFound:          "NOT_FOUND",
	ErrPermissionDenied:  "PERMISSION_DENIED",
	ErrTimeout:           "TIMEOUT",
	ErrConfigLoad:        "CONFIG_LOAD",
	ErrConfigSave:        "CONFIG_SAVE",
	ErrConfigValidation:  "CONFIG_VALIDATION",
	ErrNixNotAvailable:   "NIX_NOT_AVAILABLE",
	ErrNixCommandFailed:  "NIX_COMMAND_FAILED",
	ErrNixStoreCorrupted: "NIX_STORE_CORRUPTED",
	ErrCleaningFailed:    "CLEANING_FAILED",
	ErrCleaningTimeout:   "CLEANING_TIMEOUT",
	ErrCleanupRollback:   "CLEANUP_ROLLBACK",
}

// String returns string representation of error code.
func (e ErrorCode) String() string {
	if str, ok := errorCodeStrings[e]; ok {
		return str
	}
	return "UNKNOWN_ERROR_CODE"
}
