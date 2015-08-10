package logger

import (
	"log"
	"os"
)

var (
	enableDebug = os.Getenv("debug") != ""
	enableTrace = os.Getenv("trace") != ""

	lg = log.New(os.Stdout, "", log.LstdFlags)
)

func init() {
	if enableDebug || enableTrace {
		lg.Printf("LOGGER: logging debug logs")
	}

	if enableTrace {
		lg.Printf("LOGGER: logging trace logs")
	}
}

// Logger logs stuff
type Logger struct {
	prefix string
}

// New creates a logger with prefix
func New(prefix string) (l Logger) {
	return Logger{prefix}
}

// Log prints important information.
func (l *Logger) Log(logs ...interface{}) {
	l.print("INFO", logs)
}

// Error prints error messages. Panics if PANIC env is set.
func (l *Logger) Error(logs ...interface{}) {
	l.print("ERROR", logs)
}

// Debug prints debug messages if DEBUG env or TRACE env is set.
func (l *Logger) Debug(logs ...interface{}) {
	if enableDebug || enableTrace {
		l.print("DEBUG", logs)
	}
}

// Trace prints verbose debug messages if TRACE env is set.
func (l *Logger) Trace(logs ...interface{}) {
	if enableTrace {
		l.print("TRACE", logs)
	}
}

func (l *Logger) print(prefix string, logs []interface{}) {
	lg.Printf(l.prefix + "." + prefix + ":")

	for _, item := range logs {
		lg.Printf("%+v: ", item)
	}
}
