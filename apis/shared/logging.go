package shared

// VerbosityLevel is logging verbosity levels
type VerbosityLevel string

const (
	// NoLogs outputs no logs
	NoLogs VerbosityLevel = "off"
	// FatalLogs outputs only fatal logs
	FatalLogs VerbosityLevel = "fatal"
	// ErrorLogs outputs only error logs
	ErrorLogs VerbosityLevel = "error"
	// WarnLogs outputs only warning logs
	WarnLogs VerbosityLevel = "warn"
	// InfoLogs outputs only informational logs
	InfoLogs VerbosityLevel = "info"
	// DebugLogs outputs only debugging logs
	DebugLogs VerbosityLevel = "debug"
	// TraceLogs outputs only tracing logs
	TraceLogs VerbosityLevel = "trace"
	// AllLogs outputs only all logs
	AllLogs VerbosityLevel = "all"
	// NoticeLogs outputs only notice logs
	NoticeLogs VerbosityLevel = "notice"
	// CriticalLogs outputs only critical logs
	CriticalLogs VerbosityLevel = "crit"
	// PanicLogs outputs only panic logs
	PanicLogs VerbosityLevel = "panic"
	// NoneLogs outputs no logs
	NoneLogs VerbosityLevel = "none"
)
