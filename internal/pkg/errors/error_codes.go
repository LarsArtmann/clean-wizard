package errors

// ErrorCode represents standardized error codes
type ErrorCode int

const (
	// General errors
	ErrUnknown ErrorCode = iota
	ErrInvalidInput
	ErrNotFound
	ErrPermissionDenied
	ErrTimeout

	// Configuration errors
	ErrConfigLoad
	ErrConfigSave
	ErrConfigValidation

	// Nix-specific errors
	ErrNixNotAvailable
	ErrNixCommandFailed
	ErrNixStoreCorrupted

	// Cleaning errors
	ErrCleaningFailed
	ErrCleaningTimeout
	ErrCleanupRollback
)

// String returns string representation of error code
func (e ErrorCode) String() string {
	switch e {
	case ErrUnknown:
		return "UNKNOWN"
	case ErrInvalidInput:
		return "INVALID_INPUT"
	case ErrNotFound:
		return "NOT_FOUND"
	case ErrPermissionDenied:
		return "PERMISSION_DENIED"
	case ErrTimeout:
		return "TIMEOUT"
	case ErrConfigLoad:
		return "CONFIG_LOAD"
	case ErrConfigSave:
		return "CONFIG_SAVE"
	case ErrConfigValidation:
		return "CONFIG_VALIDATION"
	case ErrNixNotAvailable:
		return "NIX_NOT_AVAILABLE"
	case ErrNixCommandFailed:
		return "NIX_COMMAND_FAILED"
	case ErrNixStoreCorrupted:
		return "NIX_STORE_CORRUPTED"
	case ErrCleaningFailed:
		return "CLEANING_FAILED"
	case ErrCleaningTimeout:
		return "CLEANING_TIMEOUT"
	case ErrCleanupRollback:
		return "CLEANUP_ROLLBACK"
	default:
		return "UNKNOWN_ERROR_CODE"
	}
}
