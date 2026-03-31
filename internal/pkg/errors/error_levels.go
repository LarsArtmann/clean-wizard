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

// errorLevelData provides string representations for each error level.
var errorLevelData = [...]struct {
	stringRepr string
	logRepr    string
}{
	LevelDebug: {stringRepr: "DEBUG", logRepr: "debug"},
	LevelInfo:  {stringRepr: "INFO", logRepr: "info"},
	LevelWarn:  {stringRepr: "WARN", logRepr: "warning"},
	LevelError: {stringRepr: "ERROR", logRepr: "error"},
	LevelFatal: {stringRepr: "FATAL", logRepr: "fatal"},
	LevelPanic: {stringRepr: "PANIC", logRepr: "panic"},
}

// String returns string representation of error level.
func (e ErrorLevel) String() string {
	if e >= 0 && e < ErrorLevel(len(errorLevelData)) {
		return errorLevelData[e].stringRepr
	}

	return "UNKNOWN_LEVEL"
}

// LogLevel returns the corresponding logrus level.
func (e ErrorLevel) LogLevel() string {
	if e >= 0 && e < ErrorLevel(len(errorLevelData)) {
		return errorLevelData[e].logRepr
	}

	return "error"
}
