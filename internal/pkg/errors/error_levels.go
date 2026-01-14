package errors

// ErrorLevel represents severity level of errors.
type ErrorLevel int

const (
	// Error levels.
	LevelDebug ErrorLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

// String returns string representation of error level.
func (e ErrorLevel) String() string {
	switch e {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelPanic:
		return "PANIC"
	default:
		return "UNKNOWN_LEVEL"
	}
}

// LogLevel returns the corresponding logrus level.
func (e ErrorLevel) LogLevel() string {
	switch e {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warning"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelPanic:
		return "panic"
	default:
		return "error"
	}
}
